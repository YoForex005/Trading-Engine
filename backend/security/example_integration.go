package security

import (
	"fmt"
	"log"
	"net/http"
	"time"
)

// This file demonstrates how to integrate all security components
// into the RTX Trading Engine

// SecuritySetup initializes all security components
type SecuritySetup struct {
	WAF                *WAF
	CSRF               *CSRFProtection
	SessionManager     *SessionManager
	AuditLogger        *AuditLogger
	EncryptionService  *EncryptionService
	APIKeyManager      *APIKeyManager
	ComplianceChecker  *ComplianceChecker
	VulnerabilityScanner *VulnerabilityScanner
	SecurityMiddleware *SecurityMiddleware
}

// InitializeSecurity sets up all security components with production-ready configuration
func InitializeSecurity() (*SecuritySetup, error) {
	// 1. Initialize Audit Logger (should be first for logging other components)
	auditLogger, err := NewAuditLogger("/var/log/rtx/audit")
	if err != nil {
		return nil, fmt.Errorf("failed to initialize audit logger: %w", err)
	}

	log.Println("[Security] Audit logger initialized")

	// 2. Initialize Encryption Service
	// In production, load master key from secure vault (HashiCorp Vault, AWS KMS, etc.)
	masterKey := getSecureMasterKey() // Helper function to retrieve secure key
	encryptionService := NewEncryptionService(masterKey)

	log.Println("[Security] Encryption service initialized")

	// 3. Initialize WAF
	wafConfig := &WAFConfig{
		MaxRequestsPerMinute:   100,
		MaxRequestsPerIP:       20,
		BurstSize:              10,
		BlockDuration:          15 * time.Minute,
		MaxFailedAttempts:      5,
		MaxRequestBodySize:     1 << 20, // 1MB
		MaxURLLength:           2048,
		MaxHeaderSize:          8192,
		MaxConcurrentConns:     10000,
		SlowLorisTimeout:       30 * time.Second,
		AdminWhitelist:         getAdminWhitelist(), // Load from config
		WhitelistEnabled:       true,
	}
	waf := NewWAF(wafConfig)

	log.Println("[Security] WAF initialized with rate limiting and IP blocking")

	// 4. Initialize CSRF Protection
	csrfSecret := getSecureCSRFSecret() // Load from secure storage
	csrf := NewCSRFProtection(csrfSecret)

	log.Println("[Security] CSRF protection initialized")

	// 5. Initialize Session Manager
	sessionConfig := &SessionConfig{
		Timeout:              30 * time.Minute,
		MaxSessions:          10000,
		MaxConcurrentPerUser: 3,
		RequireIP:            true,
	}
	sessionManager := NewSessionManager(sessionConfig, auditLogger)

	log.Println("[Security] Session manager initialized")

	// 6. Initialize API Key Manager with 90-day rotation
	apiKeyManager := NewAPIKeyManager(
		90*24*time.Hour, // 90 days
		auditLogger,
		encryptionService,
	)

	log.Println("[Security] API key manager initialized with 90-day rotation")

	// 7. Initialize Compliance Checker
	complianceChecker := NewComplianceChecker(auditLogger)

	log.Println("[Security] Compliance checker initialized")

	// 8. Initialize Vulnerability Scanner
	vulnerabilityScanner := NewVulnerabilityScanner(auditLogger)

	log.Println("[Security] Vulnerability scanner initialized")

	// 9. Create Security Middleware
	securityMiddleware := NewSecurityMiddleware(waf, csrf, sessionManager, auditLogger)

	log.Println("[Security] Security middleware initialized")

	return &SecuritySetup{
		WAF:                  waf,
		CSRF:                 csrf,
		SessionManager:       sessionManager,
		AuditLogger:          auditLogger,
		EncryptionService:    encryptionService,
		APIKeyManager:        apiKeyManager,
		ComplianceChecker:    complianceChecker,
		VulnerabilityScanner: vulnerabilityScanner,
		SecurityMiddleware:   securityMiddleware,
	}, nil
}

// RunSecurityChecks performs startup security verification
func (s *SecuritySetup) RunSecurityChecks() error {
	log.Println("[Security] Running security checks...")

	// 1. Run compliance checks
	complianceReport, err := s.ComplianceChecker.RunAllChecks()
	if err != nil {
		return fmt.Errorf("compliance check failed: %w", err)
	}

	if complianceReport.Failed > 0 {
		log.Printf("[Security] WARNING: %d compliance checks failed", complianceReport.Failed)
		for _, result := range complianceReport.Results {
			if result.Status == "FAIL" {
				log.Printf("[Security] FAIL: %s - %s", result.RuleName, result.Message)
			}
		}
	}

	log.Printf("[Security] Compliance: %s (Passed: %d, Failed: %d)",
		complianceReport.OverallStatus,
		complianceReport.Passed,
		complianceReport.Failed,
	)

	// 2. Run vulnerability scan
	scanReport, err := s.VulnerabilityScanner.ScanDirectory(".")
	if err != nil {
		log.Printf("[Security] WARNING: Vulnerability scan failed: %v", err)
	} else {
		log.Printf("[Security] Vulnerability scan: %d files scanned, %d vulnerabilities found (Critical: %d, High: %d)",
			scanReport.FilesScanned,
			scanReport.Vulnerabilities,
			scanReport.Critical,
			scanReport.High,
		)

		if scanReport.Critical > 0 {
			log.Printf("[Security] CRITICAL: %d critical vulnerabilities found!", scanReport.Critical)
			// In production, you might want to fail startup if critical vulnerabilities exist
		}
	}

	log.Println("[Security] Security checks completed")
	return nil
}

// ApplyToRouter applies security middleware to HTTP routes
func (s *SecuritySetup) ApplyToRouter(mux *http.ServeMux) {
	// Example route protection

	// Public endpoints - only security headers
	publicHandler := SecurityHeadersMiddleware(http.HandlerFunc(handlePublic))
	mux.Handle("/", publicHandler)

	// Login endpoint - security headers + input sanitization
	loginHandler := SecurityHeadersMiddleware(
		InputSanitizationMiddleware(
			http.HandlerFunc(handleLogin),
		),
	)
	mux.Handle("/login", loginHandler)

	// API endpoints - full protection (WAF + CSRF + Session)
	apiHandler := s.SecurityMiddleware.Protect(http.HandlerFunc(handleAPI))
	mux.Handle("/api/", apiHandler)

	// Admin endpoints - extra protection (IP whitelist)
	adminHandler := s.SecurityMiddleware.AdminProtect(http.HandlerFunc(handleAdmin))
	mux.Handle("/admin/", adminHandler)

	log.Println("[Security] Security middleware applied to all routes")
}

// Helper functions (implement these based on your infrastructure)

func getSecureMasterKey() string {
	// In production: Load from HashiCorp Vault, AWS Secrets Manager, or env var
	// NEVER hardcode in production!
	// This is just an example
	// return os.Getenv("MASTER_ENCRYPTION_KEY")
	return "your-secure-master-key-from-vault"
}

func getSecureCSRFSecret() string {
	// In production: Load from secure storage
	// return os.Getenv("CSRF_SECRET")
	return "your-secure-csrf-secret"
}

func getAdminWhitelist() []string {
	// In production: Load from config file or database
	return []string{
		"127.0.0.1",
		"::1",
		// Add your admin IPs
	}
}

// Example handlers (replace with your actual handlers)

func handlePublic(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("Public endpoint"))
}

func handleLogin(w http.ResponseWriter, r *http.Request) {
	// Your login logic here
	w.Write([]byte("Login endpoint"))
}

func handleAPI(w http.ResponseWriter, r *http.Request) {
	// Your API logic here
	w.Write([]byte("Protected API endpoint"))
}

func handleAdmin(w http.ResponseWriter, r *http.Request) {
	// Your admin logic here
	w.Write([]byte("Admin endpoint"))
}

// Example: Complete server initialization
func ExampleServerSetup() {
	// Initialize security
	security, err := InitializeSecurity()
	if err != nil {
		log.Fatalf("Failed to initialize security: %v", err)
	}

	// Run security checks
	if err := security.RunSecurityChecks(); err != nil {
		log.Fatalf("Security checks failed: %v", err)
	}

	// Setup router
	mux := http.NewServeMux()
	security.ApplyToRouter(mux)

	// Generate API keys for external services
	oandaKey, _ := security.APIKeyManager.GenerateKey("oanda")
	log.Printf("[Security] Generated API key for OANDA: %s", oandaKey.ID)

	// Start server
	log.Println("[Server] Starting RTX Trading Engine on :8080")
	log.Fatal(http.ListenAndServe(":8080", mux))
}

// Example: Scheduled security tasks
func RunScheduledSecurityTasks(security *SecuritySetup) {
	// Daily: Run compliance checks
	dailyTicker := time.NewTicker(24 * time.Hour)
	go func() {
		for range dailyTicker.C {
			report, _ := security.ComplianceChecker.RunAllChecks()
			log.Printf("[Security] Daily compliance check: %s", report.OverallStatus)
		}
	}()

	// Weekly: Run vulnerability scan
	weeklyTicker := time.NewTicker(7 * 24 * time.Hour)
	go func() {
		for range weeklyTicker.C {
			report, _ := security.VulnerabilityScanner.ScanDirectory(".")
			log.Printf("[Security] Weekly vulnerability scan: %d vulnerabilities found", report.Vulnerabilities)
		}
	}()

	// Monthly: Generate security report
	monthlyTicker := time.NewTicker(30 * 24 * time.Hour)
	go func() {
		for range monthlyTicker.C {
			generateSecurityReport(security)
		}
	}()
}

func generateSecurityReport(security *SecuritySetup) {
	log.Println("[Security] Generating monthly security report...")

	// WAF statistics
	wafStats := security.WAF.GetStats()
	log.Printf("[Security] WAF Stats: %+v", wafStats)

	// Session statistics
	activeSessions := security.SessionManager.GetActiveSessions()
	log.Printf("[Security] Active Sessions: %d", activeSessions)

	// API key rotation status
	rotationStatus := security.APIKeyManager.GetRotationStatus()
	log.Printf("[Security] API Key Rotation Status: %+v", rotationStatus)

	// Blocked IPs
	blockedIPs := security.WAF.GetBlockedIPs()
	log.Printf("[Security] Blocked IPs: %d", len(blockedIPs))

	// Compliance status
	complianceReport, _ := security.ComplianceChecker.RunAllChecks()
	log.Printf("[Security] Compliance: %s", complianceReport.OverallStatus)

	// Save report to file or send to monitoring system
	// ...
}

// Example: Emergency security procedures
func EmergencySecurityLockdown(security *SecuritySetup) {
	log.Println("[Security] EMERGENCY LOCKDOWN INITIATED")

	// 1. Invalidate all sessions
	// (You'd need to add this method to SessionManager)
	log.Println("[Security] Invalidating all sessions...")

	// 2. Rotate all API keys
	log.Println("[Security] Rotating all API keys...")
	services := []string{"oanda", "binance", "yofx"} // Your services
	for _, service := range services {
		security.APIKeyManager.RotateKey(service, "emergency_lockdown")
	}

	// 3. Enable aggressive rate limiting
	log.Println("[Security] Enabling aggressive rate limiting...")
	security.WAF.config.MaxRequestsPerIP = 5
	security.WAF.config.MaxConcurrentConns = 1000

	// 4. Log the incident
	security.AuditLogger.LogSecurityIncident(
		"emergency",
		"security_lockdown",
		"",
		"Emergency security lockdown initiated",
		map[string]interface{}{
			"timestamp": time.Now(),
		},
	)

	log.Println("[Security] Emergency lockdown complete")
}
