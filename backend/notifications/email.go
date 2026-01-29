package notifications

import (
	"bytes"
	"context"
	"fmt"
	"html/template"
	"net/smtp"
	"strings"
	"sync"
	"time"
)

// EmailProvider interface for different email service providers
type EmailProvider interface {
	Send(ctx context.Context, to []string, subject, body, htmlBody string) (string, error)
	SendTemplate(ctx context.Context, to []string, templateID string, data map[string]interface{}) (string, error)
}

// SMTPConfig holds SMTP server configuration
type SMTPConfig struct {
	Host     string
	Port     int
	Username string
	Password string
	From     string
	FromName string
	UseTLS   bool
}

// SMTPProvider implements email sending via SMTP
type SMTPProvider struct {
	config    SMTPConfig
	templates map[string]*template.Template
	mu        sync.RWMutex
}

// NewSMTPProvider creates a new SMTP email provider
func NewSMTPProvider(config SMTPConfig) *SMTPProvider {
	return &SMTPProvider{
		config:    config,
		templates: make(map[string]*template.Template),
	}
}

// LoadTemplate loads an email template
func (p *SMTPProvider) LoadTemplate(name, htmlContent string) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	tmpl, err := template.New(name).Parse(htmlContent)
	if err != nil {
		return fmt.Errorf("failed to parse template: %w", err)
	}

	p.templates[name] = tmpl
	return nil
}

// Send sends a plain text or HTML email
func (p *SMTPProvider) Send(ctx context.Context, to []string, subject, body, htmlBody string) (string, error) {
	auth := smtp.PlainAuth("", p.config.Username, p.config.Password, p.config.Host)

	var msg bytes.Buffer
	msg.WriteString(fmt.Sprintf("From: %s <%s>\r\n", p.config.FromName, p.config.From))
	msg.WriteString(fmt.Sprintf("To: %s\r\n", strings.Join(to, ",")))
	msg.WriteString(fmt.Sprintf("Subject: %s\r\n", subject))
	msg.WriteString("MIME-Version: 1.0\r\n")

	if htmlBody != "" {
		msg.WriteString("Content-Type: multipart/alternative; boundary=\"boundary\"\r\n\r\n")
		msg.WriteString("--boundary\r\n")
		msg.WriteString("Content-Type: text/plain; charset=\"UTF-8\"\r\n\r\n")
		msg.WriteString(body)
		msg.WriteString("\r\n\r\n--boundary\r\n")
		msg.WriteString("Content-Type: text/html; charset=\"UTF-8\"\r\n\r\n")
		msg.WriteString(htmlBody)
		msg.WriteString("\r\n\r\n--boundary--")
	} else {
		msg.WriteString("Content-Type: text/plain; charset=\"UTF-8\"\r\n\r\n")
		msg.WriteString(body)
	}

	addr := fmt.Sprintf("%s:%d", p.config.Host, p.config.Port)
	err := smtp.SendMail(addr, auth, p.config.From, to, msg.Bytes())
	if err != nil {
		return "", fmt.Errorf("failed to send email: %w", err)
	}

	// Generate a message ID (in production, use provider's response)
	messageID := fmt.Sprintf("%d-%s", time.Now().UnixNano(), to[0])
	return messageID, nil
}

// SendTemplate renders and sends an email using a template
func (p *SMTPProvider) SendTemplate(ctx context.Context, to []string, templateID string, data map[string]interface{}) (string, error) {
	p.mu.RLock()
	tmpl, exists := p.templates[templateID]
	p.mu.RUnlock()

	if !exists {
		return "", fmt.Errorf("template not found: %s", templateID)
	}

	var htmlBody bytes.Buffer
	if err := tmpl.Execute(&htmlBody, data); err != nil {
		return "", fmt.Errorf("failed to render template: %w", err)
	}

	// Extract subject from data or use default
	subject, _ := data["subject"].(string)
	if subject == "" {
		subject = "Notification from Trading Platform"
	}

	// Generate plain text version (simple strip tags for now)
	plainText := stripHTML(htmlBody.String())

	return p.Send(ctx, to, subject, plainText, htmlBody.String())
}

// EmailNotifier handles email notifications
type EmailNotifier struct {
	provider      EmailProvider
	defaultFrom   string
	defaultFromName string
	rateLimiter   *RateLimiter
}

// NewEmailNotifier creates a new email notifier
func NewEmailNotifier(provider EmailProvider, from, fromName string, rateLimiter *RateLimiter) *EmailNotifier {
	return &EmailNotifier{
		provider:        provider,
		defaultFrom:     from,
		defaultFromName: fromName,
		rateLimiter:     rateLimiter,
	}
}

// Send sends an email notification
func (n *EmailNotifier) Send(ctx context.Context, notif *Notification, recipient string) (*DeliveryRecord, error) {
	// Check rate limit
	if !n.rateLimiter.Allow(notif.UserID, ChannelEmail) {
		return nil, fmt.Errorf("rate limit exceeded for user %s", notif.UserID)
	}

	record := &DeliveryRecord{
		ID:             generateID(),
		NotificationID: notif.ID,
		UserID:         notif.UserID,
		Channel:        ChannelEmail,
		Status:         StatusPending,
		Attempts:       1,
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}

	// Extract email body from notification data
	body := notif.Message
	htmlBody, _ := notif.Data["html_body"].(string)

	now := time.Now()
	record.LastAttemptAt = &now

	messageID, err := n.provider.Send(ctx, []string{recipient}, notif.Subject, body, htmlBody)
	if err != nil {
		record.Status = StatusFailed
		record.Error = err.Error()
		return record, err
	}

	record.Status = StatusSent
	record.ProviderID = messageID
	record.DeliveredAt = &now

	return record, nil
}

// SendBatch sends multiple emails efficiently
func (n *EmailNotifier) SendBatch(ctx context.Context, notifications []Notification, recipient string) ([]*DeliveryRecord, error) {
	records := make([]*DeliveryRecord, 0, len(notifications))

	for _, notif := range notifications {
		record, err := n.Send(ctx, &notif, recipient)
		if err != nil {
			// Log error but continue with other notifications
			continue
		}
		records = append(records, record)
	}

	return records, nil
}

// Common email templates

const (
	// Welcome email template
	WelcomeEmailTemplate = `
<!DOCTYPE html>
<html>
<head>
    <style>
        body { font-family: Arial, sans-serif; line-height: 1.6; color: #333; }
        .container { max-width: 600px; margin: 0 auto; padding: 20px; }
        .header { background: #4CAF50; color: white; padding: 20px; text-align: center; }
        .content { padding: 20px; background: #f9f9f9; }
        .button { display: inline-block; padding: 10px 20px; background: #4CAF50; color: white; text-decoration: none; border-radius: 5px; }
        .footer { text-align: center; padding: 20px; font-size: 12px; color: #666; }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <h1>Welcome to {{.PlatformName}}</h1>
        </div>
        <div class="content">
            <h2>Hello {{.UserName}},</h2>
            <p>Welcome to our trading platform! Your account has been successfully created.</p>
            <p>You can now start trading with access to:</p>
            <ul>
                <li>Real-time market data</li>
                <li>Advanced charting tools</li>
                <li>Risk management features</li>
                <li>24/7 customer support</li>
            </ul>
            <p style="text-align: center;">
                <a href="{{.LoginURL}}" class="button">Get Started</a>
            </p>
        </div>
        <div class="footer">
            <p>If you have any questions, contact us at {{.SupportEmail}}</p>
            <p><a href="{{.UnsubscribeURL}}">Unsubscribe</a></p>
        </div>
    </div>
</body>
</html>`

	// Order executed template
	OrderExecutedTemplate = `
<!DOCTYPE html>
<html>
<head>
    <style>
        body { font-family: Arial, sans-serif; line-height: 1.6; color: #333; }
        .container { max-width: 600px; margin: 0 auto; padding: 20px; }
        .header { background: #2196F3; color: white; padding: 20px; text-align: center; }
        .content { padding: 20px; background: #f9f9f9; }
        .details { background: white; padding: 15px; border-left: 4px solid #2196F3; margin: 20px 0; }
        .details-row { display: flex; justify-content: space-between; padding: 5px 0; }
        .label { font-weight: bold; }
        .footer { text-align: center; padding: 20px; font-size: 12px; color: #666; }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <h1>Order Executed</h1>
        </div>
        <div class="content">
            <p>Your order has been executed successfully.</p>
            <div class="details">
                <div class="details-row">
                    <span class="label">Order ID:</span>
                    <span>{{.OrderID}}</span>
                </div>
                <div class="details-row">
                    <span class="label">Symbol:</span>
                    <span>{{.Symbol}}</span>
                </div>
                <div class="details-row">
                    <span class="label">Type:</span>
                    <span>{{.OrderType}}</span>
                </div>
                <div class="details-row">
                    <span class="label">Side:</span>
                    <span>{{.Side}}</span>
                </div>
                <div class="details-row">
                    <span class="label">Quantity:</span>
                    <span>{{.Quantity}}</span>
                </div>
                <div class="details-row">
                    <span class="label">Price:</span>
                    <span>{{.Price}}</span>
                </div>
                <div class="details-row">
                    <span class="label">Time:</span>
                    <span>{{.ExecutedAt}}</span>
                </div>
            </div>
        </div>
        <div class="footer">
            <p><a href="{{.UnsubscribeURL}}">Unsubscribe</a></p>
        </div>
    </div>
</body>
</html>`

	// Margin call warning template
	MarginCallTemplate = `
<!DOCTYPE html>
<html>
<head>
    <style>
        body { font-family: Arial, sans-serif; line-height: 1.6; color: #333; }
        .container { max-width: 600px; margin: 0 auto; padding: 20px; }
        .header { background: #FF5722; color: white; padding: 20px; text-align: center; }
        .warning { background: #FFF3CD; border-left: 4px solid #FF5722; padding: 15px; margin: 20px 0; }
        .content { padding: 20px; background: #f9f9f9; }
        .button { display: inline-block; padding: 10px 20px; background: #FF5722; color: white; text-decoration: none; border-radius: 5px; }
        .footer { text-align: center; padding: 20px; font-size: 12px; color: #666; }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <h1>⚠️ Margin Call Warning</h1>
        </div>
        <div class="content">
            <div class="warning">
                <h3>Immediate Action Required</h3>
                <p>Your account margin level has fallen below the required threshold.</p>
                <p><strong>Current Margin Level: {{.MarginLevel}}%</strong></p>
                <p><strong>Required Margin Level: {{.RequiredMarginLevel}}%</strong></p>
            </div>
            <p>To avoid automatic position closure (stop-out), please take one of the following actions:</p>
            <ul>
                <li>Deposit additional funds to your account</li>
                <li>Close some open positions</li>
                <li>Reduce position sizes</li>
            </ul>
            <p style="text-align: center;">
                <a href="{{.AccountURL}}" class="button">Manage Account</a>
            </p>
        </div>
        <div class="footer">
            <p>This is a critical notification and cannot be unsubscribed from.</p>
        </div>
    </div>
</body>
</html>`
)

// stripHTML removes HTML tags from a string (basic implementation)
func stripHTML(html string) string {
	// Simple implementation - in production use a proper HTML parser
	text := strings.ReplaceAll(html, "<br>", "\n")
	text = strings.ReplaceAll(text, "<br/>", "\n")
	text = strings.ReplaceAll(text, "</p>", "\n")

	inTag := false
	var result strings.Builder
	for _, r := range text {
		if r == '<' {
			inTag = true
			continue
		}
		if r == '>' {
			inTag = false
			continue
		}
		if !inTag {
			result.WriteRune(r)
		}
	}

	return result.String()
}
