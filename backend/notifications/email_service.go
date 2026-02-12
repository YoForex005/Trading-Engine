package notifications

import (
	"crypto/tls"
	"fmt"
	"log"
	"net/smtp"
	"os"
	"strconv"
	"strings"
)

// SimpleEmailService implements the EmailService interface for NotificationManager
// This is a simplified service that reads SMTP config from environment variables
type SimpleEmailService struct {
	host     string
	port     int
	user     string
	password string
	from     string
	useTLS   bool
	isMock   bool
}

// NewSimpleEmailService creates a new email service
// If SMTP environment variables are not set, returns a mock service that logs to stdout
func NewSimpleEmailService() *SimpleEmailService {
	host := os.Getenv("SMTP_HOST")
	portStr := os.Getenv("SMTP_PORT")
	user := os.Getenv("SMTP_USER")
	password := os.Getenv("SMTP_PASSWORD")
	from := os.Getenv("SMTP_FROM")

	// Check if SMTP is configured
	if host == "" || portStr == "" || user == "" || password == "" {
		log.Println("[EmailService] SMTP not configured - using mock mode (emails will be logged to stdout)")
		return &SimpleEmailService{isMock: true}
	}

	// Parse port
	port, err := strconv.Atoi(portStr)
	if err != nil {
		log.Printf("[EmailService] Invalid SMTP_PORT: %v - using mock mode", err)
		return &SimpleEmailService{isMock: true}
	}

	// Default from address
	if from == "" {
		from = user
	}

	// Determine if TLS should be used (port 587 = STARTTLS, 465 = TLS)
	useTLS := port == 587 || port == 465

	log.Printf("[EmailService] SMTP configured: %s:%d (from: %s, TLS: %v)", host, port, from, useTLS)

	return &SimpleEmailService{
		host:     host,
		port:     port,
		user:     user,
		password: password,
		from:     from,
		useTLS:   useTLS,
		isMock:   false,
	}
}

// SendEmail sends an email (implements EmailService interface)
func (s *SimpleEmailService) SendEmail(to, subject, body string) error {
	// Mock mode - just log
	if s.isMock {
		return s.mockSendEmail(to, subject, body)
	}

	// Real SMTP send
	return s.sendViaSMTP(to, subject, body)
}

// mockSendEmail logs email to stdout instead of sending
func (s *SimpleEmailService) mockSendEmail(to, subject, body string) error {
	log.Println("========================================")
	log.Println("[EmailService] Mock Email (not actually sent)")
	log.Printf("[EmailService] To: %s", to)
	log.Printf("[EmailService] Subject: %s", subject)
	log.Printf("[EmailService] Body Length: %d bytes", len(body))
	log.Println("========================================")
	return nil
}

// sendViaSMTP sends email via SMTP server
func (s *SimpleEmailService) sendViaSMTP(to, subject, body string) error {
	// Build message with headers
	msg := s.buildEmailMessage(to, subject, body)

	// SMTP server address
	addr := fmt.Sprintf("%s:%d", s.host, s.port)

	// Authentication
	auth := smtp.PlainAuth("", s.user, s.password, s.host)

	// Send based on port
	switch s.port {
	case 587: // STARTTLS
		return s.sendWithStartTLS(addr, auth, []string{to}, msg)
	case 465: // TLS
		return s.sendWithTLS(addr, auth, []string{to}, msg)
	default: // Plain SMTP (port 25)
		return smtp.SendMail(addr, auth, s.from, []string{to}, msg)
	}
}

// sendWithStartTLS sends email using STARTTLS (port 587)
func (s *SimpleEmailService) sendWithStartTLS(addr string, auth smtp.Auth, to []string, msg []byte) error {
	// Connect to SMTP server
	client, err := smtp.Dial(addr)
	if err != nil {
		return fmt.Errorf("failed to connect to SMTP server: %w", err)
	}
	defer client.Close()

	// Start TLS
	tlsConfig := &tls.Config{
		ServerName:         s.host,
		InsecureSkipVerify: false,
	}
	if err = client.StartTLS(tlsConfig); err != nil {
		return fmt.Errorf("STARTTLS failed: %w", err)
	}

	// Authenticate
	if err = client.Auth(auth); err != nil {
		return fmt.Errorf("SMTP authentication failed: %w", err)
	}

	// Set sender
	if err = client.Mail(s.from); err != nil {
		return fmt.Errorf("failed to set sender: %w", err)
	}

	// Set recipients
	for _, recipient := range to {
		if err = client.Rcpt(recipient); err != nil {
			return fmt.Errorf("failed to set recipient %s: %w", recipient, err)
		}
	}

	// Send message body
	wc, err := client.Data()
	if err != nil {
		return fmt.Errorf("failed to initiate data transfer: %w", err)
	}
	defer wc.Close()

	if _, err = wc.Write(msg); err != nil {
		return fmt.Errorf("failed to write message: %w", err)
	}

	log.Printf("[EmailService] Email sent successfully to %s", to[0])
	return nil
}

// sendWithTLS sends email using TLS wrapper (port 465)
func (s *SimpleEmailService) sendWithTLS(addr string, auth smtp.Auth, to []string, msg []byte) error {
	// Create TLS connection
	tlsConfig := &tls.Config{
		ServerName:         s.host,
		InsecureSkipVerify: false,
	}

	conn, err := tls.Dial("tcp", addr, tlsConfig)
	if err != nil {
		return fmt.Errorf("TLS connection failed: %w", err)
	}
	defer conn.Close()

	// Create SMTP client on TLS connection
	client, err := smtp.NewClient(conn, s.host)
	if err != nil {
		return fmt.Errorf("failed to create SMTP client: %w", err)
	}
	defer client.Close()

	// Authenticate
	if err = client.Auth(auth); err != nil {
		return fmt.Errorf("SMTP authentication failed: %w", err)
	}

	// Set sender and recipients
	if err = client.Mail(s.from); err != nil {
		return fmt.Errorf("failed to set sender: %w", err)
	}

	for _, recipient := range to {
		if err = client.Rcpt(recipient); err != nil {
			return fmt.Errorf("failed to set recipient %s: %w", recipient, err)
		}
	}

	// Send message
	wc, err := client.Data()
	if err != nil {
		return fmt.Errorf("failed to initiate data transfer: %w", err)
	}
	defer wc.Close()

	if _, err = wc.Write(msg); err != nil {
		return fmt.Errorf("failed to write message: %w", err)
	}

	log.Printf("[EmailService] Email sent successfully to %s", to[0])
	return nil
}

// buildEmailMessage constructs the email message with proper headers
func (s *SimpleEmailService) buildEmailMessage(to, subject, body string) []byte {
	// Determine if body is HTML or plain text
	isHTML := strings.Contains(body, "<html>") || strings.Contains(body, "<!DOCTYPE")

	var msg strings.Builder
	msg.WriteString(fmt.Sprintf("From: %s\r\n", s.from))
	msg.WriteString(fmt.Sprintf("To: %s\r\n", to))
	msg.WriteString(fmt.Sprintf("Subject: %s\r\n", subject))
	msg.WriteString("MIME-Version: 1.0\r\n")

	if isHTML {
		msg.WriteString("Content-Type: text/html; charset=\"UTF-8\"\r\n")
	} else {
		msg.WriteString("Content-Type: text/plain; charset=\"UTF-8\"\r\n")
	}

	msg.WriteString("\r\n")
	msg.WriteString(body)

	return []byte(msg.String())
}
