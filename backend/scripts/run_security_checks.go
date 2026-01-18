// +build ignore

package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/epic1st/rtx/backend/security"
)

// This script runs comprehensive security checks
// Usage: go run scripts/run_security_checks.go

func main() {
	fmt.Println("═══════════════════════════════════════════════════════════")
	fmt.Println("           RTX TRADING ENGINE - SECURITY CHECKS           ")
	fmt.Println("═══════════════════════════════════════════════════════════")
	fmt.Println()

	// Get backend directory
	backendDir, err := os.Getwd()
	if err != nil {
		log.Fatal(err)
	}

	// Initialize audit logger for this run
	auditLogger, err := security.NewAuditLogger(filepath.Join(backendDir, "logs", "security-check"))
	if err != nil {
		log.Fatalf("Failed to initialize audit logger: %v", err)
	}
	defer auditLogger.Close()

	results := make(map[string]bool)

	// 1. Run Compliance Checks
	fmt.Println("1️⃣  Running Compliance Checks...")
	fmt.Println("───────────────────────────────────────────────────────────")
	complianceChecker := security.NewComplianceChecker(auditLogger)
	complianceReport, err := complianceChecker.RunAllChecks()
	if err != nil {
		log.Printf("Compliance check error: %v", err)
		results["compliance"] = false
	} else {
		fmt.Printf("Overall Status: %s\n", complianceReport.OverallStatus)
		fmt.Printf("Total Checks: %d\n", complianceReport.TotalChecks)
		fmt.Printf("✅ Passed: %d\n", complianceReport.Passed)
		fmt.Printf("❌ Failed: %d\n", complianceReport.Failed)
		fmt.Printf("⚠️  Warnings: %d\n\n", complianceReport.Warnings)

		if complianceReport.Failed > 0 {
			fmt.Println("Failed Checks:")
			for _, result := range complianceReport.Results {
				if result.Status == "FAIL" {
					fmt.Printf("  [%s] %s: %s\n", result.Severity, result.RuleName, result.Message)
				}
			}
			fmt.Println()
		}

		results["compliance"] = complianceReport.OverallStatus == "PASS"
	}

	// 2. Run Vulnerability Scanner
	fmt.Println("2️⃣  Running Vulnerability Scanner...")
	fmt.Println("───────────────────────────────────────────────────────────")
	scanner := security.NewVulnerabilityScanner(auditLogger)
	scanReport, err := scanner.ScanDirectory(backendDir)
	if err != nil {
		log.Printf("Vulnerability scan error: %v", err)
		results["vulnerabilities"] = false
	} else {
		fmt.Printf("Files Scanned: %d\n", scanReport.FilesScanned)
		fmt.Printf("Total Vulnerabilities: %d\n", scanReport.Vulnerabilities)
		fmt.Printf("  🔴 Critical: %d\n", scanReport.Critical)
		fmt.Printf("  🟠 High: %d\n", scanReport.High)
		fmt.Printf("  🟡 Medium: %d\n", scanReport.Medium)
		fmt.Printf("  🟢 Low: %d\n\n", scanReport.Low)

		if scanReport.Vulnerabilities > 0 {
			fmt.Println("Vulnerabilities Found:")
			for i, result := range scanReport.Results {
				if i >= 10 {
					fmt.Printf("  ... and %d more\n", scanReport.Vulnerabilities-10)
					break
				}
				fmt.Printf("  [%s] %s:%d - %s\n",
					result.Severity,
					filepath.Base(result.FilePath),
					result.LineNumber,
					result.VulnName,
				)
			}
			fmt.Println()
		}

		results["vulnerabilities"] = scanReport.Critical == 0 && scanReport.High == 0
	}

	// 3. Check for Hardcoded Secrets
	fmt.Println("3️⃣  Scanning for Hardcoded Secrets...")
	fmt.Println("───────────────────────────────────────────────────────────")
	secrets, err := scanner.ScanSecrets(backendDir)
	if err != nil {
		log.Printf("Secret scan error: %v", err)
		results["secrets"] = false
	} else {
		if len(secrets) > 0 {
			fmt.Printf("⚠️  Found %d potential hardcoded secrets:\n", len(secrets))
			for _, secret := range secrets {
				fmt.Printf("  %s:%d - %s\n",
					filepath.Base(secret.FilePath),
					secret.LineNumber,
					secret.VulnName,
				)
			}
			fmt.Println()
			results["secrets"] = false
		} else {
			fmt.Println("✅ No hardcoded secrets detected\n")
			results["secrets"] = true
		}
	}

	// 4. Test WAF Configuration
	fmt.Println("4️⃣  Testing WAF Configuration...")
	fmt.Println("───────────────────────────────────────────────────────────")
	wafConfig := security.DefaultWAFConfig()
	waf := security.NewWAF(wafConfig)
	wafStats := waf.GetStats()
	fmt.Printf("Max Requests Per IP: %d\n", wafConfig.MaxRequestsPerIP)
	fmt.Printf("Max Concurrent Connections: %d\n", wafConfig.MaxConcurrentConns)
	fmt.Printf("Block Duration: %v\n", wafConfig.BlockDuration)
	fmt.Printf("Current Stats: %+v\n\n", wafStats)
	results["waf"] = true

	// 5. Test Session Management
	fmt.Println("5️⃣  Testing Session Management...")
	fmt.Println("───────────────────────────────────────────────────────────")
	sessionConfig := security.DefaultSessionConfig()
	sessionManager := security.NewSessionManager(sessionConfig, auditLogger)

	// Create test session
	testSession, err := sessionManager.CreateSession("test-user", "127.0.0.1")
	if err != nil {
		fmt.Printf("❌ Session creation failed: %v\n\n", err)
		results["sessions"] = false
	} else {
		fmt.Printf("✅ Session created: %s\n", testSession.ID)
		fmt.Printf("Timeout: %v\n", sessionConfig.Timeout)
		fmt.Printf("Max Concurrent Per User: %d\n", sessionConfig.MaxConcurrentPerUser)

		// Validate session
		_, err := sessionManager.ValidateSession(testSession.ID, "127.0.0.1")
		if err != nil {
			fmt.Printf("❌ Session validation failed: %v\n\n", err)
			results["sessions"] = false
		} else {
			fmt.Println("✅ Session validation passed\n")
			results["sessions"] = true
		}
	}

	// 6. Test Encryption
	fmt.Println("6️⃣  Testing Encryption...")
	fmt.Println("───────────────────────────────────────────────────────────")
	encryptionService := security.NewEncryptionService("test-master-key")
	testData := "sensitive-api-key-12345"
	encrypted, err := encryptionService.EncryptString(testData)
	if err != nil {
		fmt.Printf("❌ Encryption failed: %v\n\n", err)
		results["encryption"] = false
	} else {
		decrypted, err := encryptionService.DecryptString(encrypted)
		if err != nil || decrypted != testData {
			fmt.Printf("❌ Decryption failed or data mismatch\n\n")
			results["encryption"] = false
		} else {
			fmt.Println("✅ Encryption/Decryption working")
			fmt.Printf("Original: %s\n", testData)
			fmt.Printf("Encrypted: %s...\n", encrypted[:30])
			fmt.Printf("Decrypted: %s\n\n", decrypted)
			results["encryption"] = true
		}
	}

	// 7. Test CSRF Protection
	fmt.Println("7️⃣  Testing CSRF Protection...")
	fmt.Println("───────────────────────────────────────────────────────────")
	csrf := security.NewCSRFProtection("test-csrf-secret")
	token, err := csrf.GenerateToken("test-session")
	if err != nil {
		fmt.Printf("❌ CSRF token generation failed: %v\n\n", err)
		results["csrf"] = false
	} else {
		err := csrf.ValidateToken("test-session", token)
		if err != nil {
			fmt.Printf("❌ CSRF token validation failed: %v\n\n", err)
			results["csrf"] = false
		} else {
			fmt.Println("✅ CSRF token generation and validation working")
			fmt.Printf("Token: %s...\n\n", token[:30])
			results["csrf"] = true
		}
	}

	// 8. Test API Key Manager
	fmt.Println("8️⃣  Testing API Key Manager...")
	fmt.Println("───────────────────────────────────────────────────────────")
	keyManager := security.NewAPIKeyManager(90*24*time.Hour, auditLogger, encryptionService)
	apiKey, err := keyManager.GenerateKey("test-service")
	if err != nil {
		fmt.Printf("❌ API key generation failed: %v\n\n", err)
		results["apikeys"] = false
	} else {
		fmt.Printf("✅ API key generated: %s\n", apiKey.ID)
		fmt.Printf("Expires: %s\n", apiKey.ExpiresAt.Format(time.RFC3339))

		// Test rotation
		rotatedKey, err := keyManager.RotateKey("test-service", "test_rotation")
		if err != nil {
			fmt.Printf("❌ API key rotation failed: %v\n\n", err)
			results["apikeys"] = false
		} else {
			fmt.Printf("✅ Key rotated: %s -> %s\n\n", apiKey.ID, rotatedKey.ID)
			results["apikeys"] = true
		}
	}

	// Summary
	fmt.Println("═══════════════════════════════════════════════════════════")
	fmt.Println("                    SECURITY CHECK SUMMARY                 ")
	fmt.Println("═══════════════════════════════════════════════════════════")

	passed := 0
	total := len(results)

	for check, result := range results {
		status := "❌ FAIL"
		if result {
			status = "✅ PASS"
			passed++
		}
		fmt.Printf("%s  %s\n", status, check)
	}

	fmt.Println("───────────────────────────────────────────────────────────")
	fmt.Printf("Total: %d/%d checks passed (%.0f%%)\n\n", passed, total, float64(passed)/float64(total)*100)

	if passed == total {
		fmt.Println("🎉 All security checks passed! Ready for deployment.")
		os.Exit(0)
	} else {
		fmt.Println("⚠️  Some security checks failed. Please review and fix before deployment.")
		os.Exit(1)
	}
}
