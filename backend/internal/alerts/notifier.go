package alerts

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"sync"
	"time"

	"github.com/google/uuid"
)

// EmailConfig holds SMTP configuration
type EmailConfig struct {
	SMTPHost     string
	SMTPPort     int
	SMTPUsername string
	SMTPPassword string
	FromAddress  string
}

// SMSConfig holds Twilio configuration
type SMSConfig struct {
	AccountSID string
	AuthToken  string
	FromNumber string
}

// Notifier handles multi-channel alert notifications
type Notifier struct {
	// Configuration
	emailConfig EmailConfig
	smsConfig   SMSConfig

	// WebSocket hub for in-dashboard notifications
	wsHub WSHub

	// Notification queue
	queue      chan *Notification
	queueMu    sync.Mutex
	processing map[string]*Notification
	processMu  sync.RWMutex

	// Control
	stopChan chan struct{}
	wg       sync.WaitGroup
}

// WSHub interface for WebSocket broadcasting
type WSHub interface {
	BroadcastAlert(alert *Alert)
}

// NewNotifier creates a notification dispatcher
func NewNotifier(wsHub WSHub) *Notifier {
	return &Notifier{
		queue:      make(chan *Notification, 1000),
		processing: make(map[string]*Notification),
		stopChan:   make(chan struct{}),
		wsHub:      wsHub,
		emailConfig: EmailConfig{
			SMTPHost:     os.Getenv("SMTP_HOST"),
			SMTPPort:     getEnvAsInt("SMTP_PORT", 587),
			SMTPUsername: os.Getenv("SMTP_USERNAME"),
			SMTPPassword: os.Getenv("SMTP_PASSWORD"),
			FromAddress:  os.Getenv("SMTP_FROM_ADDRESS"),
		},
		smsConfig: SMSConfig{
			AccountSID: os.Getenv("TWILIO_ACCOUNT_SID"),
			AuthToken:  os.Getenv("TWILIO_AUTH_TOKEN"),
			FromNumber: os.Getenv("TWILIO_FROM_NUMBER"),
		},
	}
}

// Start begins processing notification queue
func (n *Notifier) Start() {
	log.Println("[Notifier] Starting notification workers")

	// Start 3 worker goroutines for parallel processing
	for i := 0; i < 3; i++ {
		n.wg.Add(1)
		go n.worker(i)
	}
}

// Stop halts notification processing
func (n *Notifier) Stop() {
	close(n.stopChan)
	n.wg.Wait()
	log.Println("[Notifier] Stopped")
}

// worker processes notifications from queue
func (n *Notifier) worker(id int) {
	defer n.wg.Done()

	log.Printf("[Notifier] Worker %d started", id)

	for {
		select {
		case notification := <-n.queue:
			n.processNotification(notification)
		case <-n.stopChan:
			log.Printf("[Notifier] Worker %d stopped", id)
			return
		}
	}
}

// Dispatch creates and queues a notification
func (n *Notifier) Dispatch(alert *Alert, channel NotificationChannel) {
	notification := &Notification{
		ID:        uuid.New().String(),
		AlertID:   alert.ID,
		AccountID: alert.AccountID,
		Channel:   channel,
		Subject:   n.formatSubject(alert),
		Body:      n.formatBody(alert),
		CreatedAt: time.Now(),
	}

	// Set recipient based on channel
	switch channel {
	case ChannelEmail:
		// TODO: Fetch user email from account service
		notification.To = "user@example.com"
	case ChannelSMS:
		// TODO: Fetch user phone from account service
		notification.To = "+1234567890"
	case ChannelDashboard:
		notification.To = alert.AccountID
	}

	// Queue for processing
	select {
	case n.queue <- notification:
		log.Printf("[Notifier] Queued %s notification for alert %s", channel, alert.ID)
	default:
		log.Printf("[Notifier] WARN: Queue full, dropping notification for alert %s", alert.ID)
	}
}

// processNotification handles delivery to specific channel
func (n *Notifier) processNotification(notification *Notification) {
	// Track in processing map
	n.processMu.Lock()
	n.processing[notification.ID] = notification
	n.processMu.Unlock()

	defer func() {
		n.processMu.Lock()
		delete(n.processing, notification.ID)
		n.processMu.Unlock()
	}()

	var err error

	switch notification.Channel {
	case ChannelDashboard:
		err = n.sendDashboard(notification)
	case ChannelEmail:
		err = n.sendEmail(notification)
	case ChannelSMS:
		err = n.sendSMS(notification)
	case ChannelWebhook:
		err = n.sendWebhook(notification)
	default:
		err = fmt.Errorf("unknown channel: %s", notification.Channel)
	}

	if err != nil {
		log.Printf("[Notifier] Failed to send %s notification %s: %v",
			notification.Channel, notification.ID, err)
		notification.Error = err.Error()
		notification.Retries++

		// Retry logic (max 3 retries)
		if notification.Retries < 3 {
			time.Sleep(time.Duration(notification.Retries) * 10 * time.Second) // Exponential backoff
			n.queue <- notification
		}
	} else {
		now := time.Now()
		notification.SentAt = &now
		log.Printf("[Notifier] Sent %s notification %s successfully", notification.Channel, notification.ID)
	}
}

// sendDashboard broadcasts alert via WebSocket
func (n *Notifier) sendDashboard(notification *Notification) error {
	if n.wsHub == nil {
		return fmt.Errorf("WebSocket hub not configured")
	}

	// Retrieve original alert (in real implementation, fetch from store)
	// For now, create a minimal alert object
	alert := &Alert{
		ID:        notification.AlertID,
		AccountID: notification.AccountID,
		Title:     notification.Subject,
		Message:   notification.Body,
		CreatedAt: notification.CreatedAt,
	}

	n.wsHub.BroadcastAlert(alert)
	return nil
}

// sendEmail sends email notification via SMTP
func (n *Notifier) sendEmail(notification *Notification) error {
	// Check if SMTP is configured
	if n.emailConfig.SMTPHost == "" {
		log.Printf("[Notifier] SMTP not configured, skipping email notification")
		return nil // Don't fail if email not configured
	}

	// TODO: Implement actual SMTP sending
	// For MVP, just log the email
	log.Printf("[Notifier] [EMAIL] To: %s, Subject: %s, Body: %s",
		notification.To, notification.Subject, notification.Body)

	// In production, use net/smtp or a library like gomail:
	/*
	import "gopkg.in/mail.v2"

	m := mail.NewMessage()
	m.SetHeader("From", n.emailConfig.FromAddress)
	m.SetHeader("To", notification.To)
	m.SetHeader("Subject", notification.Subject)
	m.SetBody("text/html", notification.Body)

	d := mail.NewDialer(n.emailConfig.SMTPHost, n.emailConfig.SMTPPort,
		n.emailConfig.SMTPUsername, n.emailConfig.SMTPPassword)

	if err := d.DialAndSend(m); err != nil {
		return err
	}
	*/

	return nil
}

// sendSMS sends SMS notification via Twilio
func (n *Notifier) sendSMS(notification *Notification) error {
	// Check if Twilio is configured
	if n.smsConfig.AccountSID == "" {
		log.Printf("[Notifier] Twilio not configured, skipping SMS notification")
		return nil // Don't fail if SMS not configured
	}

	// TODO: Implement actual Twilio API call
	// For MVP, just log the SMS
	log.Printf("[Notifier] [SMS] To: %s, Message: %s",
		notification.To, notification.Body)

	// In production, use Twilio SDK:
	/*
	import "github.com/twilio/twilio-go"
	import openapi "github.com/twilio/twilio-go/rest/api/v2010"

	client := twilio.NewRestClientWithParams(twilio.ClientParams{
		Username: n.smsConfig.AccountSID,
		Password: n.smsConfig.AuthToken,
	})

	params := &openapi.CreateMessageParams{}
	params.SetTo(notification.To)
	params.SetFrom(n.smsConfig.FromNumber)
	params.SetBody(notification.Body)

	_, err := client.Api.CreateMessage(params)
	if err != nil {
		return err
	}
	*/

	return nil
}

// sendWebhook posts alert to webhook URL
func (n *Notifier) sendWebhook(notification *Notification) error {
	// TODO: Implement webhook HTTP POST with retry logic
	log.Printf("[Notifier] [WEBHOOK] To: %s, Payload: %s",
		notification.To, notification.Body)

	// In production:
	/*
	import "net/http"
	import "bytes"

	payload, _ := json.Marshal(map[string]interface{}{
		"alert_id": notification.AlertID,
		"account_id": notification.AccountID,
		"subject": notification.Subject,
		"body": notification.Body,
		"timestamp": notification.CreatedAt,
	})

	req, err := http.NewRequest("POST", notification.To, bytes.NewBuffer(payload))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return fmt.Errorf("webhook returned status %d", resp.StatusCode)
	}
	*/

	return nil
}

// formatSubject creates email/SMS subject line
func (n *Notifier) formatSubject(alert *Alert) string {
	return fmt.Sprintf("[%s] %s", alert.Severity, alert.Title)
}

// formatBody creates notification message body
func (n *Notifier) formatBody(alert *Alert) string {
	// For dashboard/webhook, use JSON
	if alert.Type == AlertTypeThreshold {
		return fmt.Sprintf("%s\n\nCurrent value: %.2f\nThreshold: %.2f\nTime: %s",
			alert.Message, alert.Value, alert.Threshold, alert.CreatedAt.Format(time.RFC3339))
	}

	return fmt.Sprintf("%s\n\nTime: %s", alert.Message, alert.CreatedAt.Format(time.RFC3339))
}

// GetQueueStatus returns current notification queue status
func (n *Notifier) GetQueueStatus() map[string]interface{} {
	n.processMu.RLock()
	processingCount := len(n.processing)
	n.processMu.RUnlock()

	return map[string]interface{}{
		"queueLength":    len(n.queue),
		"processing":     processingCount,
		"queueCapacity":  cap(n.queue),
	}
}

// Helper to parse environment variable as int
func getEnvAsInt(key string, defaultVal int) int {
	if val := os.Getenv(key); val != "" {
		var result int
		if _, err := fmt.Sscanf(val, "%d", &result); err == nil {
			return result
		}
	}
	return defaultVal
}

// BroadcastAlert is a method to satisfy WSHub interface for testing
// In production, this would delegate to the actual WebSocket hub
type MockWSHub struct{}

func (m *MockWSHub) BroadcastAlert(alert *Alert) {
	data, _ := json.Marshal(map[string]interface{}{
		"type":      "alert",
		"alert":     alert,
		"timestamp": time.Now().Unix(),
	})
	log.Printf("[MockWSHub] Broadcasting alert: %s", string(data))
}
