package admin

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// EmailTemplate represents an email template with all metadata
type EmailTemplate struct {
	ID           string            `json:"id"`
	Name         string            `json:"name"`
	Slug         string            `json:"slug"`
	Subject      string            `json:"subject"`
	HTMLBody     string            `json:"htmlBody"`
	TextBody     string            `json:"textBody"`
	Category     string            `json:"category"` // account/trading/payment/security/marketing/system
	Variables    []string          `json:"variables"` // Available merge tags like {{client_name}}, {{amount}}
	Status       string            `json:"status"` // active/draft/archived
	LastModified time.Time         `json:"lastModified"`
	ModifiedBy   string            `json:"modifiedBy"`
}

// EmailTemplateService manages email templates
type EmailTemplateService struct {
	mu        sync.RWMutex
	templates map[string]*EmailTemplate
	nextID    int
}

// NewEmailTemplateService creates a new email template service
func NewEmailTemplateService() *EmailTemplateService {
	svc := &EmailTemplateService{
		templates: make(map[string]*EmailTemplate),
		nextID:    21,
	}

	// Create 20 default templates
	svc.createDefaultTemplates()

	return svc
}

func (s *EmailTemplateService) createDefaultTemplates() {
	templates := []struct {
		name      string
		slug      string
		subject   string
		htmlBody  string
		textBody  string
		category  string
		variables []string
	}{
		{
			name:      "Welcome Email",
			slug:      "welcome-email",
			subject:   "Welcome to {{platform_name}}!",
			htmlBody:  "<h1>Welcome, {{client_name}}!</h1><p>Thank you for joining {{platform_name}}. Your account {{account_number}} has been successfully created.</p><p>Get started by making your first deposit and exploring our trading platform.</p>",
			textBody:  "Welcome, {{client_name}}! Thank you for joining {{platform_name}}. Your account {{account_number}} has been successfully created. Get started by making your first deposit and exploring our trading platform.",
			category:  "account",
			variables: []string{"{{client_name}}", "{{platform_name}}", "{{account_number}}", "{{login_url}}"},
		},
		{
			name:      "Account Verification",
			slug:      "account-verification",
			subject:   "Verify Your {{platform_name}} Account",
			htmlBody:  "<h1>Verify Your Email</h1><p>Hi {{client_name}},</p><p>Please verify your email address by clicking the link below:</p><p><a href='{{verification_link}}'>Verify Email</a></p><p>This link expires in 24 hours.</p>",
			textBody:  "Hi {{client_name}}, Please verify your email address by clicking the link: {{verification_link}}. This link expires in 24 hours.",
			category:  "account",
			variables: []string{"{{client_name}}", "{{platform_name}}", "{{verification_link}}", "{{expiry_time}}"},
		},
		{
			name:      "Password Reset",
			slug:      "password-reset",
			subject:   "Reset Your Password",
			htmlBody:  "<h1>Password Reset Request</h1><p>Hi {{client_name}},</p><p>We received a request to reset your password. Click the link below to set a new password:</p><p><a href='{{reset_link}}'>Reset Password</a></p><p>If you didn't request this, please ignore this email.</p>",
			textBody:  "Hi {{client_name}}, We received a request to reset your password. Click the link to set a new password: {{reset_link}}. If you didn't request this, please ignore this email.",
			category:  "security",
			variables: []string{"{{client_name}}", "{{reset_link}}", "{{expiry_time}}", "{{ip_address}}"},
		},
		{
			name:      "Deposit Confirmed",
			slug:      "deposit-confirmed",
			subject:   "Deposit of {{amount}} {{currency}} Confirmed",
			htmlBody:  "<h1>Deposit Successful</h1><p>Hi {{client_name}},</p><p>Your deposit of {{amount}} {{currency}} has been confirmed and credited to your account {{account_number}}.</p><p>Transaction ID: {{transaction_id}}</p><p>New Balance: {{new_balance}} {{currency}}</p>",
			textBody:  "Hi {{client_name}}, Your deposit of {{amount}} {{currency}} has been confirmed and credited to your account {{account_number}}. Transaction ID: {{transaction_id}}. New Balance: {{new_balance}} {{currency}}.",
			category:  "payment",
			variables: []string{"{{client_name}}", "{{amount}}", "{{currency}}", "{{account_number}}", "{{transaction_id}}", "{{new_balance}}", "{{deposit_method}}"},
		},
		{
			name:      "Withdrawal Approved",
			slug:      "withdrawal-approved",
			subject:   "Withdrawal of {{amount}} {{currency}} Approved",
			htmlBody:  "<h1>Withdrawal Approved</h1><p>Hi {{client_name}},</p><p>Your withdrawal request of {{amount}} {{currency}} has been approved and is being processed.</p><p>Transaction ID: {{transaction_id}}</p><p>Expected arrival: {{expected_arrival}}</p>",
			textBody:  "Hi {{client_name}}, Your withdrawal request of {{amount}} {{currency}} has been approved and is being processed. Transaction ID: {{transaction_id}}. Expected arrival: {{expected_arrival}}.",
			category:  "payment",
			variables: []string{"{{client_name}}", "{{amount}}", "{{currency}}", "{{transaction_id}}", "{{expected_arrival}}", "{{withdrawal_method}}"},
		},
		{
			name:      "Withdrawal Rejected",
			slug:      "withdrawal-rejected",
			subject:   "Withdrawal Request Rejected",
			htmlBody:  "<h1>Withdrawal Rejected</h1><p>Hi {{client_name}},</p><p>Your withdrawal request of {{amount}} {{currency}} has been rejected.</p><p>Reason: {{rejection_reason}}</p><p>If you have questions, please contact our support team.</p>",
			textBody:  "Hi {{client_name}}, Your withdrawal request of {{amount}} {{currency}} has been rejected. Reason: {{rejection_reason}}. If you have questions, please contact our support team.",
			category:  "payment",
			variables: []string{"{{client_name}}", "{{amount}}", "{{currency}}", "{{rejection_reason}}", "{{support_email}}"},
		},
		{
			name:      "Margin Call Warning",
			slug:      "margin-call-warning",
			subject:   "⚠️ Margin Call Warning - Account {{account_number}}",
			htmlBody:  "<h1 style='color: #F59E0B;'>⚠️ Margin Call Warning</h1><p>Hi {{client_name}},</p><p>Your account {{account_number}} has reached the margin call level.</p><p>Current Equity: {{equity}} {{currency}}</p><p>Margin Level: {{margin_level}}%</p><p>Please deposit funds or close positions to avoid stop out.</p>",
			textBody:  "⚠️ MARGIN CALL WARNING - Hi {{client_name}}, Your account {{account_number}} has reached the margin call level. Current Equity: {{equity}} {{currency}}. Margin Level: {{margin_level}}%. Please deposit funds or close positions to avoid stop out.",
			category:  "trading",
			variables: []string{"{{client_name}}", "{{account_number}}", "{{equity}}", "{{currency}}", "{{margin_level}}", "{{required_margin}}", "{{free_margin}}"},
		},
		{
			name:      "Stop Out Notification",
			slug:      "stop-out-notification",
			subject:   "🚨 Stop Out Executed - Account {{account_number}}",
			htmlBody:  "<h1 style='color: #EF4444;'>🚨 Stop Out Executed</h1><p>Hi {{client_name}},</p><p>Your account {{account_number}} has reached the stop out level. Positions have been automatically closed:</p><ul>{{closed_positions}}</ul><p>Remaining Balance: {{balance}} {{currency}}</p>",
			textBody:  "🚨 STOP OUT EXECUTED - Hi {{client_name}}, Your account {{account_number}} has reached the stop out level. Positions have been automatically closed. Remaining Balance: {{balance}} {{currency}}.",
			category:  "trading",
			variables: []string{"{{client_name}}", "{{account_number}}", "{{balance}}", "{{currency}}", "{{closed_positions}}", "{{total_loss}}"},
		},
		{
			name:      "Trade Confirmation",
			slug:      "trade-confirmation",
			subject:   "Trade Executed - {{symbol}} {{side}}",
			htmlBody:  "<h1>Trade Confirmation</h1><p>Hi {{client_name}},</p><p>Your trade has been executed:</p><ul><li>Symbol: {{symbol}}</li><li>Side: {{side}}</li><li>Volume: {{volume}}</li><li>Entry Price: {{entry_price}}</li><li>Order ID: {{order_id}}</li></ul>",
			textBody:  "Trade Confirmation - Hi {{client_name}}, Your trade has been executed: Symbol: {{symbol}}, Side: {{side}}, Volume: {{volume}}, Entry Price: {{entry_price}}, Order ID: {{order_id}}.",
			category:  "trading",
			variables: []string{"{{client_name}}", "{{symbol}}", "{{side}}", "{{volume}}", "{{entry_price}}", "{{order_id}}", "{{account_number}}"},
		},
		{
			name:      "Daily Statement",
			slug:      "daily-statement",
			subject:   "Daily Statement - {{date}}",
			htmlBody:  "<h1>Daily Statement</h1><p>Hi {{client_name}},</p><p>Your daily trading statement for {{date}}:</p><ul><li>Opening Balance: {{opening_balance}}</li><li>Profit/Loss: {{pnl}}</li><li>Closing Balance: {{closing_balance}}</li><li>Trades: {{trade_count}}</li></ul>",
			textBody:  "Daily Statement - Hi {{client_name}}, Your daily trading statement for {{date}}: Opening Balance: {{opening_balance}}, Profit/Loss: {{pnl}}, Closing Balance: {{closing_balance}}, Trades: {{trade_count}}.",
			category:  "trading",
			variables: []string{"{{client_name}}", "{{date}}", "{{opening_balance}}", "{{pnl}}", "{{closing_balance}}", "{{trade_count}}", "{{currency}}"},
		},
		{
			name:      "Monthly Statement",
			slug:      "monthly-statement",
			subject:   "Monthly Statement - {{month}}",
			htmlBody:  "<h1>Monthly Statement</h1><p>Hi {{client_name}},</p><p>Your monthly trading statement for {{month}}:</p><ul><li>Opening Balance: {{opening_balance}}</li><li>Total Profit/Loss: {{total_pnl}}</li><li>Closing Balance: {{closing_balance}}</li><li>Total Trades: {{total_trades}}</li><li>Win Rate: {{win_rate}}%</li></ul>",
			textBody:  "Monthly Statement - Hi {{client_name}}, Your monthly statement for {{month}}: Opening Balance: {{opening_balance}}, Total P/L: {{total_pnl}}, Closing Balance: {{closing_balance}}, Total Trades: {{total_trades}}, Win Rate: {{win_rate}}%.",
			category:  "trading",
			variables: []string{"{{client_name}}", "{{month}}", "{{opening_balance}}", "{{total_pnl}}", "{{closing_balance}}", "{{total_trades}}", "{{win_rate}}", "{{currency}}"},
		},
		{
			name:      "KYC Approved",
			slug:      "kyc-approved",
			subject:   "KYC Verification Approved",
			htmlBody:  "<h1>KYC Approved</h1><p>Hi {{client_name}},</p><p>Your KYC verification has been approved! Your account is now fully verified.</p><p>You can now:</p><ul><li>Make deposits and withdrawals</li><li>Access all trading features</li><li>Enjoy higher limits</li></ul>",
			textBody:  "Hi {{client_name}}, Your KYC verification has been approved! Your account is now fully verified. You can now make deposits and withdrawals, access all trading features, and enjoy higher limits.",
			category:  "account",
			variables: []string{"{{client_name}}", "{{account_number}}", "{{verification_date}}", "{{platform_name}}"},
		},
		{
			name:      "KYC Rejected",
			slug:      "kyc-rejected",
			subject:   "KYC Verification Requires Additional Information",
			htmlBody:  "<h1>KYC Verification Update</h1><p>Hi {{client_name}},</p><p>We need additional information to verify your account.</p><p>Reason: {{rejection_reason}}</p><p>Please upload the requested documents in your account settings.</p>",
			textBody:  "Hi {{client_name}}, We need additional information to verify your account. Reason: {{rejection_reason}}. Please upload the requested documents in your account settings.",
			category:  "account",
			variables: []string{"{{client_name}}", "{{rejection_reason}}", "{{required_documents}}", "{{support_email}}"},
		},
		{
			name:      "Promotion Announcement",
			slug:      "promotion-announcement",
			subject:   "🎉 {{promotion_title}}",
			htmlBody:  "<h1>🎉 {{promotion_title}}</h1><p>Hi {{client_name}},</p><p>{{promotion_description}}</p><p>Promotion Period: {{start_date}} - {{end_date}}</p><p>Terms and conditions apply.</p>",
			textBody:  "🎉 {{promotion_title}} - Hi {{client_name}}, {{promotion_description}}. Promotion Period: {{start_date}} - {{end_date}}. Terms and conditions apply.",
			category:  "marketing",
			variables: []string{"{{client_name}}", "{{promotion_title}}", "{{promotion_description}}", "{{start_date}}", "{{end_date}}", "{{terms_url}}"},
		},
		{
			name:      "Referral Bonus",
			slug:      "referral-bonus",
			subject:   "Referral Bonus Credited - {{amount}} {{currency}}",
			htmlBody:  "<h1>Referral Bonus</h1><p>Hi {{client_name}},</p><p>Great news! You've earned a referral bonus of {{amount}} {{currency}}.</p><p>Referred User: {{referred_user}}</p><p>The bonus has been credited to your account {{account_number}}.</p>",
			textBody:  "Hi {{client_name}}, Great news! You've earned a referral bonus of {{amount}} {{currency}}. Referred User: {{referred_user}}. The bonus has been credited to your account {{account_number}}.",
			category:  "marketing",
			variables: []string{"{{client_name}}", "{{amount}}", "{{currency}}", "{{referred_user}}", "{{account_number}}", "{{total_referrals}}"},
		},
		{
			name:      "2FA Setup",
			slug:      "2fa-setup",
			subject:   "Two-Factor Authentication Enabled",
			htmlBody:  "<h1>2FA Enabled</h1><p>Hi {{client_name}},</p><p>Two-factor authentication has been successfully enabled on your account.</p><p>Your account is now more secure. You'll need your authentication code when logging in.</p><p>If you didn't enable this, please contact support immediately.</p>",
			textBody:  "Hi {{client_name}}, Two-factor authentication has been successfully enabled on your account. Your account is now more secure. If you didn't enable this, please contact support immediately.",
			category:  "security",
			variables: []string{"{{client_name}}", "{{enabled_date}}", "{{device}}", "{{ip_address}}", "{{support_email}}"},
		},
		{
			name:      "Login Alert",
			slug:      "login-alert",
			subject:   "New Login Detected - {{device}}",
			htmlBody:  "<h1>New Login Detected</h1><p>Hi {{client_name}},</p><p>A new login was detected on your account:</p><ul><li>Time: {{login_time}}</li><li>Device: {{device}}</li><li>Location: {{location}}</li><li>IP Address: {{ip_address}}</li></ul><p>If this wasn't you, please secure your account immediately.</p>",
			textBody:  "New Login Detected - Hi {{client_name}}, A new login was detected on your account. Time: {{login_time}}, Device: {{device}}, Location: {{location}}, IP: {{ip_address}}. If this wasn't you, please secure your account immediately.",
			category:  "security",
			variables: []string{"{{client_name}}", "{{login_time}}", "{{device}}", "{{location}}", "{{ip_address}}", "{{secure_account_url}}"},
		},
		{
			name:      "Account Suspended",
			slug:      "account-suspended",
			subject:   "Account Temporarily Suspended",
			htmlBody:  "<h1>Account Suspended</h1><p>Hi {{client_name}},</p><p>Your account {{account_number}} has been temporarily suspended.</p><p>Reason: {{suspension_reason}}</p><p>To resolve this issue, please contact our support team at {{support_email}}.</p>",
			textBody:  "Hi {{client_name}}, Your account {{account_number}} has been temporarily suspended. Reason: {{suspension_reason}}. To resolve this issue, please contact our support team at {{support_email}}.",
			category:  "security",
			variables: []string{"{{client_name}}", "{{account_number}}", "{{suspension_reason}}", "{{suspension_date}}", "{{support_email}}"},
		},
		{
			name:      "Maintenance Notice",
			slug:      "maintenance-notice",
			subject:   "Scheduled Maintenance - {{maintenance_date}}",
			htmlBody:  "<h1>Scheduled Maintenance</h1><p>Hi {{client_name}},</p><p>Our platform will be undergoing scheduled maintenance:</p><ul><li>Start: {{start_time}}</li><li>End: {{end_time}}</li><li>Duration: {{duration}}</li></ul><p>During this time, trading may be temporarily unavailable. We apologize for any inconvenience.</p>",
			textBody:  "Scheduled Maintenance - Hi {{client_name}}, Our platform will be undergoing scheduled maintenance from {{start_time}} to {{end_time}} ({{duration}}). During this time, trading may be temporarily unavailable.",
			category:  "system",
			variables: []string{"{{client_name}}", "{{maintenance_date}}", "{{start_time}}", "{{end_time}}", "{{duration}}", "{{platform_name}}"},
		},
		{
			name:      "Newsletter",
			slug:      "newsletter",
			subject:   "{{newsletter_title}}",
			htmlBody:  "<h1>{{newsletter_title}}</h1><p>Hi {{client_name}},</p><div>{{newsletter_content}}</div><p>Stay connected with {{platform_name}}!</p>",
			textBody:  "{{newsletter_title}} - Hi {{client_name}}, {{newsletter_content}}. Stay connected with {{platform_name}}!",
			category:  "marketing",
			variables: []string{"{{client_name}}", "{{newsletter_title}}", "{{newsletter_content}}", "{{platform_name}}", "{{unsubscribe_url}}"},
		},
	}

	for i, tmpl := range templates {
		id := fmt.Sprintf("tmpl-%d", i+1)
		s.templates[id] = &EmailTemplate{
			ID:           id,
			Name:         tmpl.name,
			Slug:         tmpl.slug,
			Subject:      tmpl.subject,
			HTMLBody:     tmpl.htmlBody,
			TextBody:     tmpl.textBody,
			Category:     tmpl.category,
			Variables:    tmpl.variables,
			Status:       "active",
			LastModified: time.Now(),
			ModifiedBy:   "SYSTEM",
		}
	}

	log.Printf("[EmailTemplates] Created 20 default templates")
}

// ListTemplates returns all templates with optional category filter
func (s *EmailTemplateService) ListTemplates(category string) []*EmailTemplate {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var result []*EmailTemplate
	for _, tmpl := range s.templates {
		if category == "" || tmpl.Category == category {
			if tmpl.Status != "archived" {
				result = append(result, tmpl)
			}
		}
	}

	return result
}

// GetTemplate returns a single template by ID
func (s *EmailTemplateService) GetTemplate(id string) (*EmailTemplate, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	tmpl, ok := s.templates[id]
	return tmpl, ok
}

// UpdateTemplate updates a template
func (s *EmailTemplateService) UpdateTemplate(id string, name, subject, htmlBody, textBody, modifiedBy string) (*EmailTemplate, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	tmpl, ok := s.templates[id]
	if !ok {
		return nil, fmt.Errorf("template not found")
	}

	if name != "" {
		tmpl.Name = name
	}
	if subject != "" {
		tmpl.Subject = subject
	}
	if htmlBody != "" {
		tmpl.HTMLBody = htmlBody
	}
	if textBody != "" {
		tmpl.TextBody = textBody
	}
	tmpl.LastModified = time.Now()
	tmpl.ModifiedBy = modifiedBy

	log.Printf("[EmailTemplates] Template %s updated by %s", id, modifiedBy)

	return tmpl, nil
}

// CreateTemplate creates a new custom template
func (s *EmailTemplateService) CreateTemplate(name, slug, subject, htmlBody, textBody, category string, variables []string, createdBy string) (*EmailTemplate, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Generate new ID
	id := fmt.Sprintf("tmpl-%d", s.nextID)
	s.nextID++

	tmpl := &EmailTemplate{
		ID:           id,
		Name:         name,
		Slug:         slug,
		Subject:      subject,
		HTMLBody:     htmlBody,
		TextBody:     textBody,
		Category:     category,
		Variables:    variables,
		Status:       "draft",
		LastModified: time.Now(),
		ModifiedBy:   createdBy,
	}

	s.templates[id] = tmpl

	log.Printf("[EmailTemplates] Created new template %s by %s", id, createdBy)

	return tmpl, nil
}

// ArchiveTemplate soft-deletes a template
func (s *EmailTemplateService) ArchiveTemplate(id string, archivedBy string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	tmpl, ok := s.templates[id]
	if !ok {
		return fmt.Errorf("template not found")
	}

	tmpl.Status = "archived"
	tmpl.LastModified = time.Now()
	tmpl.ModifiedBy = archivedBy

	log.Printf("[EmailTemplates] Template %s archived by %s", id, archivedBy)

	return nil
}

// GetCategories returns category list with template counts
func (s *EmailTemplateService) GetCategories() map[string]int {
	s.mu.RLock()
	defer s.mu.RUnlock()

	categories := make(map[string]int)
	for _, tmpl := range s.templates {
		if tmpl.Status != "archived" {
			categories[tmpl.Category]++
		}
	}

	return categories
}

// RenderPreview renders a template with sample data
func (s *EmailTemplateService) RenderPreview(id string, sampleData map[string]string) (string, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	tmpl, ok := s.templates[id]
	if !ok {
		return "", fmt.Errorf("template not found")
	}

	// Simple variable replacement for preview
	htmlBody := tmpl.HTMLBody
	for key, value := range sampleData {
		htmlBody = strings.ReplaceAll(htmlBody, key, value)
	}

	return htmlBody, nil
}

// =====================
// HTTP HANDLERS
// =====================

// HandleListTemplates handles GET /admin/email-templates
func (s *EmailTemplateService) HandleListTemplates(w http.ResponseWriter, r *http.Request) {
	// CORS
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	// JWT Auth
	tokenString := r.Header.Get("Authorization")
	if tokenString == "" {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	tokenString = strings.TrimPrefix(tokenString, "Bearer ")

	_, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		return []byte("rtx-secret-key"), nil
	})
	if err != nil {
		http.Error(w, "Invalid token", http.StatusUnauthorized)
		return
	}

	// Get category filter
	category := r.URL.Query().Get("category")

	templates := s.ListTemplates(category)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success":   true,
		"templates": templates,
		"total":     len(templates),
	})
}

// HandleGetTemplate handles GET /admin/email-templates/:id
func (s *EmailTemplateService) HandleGetTemplate(w http.ResponseWriter, r *http.Request) {
	// CORS
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	// JWT Auth
	tokenString := r.Header.Get("Authorization")
	if tokenString == "" {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	tokenString = strings.TrimPrefix(tokenString, "Bearer ")

	_, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		return []byte("rtx-secret-key"), nil
	})
	if err != nil {
		http.Error(w, "Invalid token", http.StatusUnauthorized)
		return
	}

	// Extract ID from path
	parts := strings.Split(r.URL.Path, "/")
	if len(parts) < 4 {
		http.Error(w, "Invalid template ID", http.StatusBadRequest)
		return
	}
	id := parts[3]

	tmpl, ok := s.GetTemplate(id)
	if !ok {
		http.Error(w, "Template not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success":  true,
		"template": tmpl,
	})
}

// HandleUpdateTemplate handles PUT /admin/email-templates/:id
func (s *EmailTemplateService) HandleUpdateTemplate(w http.ResponseWriter, r *http.Request) {
	// CORS
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "PUT, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	// JWT Auth
	tokenString := r.Header.Get("Authorization")
	if tokenString == "" {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	tokenString = strings.TrimPrefix(tokenString, "Bearer ")

	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		return []byte("rtx-secret-key"), nil
	})
	if err != nil {
		http.Error(w, "Invalid token", http.StatusUnauthorized)
		return
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		http.Error(w, "Invalid token claims", http.StatusUnauthorized)
		return
	}
	username := claims["username"].(string)

	// Extract ID from path
	parts := strings.Split(r.URL.Path, "/")
	if len(parts) < 4 {
		http.Error(w, "Invalid template ID", http.StatusBadRequest)
		return
	}
	id := parts[3]

	var req struct {
		Name     string `json:"name"`
		Subject  string `json:"subject"`
		HTMLBody string `json:"htmlBody"`
		TextBody string `json:"textBody"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	tmpl, err := s.UpdateTemplate(id, req.Name, req.Subject, req.HTMLBody, req.TextBody, username)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success":  true,
		"template": tmpl,
		"message":  "Template updated successfully",
	})
}

// HandlePreviewTemplate handles POST /admin/email-templates/:id/preview
func (s *EmailTemplateService) HandlePreviewTemplate(w http.ResponseWriter, r *http.Request) {
	// CORS
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	// JWT Auth
	tokenString := r.Header.Get("Authorization")
	if tokenString == "" {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	tokenString = strings.TrimPrefix(tokenString, "Bearer ")

	_, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		return []byte("rtx-secret-key"), nil
	})
	if err != nil {
		http.Error(w, "Invalid token", http.StatusUnauthorized)
		return
	}

	// Extract ID from path
	parts := strings.Split(r.URL.Path, "/")
	if len(parts) < 4 {
		http.Error(w, "Invalid template ID", http.StatusBadRequest)
		return
	}
	id := parts[3]

	var req struct {
		SampleData map[string]string `json:"sampleData"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	preview, err := s.RenderPreview(id, req.SampleData)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"preview": preview,
	})
}

// HandleSendTestEmail handles POST /admin/email-templates/:id/test
func (s *EmailTemplateService) HandleSendTestEmail(w http.ResponseWriter, r *http.Request) {
	// CORS
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	// JWT Auth
	tokenString := r.Header.Get("Authorization")
	if tokenString == "" {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	tokenString = strings.TrimPrefix(tokenString, "Bearer ")

	_, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		return []byte("rtx-secret-key"), nil
	})
	if err != nil {
		http.Error(w, "Invalid token", http.StatusUnauthorized)
		return
	}

	// Extract ID from path
	parts := strings.Split(r.URL.Path, "/")
	if len(parts) < 4 {
		http.Error(w, "Invalid template ID", http.StatusBadRequest)
		return
	}
	id := parts[3]

	var req struct {
		TestEmail  string            `json:"testEmail"`
		SampleData map[string]string `json:"sampleData"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// In production, this would send actual email via SMTP
	log.Printf("[EmailTemplates] Test email sent for template %s to %s", id, req.TestEmail)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": fmt.Sprintf("Test email sent to %s", req.TestEmail),
	})
}

// HandleCreateTemplate handles POST /admin/email-templates
func (s *EmailTemplateService) HandleCreateTemplate(w http.ResponseWriter, r *http.Request) {
	// CORS
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	// JWT Auth
	tokenString := r.Header.Get("Authorization")
	if tokenString == "" {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	tokenString = strings.TrimPrefix(tokenString, "Bearer ")

	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		return []byte("rtx-secret-key"), nil
	})
	if err != nil {
		http.Error(w, "Invalid token", http.StatusUnauthorized)
		return
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		http.Error(w, "Invalid token claims", http.StatusUnauthorized)
		return
	}
	username := claims["username"].(string)

	var req struct {
		Name      string   `json:"name"`
		Slug      string   `json:"slug"`
		Subject   string   `json:"subject"`
		HTMLBody  string   `json:"htmlBody"`
		TextBody  string   `json:"textBody"`
		Category  string   `json:"category"`
		Variables []string `json:"variables"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	tmpl, err := s.CreateTemplate(req.Name, req.Slug, req.Subject, req.HTMLBody, req.TextBody, req.Category, req.Variables, username)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success":  true,
		"template": tmpl,
		"message":  "Template created successfully",
	})
}

// HandleArchiveTemplate handles DELETE /admin/email-templates/:id
func (s *EmailTemplateService) HandleArchiveTemplate(w http.ResponseWriter, r *http.Request) {
	// CORS
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "DELETE, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	// JWT Auth
	tokenString := r.Header.Get("Authorization")
	if tokenString == "" {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	tokenString = strings.TrimPrefix(tokenString, "Bearer ")

	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		return []byte("rtx-secret-key"), nil
	})
	if err != nil {
		http.Error(w, "Invalid token", http.StatusUnauthorized)
		return
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		http.Error(w, "Invalid token claims", http.StatusUnauthorized)
		return
	}
	username := claims["username"].(string)

	// Extract ID from path
	parts := strings.Split(r.URL.Path, "/")
	if len(parts) < 4 {
		http.Error(w, "Invalid template ID", http.StatusBadRequest)
		return
	}
	id := parts[3]

	if err := s.ArchiveTemplate(id, username); err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": "Template archived successfully",
	})
}

// HandleGetCategories handles GET /admin/email-templates/categories
func (s *EmailTemplateService) HandleGetCategories(w http.ResponseWriter, r *http.Request) {
	// CORS
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	// JWT Auth
	tokenString := r.Header.Get("Authorization")
	if tokenString == "" {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	tokenString = strings.TrimPrefix(tokenString, "Bearer ")

	_, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		return []byte("rtx-secret-key"), nil
	})
	if err != nil {
		http.Error(w, "Invalid token", http.StatusUnauthorized)
		return
	}

	categories := s.GetCategories()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success":    true,
		"categories": categories,
	})
}
