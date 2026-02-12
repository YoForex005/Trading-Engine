package notifications

import (
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/google/uuid"
)

// NotificationManager handles routing and delivery of notifications
type NotificationManager struct {
	store          *Store
	wsPublisher    WebSocketPublisher // Interface for WebSocket delivery
	emailService   EmailService       // Interface for email delivery (future)
	prefsManager   *PreferencesManager // Preferences manager for user notification preferences

	// Rate limiting: max 20 notifications per minute per user
	rateLimiter    map[string]*userRateLimit
	rateLimiterMu  sync.RWMutex

	// Deduplication: track recent notifications to prevent duplicates
	recentNotifs   map[string]time.Time // key: userID:type, value: last sent time
	recentMu       sync.RWMutex

	mu sync.RWMutex
}

// userRateLimit tracks notification rate for a user
type userRateLimit struct {
	count     int
	resetTime time.Time
	mu        sync.Mutex
}

// WebSocketPublisher interface for pushing notifications via WebSocket
type WebSocketPublisher interface {
	PublishNotification(userID string, notification *Notification) error
}

// EmailService interface for sending email notifications (future implementation)
type EmailService interface {
	SendEmail(to, subject, body string) error
}

// NewNotificationManager creates a new notification manager
func NewNotificationManager(store *Store) *NotificationManager {
	nm := &NotificationManager{
		store:        store,
		rateLimiter:  make(map[string]*userRateLimit),
		recentNotifs: make(map[string]time.Time),
	}
	
	// Start cleanup goroutine for rate limiter and deduplication cache
	go nm.cleanupWorker()
	
	return nm
}

// SetWebSocketPublisher sets the WebSocket publisher for real-time delivery
func (nm *NotificationManager) SetWebSocketPublisher(publisher WebSocketPublisher) {
	nm.mu.Lock()
	defer nm.mu.Unlock()
	nm.wsPublisher = publisher
}

// SetEmailService sets the email service for email delivery
func (nm *NotificationManager) SetEmailService(service EmailService) {
	nm.mu.Lock()
	defer nm.mu.Unlock()
	nm.emailService = service
}

// Send creates and routes a notification to appropriate channels based on severity
func (nm *NotificationManager) Send(userID, notifType string, data map[string]interface{}) error {
	// Check rate limit
	if !nm.checkRateLimit(userID) {
		log.Printf("[Notifications] Rate limit exceeded for user %s", userID)
		return fmt.Errorf("rate limit exceeded")
	}
	
	// Check deduplication (don't send same type within 5 minutes)
	if nm.isDuplicate(userID, notifType) {
		log.Printf("[Notifications] Duplicate notification %s for user %s (within 5min window)", notifType, userID)
		return nil // Silently skip duplicates
	}
	
	// Get template
	template, exists := Templates[notifType]
	if !exists {
		return fmt.Errorf("unknown notification type: %s", notifType)
	}
	
	// Create notification
	notification := &Notification{
		ID:        uuid.New().String(),
		UserID:    userID,
		Type:      notifType,
		Severity:  template.Severity,
		Category:  template.Category,
		Title:     template.Title,
		Message:   interpolateMessage(template.Message, data),
		Data:      data,
		Read:      false,
		CreatedAt: time.Now().Unix(),
	}
	
	// Add action items based on type
	notification.ActionItems = getActionItems(notifType, data)
	
	// Store notification
	nm.store.Add(notification)
	
	// Route to channels based on severity
	channels := nm.getChannelsForSeverity(template.Severity)
	
	for _, channel := range channels {
		if err := nm.deliverToChannel(notification, channel); err != nil {
			log.Printf("[Notifications] Failed to deliver to %s: %v", channel, err)
		}
	}
	
	// Mark as sent for deduplication
	nm.markSent(userID, notifType)
	
	return nil
}

// getChannelsForSeverity determines which channels to use based on severity
func (nm *NotificationManager) getChannelsForSeverity(severity Severity) []NotificationChannel {
	switch severity {
	case SeverityCritical:
		// Critical: WebSocket + Email + Sound (handled by frontend)
		return []NotificationChannel{ChannelWebSocket, ChannelEmail}
	case SeverityError:
		// Error: WebSocket + Email
		return []NotificationChannel{ChannelWebSocket, ChannelEmail}
	case SeverityWarning:
		// Warning: WebSocket only
		return []NotificationChannel{ChannelWebSocket}
	case SeverityInfo:
		// Info: WebSocket only
		return []NotificationChannel{ChannelWebSocket}
	default:
		return []NotificationChannel{ChannelWebSocket}
	}
}

// deliverToChannel sends notification to a specific channel
func (nm *NotificationManager) deliverToChannel(notification *Notification, channel NotificationChannel) error {
	nm.mu.RLock()
	defer nm.mu.RUnlock()
	
	switch channel {
	case ChannelWebSocket:
		if nm.wsPublisher != nil {
			return nm.wsPublisher.PublishNotification(notification.UserID, notification)
		}
		log.Printf("[Notifications] WebSocket publisher not configured")
		return nil
		
	case ChannelEmail:
		if nm.emailService != nil {
			// Send email in goroutine (non-blocking)
			go func() {
				// Get email address from notification data
				emailAddr, ok := notification.Data["email"].(string)
				if !ok || emailAddr == "" {
					// Fallback: use userID as email (for testing)
					emailAddr = notification.UserID
				}

				// Try to render email template if template name is provided
				var htmlBody string
				templateName, hasTemplate := notification.Data["email_template"].(string)
				if hasTemplate {
					rendered, err := RenderTemplate(templateName, notification.Data)
					if err != nil {
						log.Printf("[Notifications] Failed to render email template %s: %v", templateName, err)
						htmlBody = notification.Message
					} else {
						htmlBody = rendered
					}
				} else {
					// No template, use message as plain text
					htmlBody = notification.Message
				}

				// Get or generate subject
				subject := fmt.Sprintf("[%s] %s", notification.Severity, notification.Title)
				if customSubject, ok := notification.Data["email_subject"].(string); ok && customSubject != "" {
					subject = customSubject
				}

				// Send email
				if err := nm.emailService.SendEmail(emailAddr, subject, htmlBody); err != nil {
					log.Printf("[Notifications] Failed to send email to %s: %v", emailAddr, err)
				}
			}()
			return nil
		}
		log.Printf("[Notifications] Email service not configured")
		return nil
		
	default:
		log.Printf("[Notifications] Channel %s not implemented yet", channel)
		return nil
	}
}

// checkRateLimit enforces max 20 notifications per minute per user
func (nm *NotificationManager) checkRateLimit(userID string) bool {
	nm.rateLimiterMu.Lock()
	defer nm.rateLimiterMu.Unlock()
	
	rl, exists := nm.rateLimiter[userID]
	if !exists {
		rl = &userRateLimit{
			count:     0,
			resetTime: time.Now().Add(1 * time.Minute),
		}
		nm.rateLimiter[userID] = rl
	}
	
	rl.mu.Lock()
	defer rl.mu.Unlock()
	
	// Reset counter if time window expired
	if time.Now().After(rl.resetTime) {
		rl.count = 0
		rl.resetTime = time.Now().Add(1 * time.Minute)
	}
	
	// Check limit (max 20 per minute)
	if rl.count >= 20 {
		return false
	}
	
	rl.count++
	return true
}

// isDuplicate checks if same notification type was sent within 5 minutes
func (nm *NotificationManager) isDuplicate(userID, notifType string) bool {
	nm.recentMu.RLock()
	defer nm.recentMu.RUnlock()
	
	key := fmt.Sprintf("%s:%s", userID, notifType)
	lastSent, exists := nm.recentNotifs[key]
	
	if !exists {
		return false
	}
	
	// Check if within 5-minute window
	return time.Since(lastSent) < 5*time.Minute
}

// markSent records that notification was sent for deduplication
func (nm *NotificationManager) markSent(userID, notifType string) {
	nm.recentMu.Lock()
	defer nm.recentMu.Unlock()
	
	key := fmt.Sprintf("%s:%s", userID, notifType)
	nm.recentNotifs[key] = time.Now()
}

// cleanupWorker periodically cleans up expired rate limits and deduplication cache
func (nm *NotificationManager) cleanupWorker() {
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()
	
	for range ticker.C {
		nm.cleanup()
	}
}

// cleanup removes expired entries from rate limiter and deduplication cache
func (nm *NotificationManager) cleanup() {
	now := time.Now()
	
	// Cleanup rate limiter
	nm.rateLimiterMu.Lock()
	for userID, rl := range nm.rateLimiter {
		rl.mu.Lock()
		if now.After(rl.resetTime) {
			delete(nm.rateLimiter, userID)
		}
		rl.mu.Unlock()
	}
	nm.rateLimiterMu.Unlock()
	
	// Cleanup deduplication cache (remove entries older than 10 minutes)
	nm.recentMu.Lock()
	for key, lastSent := range nm.recentNotifs {
		if now.Sub(lastSent) > 10*time.Minute {
			delete(nm.recentNotifs, key)
		}
	}
	nm.recentMu.Unlock()
}

// interpolateMessage replaces placeholders like {symbol} with actual values
func interpolateMessage(template string, data map[string]interface{}) string {
	message := template
	for key, value := range data {
		placeholder := fmt.Sprintf("{%s}", key)
		message = replaceAll(message, placeholder, fmt.Sprintf("%v", value))
	}
	return message
}

// replaceAll is a simple string replacement helper
func replaceAll(s, old, new string) string {
	result := ""
	for {
		idx := indexOf(s, old)
		if idx == -1 {
			result += s
			break
		}
		result += s[:idx] + new
		s = s[idx+len(old):]
	}
	return result
}

// indexOf finds the index of a substring
func indexOf(s, substr string) int {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return i
		}
	}
	return -1
}

// getActionItems returns action items based on notification type
func getActionItems(notifType string, data map[string]interface{}) []ActionItem {
	switch notifType {
	case "order_filled":
		orderID, _ := data["orderId"].(string)
		return []ActionItem{
			{Label: "View Order", URL: fmt.Sprintf("/orders/%s", orderID), Type: "primary"},
		}
	case "margin_warning", "margin_call":
		return []ActionItem{
			{Label: "Add Funds", URL: "/account/deposit", Type: "danger"},
			{Label: "View Positions", URL: "/positions", Type: "secondary"},
		}
	case "login_alert":
		return []ActionItem{
			{Label: "Secure Account", URL: "/security", Type: "danger"},
		}
	case "position_closed", "stop_loss_triggered", "take_profit_triggered":
		positionID, _ := data["positionId"].(string)
		return []ActionItem{
			{Label: "View History", URL: fmt.Sprintf("/history?position=%s", positionID), Type: "primary"},
		}
	default:
		return nil
	}
}
