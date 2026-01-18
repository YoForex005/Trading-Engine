package main

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/epic1st/rtx/backend/admin"
	"github.com/epic1st/rtx/backend/fix"
)

func main() {
	// Initialize audit logger
	auditLogger := fix.NewSimpleAuditLogger()

	// Initialize provisioning service
	provisioningService, err := fix.NewProvisioningService(
		"./data/fix_credentials.json",
		"your-master-password-here", // In production, use env var or secret manager
		auditLogger,
	)
	if err != nil {
		log.Fatalf("Failed to initialize provisioning service: %v", err)
	}

	// Initialize FIX manager
	fixManager := admin.NewFIXManager(provisioningService)

	// Example 1: Provision FIX access for a new user
	fmt.Println("=== Example 1: Provision FIX Access ===")
	provisionReq := &admin.ProvisionUserRequest{
		UserID:         "user123",
		RateLimitTier:  "standard",
		MaxSessions:    3,
		ExpiresInDays:  365,
		AllowedIPs:     []string{"192.168.1.100", "192.168.1.101"},
		AccountBalance: 5000.0,
		TradingVolume:  50000.0,
		AccountAgeDays: 90,
		KYCLevel:       2,
		Groups:         []string{"verified-traders"},
		BypassRules:    false,
	}

	provisionResp, err := fixManager.ProvisionUser(provisionReq)
	if err != nil {
		log.Fatalf("Failed to provision user: %v", err)
	}

	if provisionResp.Success {
		fmt.Printf("✓ FIX access provisioned successfully!\n")
		fmt.Printf("  SenderCompID: %s\n", provisionResp.Credentials.SenderCompID)
		fmt.Printf("  TargetCompID: %s\n", provisionResp.Credentials.TargetCompID)
		fmt.Printf("  Password: %s\n", provisionResp.Credentials.Password)
		fmt.Printf("  Tier: %s\n", provisionResp.Credentials.RateLimitTier)
	} else {
		fmt.Printf("✗ Provisioning failed: %s\n", provisionResp.ErrorMessage)
		for _, rule := range provisionResp.FailedRules {
			fmt.Printf("  - %s\n", rule)
		}
	}

	// Example 2: Validate login attempt
	fmt.Println("\n=== Example 2: Validate Login ===")
	if provisionResp.Success {
		creds, err := provisioningService.ValidateLogin(
			provisionResp.Credentials.SenderCompID,
			provisionResp.Credentials.Password,
			"192.168.1.100",
		)
		if err != nil {
			fmt.Printf("✗ Login failed: %v\n", err)
		} else {
			fmt.Printf("✓ Login validated for user: %s\n", creds.UserID)

			// Register session
			sessionID := "session-001"
			err = provisioningService.RegisterSession(
				sessionID,
				creds.UserID,
				creds.SenderCompID,
				"192.168.1.100",
			)
			if err != nil {
				fmt.Printf("✗ Failed to register session: %v\n", err)
			} else {
				fmt.Printf("✓ Session registered: %s\n", sessionID)
			}
		}
	}

	// Example 3: Track message and check rate limits
	fmt.Println("\n=== Example 3: Rate Limiting ===")
	userID := "user123"

	// Simulate sending 10 messages
	for i := 0; i < 10; i++ {
		err := provisioningService.TrackMessage(userID, "session-001", false)
		if err != nil {
			fmt.Printf("✗ Message %d rejected: %v\n", i+1, err)
		} else {
			fmt.Printf("✓ Message %d accepted\n", i+1)
		}
		time.Sleep(50 * time.Millisecond)
	}

	// Check rate limit state
	state, err := fixManager.GetUserRateLimitState(userID)
	if err != nil {
		fmt.Printf("✗ Failed to get rate limit state: %v\n", err)
	} else {
		fmt.Printf("\nRate Limit State:\n")
		fmt.Printf("  Tier: %s\n", state.Tier)
		fmt.Printf("  Messages/sec: %d\n", state.MessagesPerSecond)
		fmt.Printf("  Orders/sec: %d\n", state.OrdersPerSecond)
		fmt.Printf("  Available message tokens: %d\n", state.AvailableMessageTokens)
		fmt.Printf("  Available order tokens: %d\n", state.AvailableOrderTokens)
		fmt.Printf("  Violations: %d\n", state.Violations)
	}

	// Example 4: Manage access rules
	fmt.Println("\n=== Example 4: Access Rules ===")

	// Add custom rule
	customRule := &fix.AccessRule{
		ID:          "rule_custom_vip",
		Name:        "VIP Member Only",
		Description: "User must be in VIP group",
		Type:        fix.RuleTypeGroupMembership,
		Enabled:     true,
		Priority:    110,
		Operator:    "in",
		Value:       []string{"vip", "premium"},
		ErrorMsg:    "VIP membership required for FIX API access",
	}

	err = fixManager.AddAccessRule(customRule)
	if err != nil {
		fmt.Printf("✗ Failed to add rule: %v\n", err)
	} else {
		fmt.Printf("✓ Custom rule added: %s\n", customRule.Name)
	}

	// List all rules
	rules := fixManager.ListAccessRules()
	fmt.Printf("\nActive Rules (%d):\n", len(rules))
	for _, rule := range rules {
		status := "disabled"
		if rule.Enabled {
			status = "enabled"
		}
		fmt.Printf("  - %s (%s) [%s]\n", rule.Name, rule.Type, status)
	}

	// Example 5: Admin operations
	fmt.Println("\n=== Example 5: Admin Operations ===")

	// Get system stats
	stats := fixManager.GetSystemStats()
	fmt.Printf("System Statistics:\n")
	fmt.Printf("  Total users: %d\n", stats.TotalUsers)
	fmt.Printf("  Active users: %d\n", stats.ActiveUsers)
	fmt.Printf("  Total sessions: %d\n", stats.TotalSessions)
	fmt.Printf("  Credentials: Active=%d, Revoked=%d, Suspended=%d\n",
		stats.CredentialStats.Active,
		stats.CredentialStats.Revoked,
		stats.CredentialStats.Suspended)

	// List active sessions
	allSessions := fixManager.GetActiveSessions()
	fmt.Printf("\nActive Sessions:\n")
	for userID, sessions := range allSessions {
		for _, session := range sessions {
			fmt.Printf("  - User: %s, Session: %s, IP: %s, Connected: %v\n",
				userID,
				session.SessionID,
				session.IPAddress,
				session.ConnectedAt.Format(time.RFC3339))
		}
	}

	// Example 6: Start HTTP admin server
	fmt.Println("\n=== Starting HTTP Admin Server ===")
	mux := http.NewServeMux()
	fixManager.RegisterHTTPHandlers(mux)

	fmt.Println("Admin API endpoints:")
	fmt.Println("  POST   /admin/fix/provision")
	fmt.Println("  GET    /admin/fix/credentials")
	fmt.Println("  GET    /admin/fix/sessions")
	fmt.Println("  GET    /admin/fix/rate-limits")
	fmt.Println("  GET    /admin/fix/rules")
	fmt.Println("  GET    /admin/fix/stats")
	fmt.Println("\nServer starting on :8080...")

	if err := http.ListenAndServe(":8080", mux); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}
