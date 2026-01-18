package security

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
	"time"
)

// VulnerabilityScanner scans code for security vulnerabilities
type VulnerabilityScanner struct {
	auditLogger *AuditLogger
	patterns    []VulnerabilityPattern
	mu          sync.Mutex
}

// VulnerabilityPattern defines a security vulnerability pattern
type VulnerabilityPattern struct {
	ID          string
	Name        string
	Description string
	Severity    string // CRITICAL, HIGH, MEDIUM, LOW
	Pattern     *regexp.Regexp
	FileTypes   []string // e.g., [".go", ".js"]
}

// ScanResult represents a vulnerability scan finding
type ScanResult struct {
	Timestamp    time.Time
	FilePath     string
	LineNumber   int
	VulnID       string
	VulnName     string
	Severity     string
	Description  string
	CodeSnippet  string
	Remediation  string
}

// ScanReport summarizes all scan results
type ScanReport struct {
	Timestamp      time.Time
	FilesScanned   int
	Vulnerabilities int
	Critical       int
	High           int
	Medium         int
	Low            int
	Results        []ScanResult
}

// NewVulnerabilityScanner creates a new vulnerability scanner
func NewVulnerabilityScanner(auditLogger *AuditLogger) *VulnerabilityScanner {
	scanner := &VulnerabilityScanner{
		auditLogger: auditLogger,
		patterns:    make([]VulnerabilityPattern, 0),
	}

	scanner.registerDefaultPatterns()
	return scanner
}

// registerDefaultPatterns registers common security vulnerability patterns
func (s *VulnerabilityScanner) registerDefaultPatterns() {
	// Hardcoded secrets
	s.AddPattern(VulnerabilityPattern{
		ID:          "VULN-001",
		Name:        "Hardcoded API Key",
		Description: "Hardcoded API key detected in source code",
		Severity:    "CRITICAL",
		Pattern:     regexp.MustCompile(`(?i)(api[_-]?key|apikey|api[_-]?secret)\s*[:=]\s*["']([a-zA-Z0-9_-]{20,})["']`),
		FileTypes:   []string{".go", ".js", ".ts", ".py", ".java"},
	})

	s.AddPattern(VulnerabilityPattern{
		ID:          "VULN-002",
		Name:        "Hardcoded Password",
		Description: "Hardcoded password detected in source code",
		Severity:    "CRITICAL",
		Pattern:     regexp.MustCompile(`(?i)(password|passwd|pwd)\s*[:=]\s*["']([^"']{5,})["']`),
		FileTypes:   []string{".go", ".js", ".ts", ".py", ".java"},
	})

	s.AddPattern(VulnerabilityPattern{
		ID:          "VULN-003",
		Name:        "Hardcoded JWT Secret",
		Description: "Hardcoded JWT secret detected",
		Severity:    "CRITICAL",
		Pattern:     regexp.MustCompile(`(?i)(jwt[_-]?secret|secret[_-]?key)\s*[:=]\s*["']([a-zA-Z0-9_-]{10,})["']`),
		FileTypes:   []string{".go", ".js", ".ts", ".py", ".java"},
	})

	// SQL Injection vulnerabilities
	s.AddPattern(VulnerabilityPattern{
		ID:          "VULN-004",
		Name:        "Potential SQL Injection",
		Description: "String concatenation in SQL query detected",
		Severity:    "CRITICAL",
		Pattern:     regexp.MustCompile(`(?i)(query|execute|exec)\s*\(\s*["']\s*SELECT.*\+.*["']`),
		FileTypes:   []string{".go", ".js", ".ts", ".py", ".java"},
	})

	// Weak cryptography
	s.AddPattern(VulnerabilityPattern{
		ID:          "VULN-005",
		Name:        "Weak Cryptographic Algorithm",
		Description: "Use of weak cryptographic algorithm (MD5, SHA1)",
		Severity:    "HIGH",
		Pattern:     regexp.MustCompile(`(?i)(md5|sha1)\.New\(\)`),
		FileTypes:   []string{".go"},
	})

	s.AddPattern(VulnerabilityPattern{
		ID:          "VULN-006",
		Name:        "Insecure Random",
		Description: "Use of math/rand for security-sensitive operations",
		Severity:    "HIGH",
		Pattern:     regexp.MustCompile(`math\/rand`),
		FileTypes:   []string{".go"},
	})

	// Command injection
	s.AddPattern(VulnerabilityPattern{
		ID:          "VULN-007",
		Name:        "Potential Command Injection",
		Description: "Unsafe command execution with user input",
		Severity:    "CRITICAL",
		Pattern:     regexp.MustCompile(`exec\.Command\([^)]*\+[^)]*\)`),
		FileTypes:   []string{".go"},
	})

	// Path traversal
	s.AddPattern(VulnerabilityPattern{
		ID:          "VULN-008",
		Name:        "Path Traversal Risk",
		Description: "Potential path traversal vulnerability",
		Severity:    "HIGH",
		Pattern:     regexp.MustCompile(`filepath\.Join\([^)]*userInput[^)]*\)`),
		FileTypes:   []string{".go"},
	})

	// Insecure HTTP
	s.AddPattern(VulnerabilityPattern{
		ID:          "VULN-009",
		Name:        "Insecure HTTP Client",
		Description: "HTTP client with InsecureSkipVerify enabled",
		Severity:    "HIGH",
		Pattern:     regexp.MustCompile(`InsecureSkipVerify:\s*true`),
		FileTypes:   []string{".go"},
	})

	// Sensitive data in logs
	s.AddPattern(VulnerabilityPattern{
		ID:          "VULN-010",
		Name:        "Sensitive Data in Logs",
		Description: "Logging of potentially sensitive information",
		Severity:    "MEDIUM",
		Pattern:     regexp.MustCompile(`(?i)log\.(Print|Fatal|Panic).*\b(password|token|secret|key)\b`),
		FileTypes:   []string{".go", ".js", ".ts", ".py"},
	})

	// Unsafe reflection
	s.AddPattern(VulnerabilityPattern{
		ID:          "VULN-011",
		Name:        "Unsafe Reflection",
		Description: "Use of reflection with untrusted input",
		Severity:    "HIGH",
		Pattern:     regexp.MustCompile(`reflect\.(ValueOf|TypeOf)\([^)]*userInput[^)]*\)`),
		FileTypes:   []string{".go"},
	})

	// TODO/FIXME security issues
	s.AddPattern(VulnerabilityPattern{
		ID:          "VULN-012",
		Name:        "Security TODO",
		Description: "Security-related TODO or FIXME found",
		Severity:    "LOW",
		Pattern:     regexp.MustCompile(`(?i)(TODO|FIXME).*\b(security|vuln|hack|fix|auth)\b`),
		FileTypes:   []string{".go", ".js", ".ts", ".py", ".java"},
	})
}

// AddPattern adds a custom vulnerability pattern
func (s *VulnerabilityScanner) AddPattern(pattern VulnerabilityPattern) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.patterns = append(s.patterns, pattern)
}

// ScanDirectory scans a directory recursively for vulnerabilities
func (s *VulnerabilityScanner) ScanDirectory(rootPath string) (*ScanReport, error) {
	report := &ScanReport{
		Timestamp: time.Now(),
		Results:   make([]ScanResult, 0),
	}

	err := filepath.Walk(rootPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil // Skip files with errors
		}

		if info.IsDir() {
			// Skip common directories
			if info.Name() == ".git" || info.Name() == "node_modules" || info.Name() == "vendor" {
				return filepath.SkipDir
			}
			return nil
		}

		// Check if file type should be scanned
		if s.shouldScanFile(path) {
			results, err := s.scanFile(path)
			if err == nil {
				report.FilesScanned++
				report.Results = append(report.Results, results...)
			}
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	// Count severity levels
	for _, result := range report.Results {
		report.Vulnerabilities++
		switch result.Severity {
		case "CRITICAL":
			report.Critical++
		case "HIGH":
			report.High++
		case "MEDIUM":
			report.Medium++
		case "LOW":
			report.Low++
		}
	}

	// Log scan completion
	if s.auditLogger != nil {
		s.auditLogger.Log(AuditEvent{
			Level:    AuditLevelSecurity,
			Category: "security_scan",
			Action:   "vulnerability_scan",
			Success:  true,
			Message:  fmt.Sprintf("Security scan completed: %d vulnerabilities found", report.Vulnerabilities),
			Metadata: map[string]interface{}{
				"files_scanned": report.FilesScanned,
				"critical":      report.Critical,
				"high":          report.High,
				"medium":        report.Medium,
				"low":           report.Low,
			},
		})
	}

	return report, nil
}

// shouldScanFile checks if a file should be scanned based on its extension
func (s *VulnerabilityScanner) shouldScanFile(path string) bool {
	ext := filepath.Ext(path)
	for _, pattern := range s.patterns {
		for _, fileType := range pattern.FileTypes {
			if ext == fileType {
				return true
			}
		}
	}
	return false
}

// scanFile scans a single file for vulnerabilities
func (s *VulnerabilityScanner) scanFile(path string) ([]ScanResult, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var results []ScanResult
	scanner := bufio.NewScanner(file)
	lineNumber := 0

	for scanner.Scan() {
		lineNumber++
		line := scanner.Text()

		// Check each pattern
		for _, pattern := range s.patterns {
			// Check if file type matches
			if !s.fileTypeMatches(path, pattern.FileTypes) {
				continue
			}

			if pattern.Pattern.MatchString(line) {
				result := ScanResult{
					Timestamp:   time.Now(),
					FilePath:    path,
					LineNumber:  lineNumber,
					VulnID:      pattern.ID,
					VulnName:    pattern.Name,
					Severity:    pattern.Severity,
					Description: pattern.Description,
					CodeSnippet: strings.TrimSpace(line),
					Remediation: s.getRemediation(pattern.ID),
				}
				results = append(results, result)
			}
		}
	}

	return results, scanner.Err()
}

// fileTypeMatches checks if file extension matches pattern file types
func (s *VulnerabilityScanner) fileTypeMatches(path string, fileTypes []string) bool {
	ext := filepath.Ext(path)
	for _, fileType := range fileTypes {
		if ext == fileType {
			return true
		}
	}
	return false
}

// getRemediation returns remediation advice for a vulnerability
func (s *VulnerabilityScanner) getRemediation(vulnID string) string {
	remediations := map[string]string{
		"VULN-001": "Move API keys to environment variables or secure vault",
		"VULN-002": "Use bcrypt for password hashing, never hardcode passwords",
		"VULN-003": "Store JWT secrets in environment variables",
		"VULN-004": "Use parameterized queries or prepared statements",
		"VULN-005": "Use SHA-256 or SHA-3 instead of MD5/SHA1",
		"VULN-006": "Use crypto/rand instead of math/rand for security",
		"VULN-007": "Validate and sanitize all user input before command execution",
		"VULN-008": "Use filepath.Clean and validate paths",
		"VULN-009": "Enable proper TLS certificate validation",
		"VULN-010": "Sanitize logs, avoid logging sensitive data",
		"VULN-011": "Validate all input before using with reflection",
		"VULN-012": "Review and address security TODOs",
	}

	if remediation, exists := remediations[vulnID]; exists {
		return remediation
	}

	return "Review and fix the identified issue"
}

// ScanSecrets specifically scans for hardcoded secrets
func (s *VulnerabilityScanner) ScanSecrets(path string) ([]ScanResult, error) {
	var secretResults []ScanResult

	fullReport, err := s.ScanDirectory(path)
	if err != nil {
		return nil, err
	}

	// Filter only secret-related vulnerabilities
	for _, result := range fullReport.Results {
		if strings.HasPrefix(result.VulnID, "VULN-00") && result.VulnID <= "VULN-003" {
			secretResults = append(secretResults, result)
		}
	}

	return secretResults, nil
}

// GenerateReport generates a formatted scan report
func (s *VulnerabilityScanner) GenerateReport(report *ScanReport) string {
	var sb strings.Builder

	sb.WriteString("═══════════════════════════════════════════════════════════\n")
	sb.WriteString("           SECURITY VULNERABILITY SCAN REPORT             \n")
	sb.WriteString("═══════════════════════════════════════════════════════════\n\n")

	sb.WriteString(fmt.Sprintf("Scan Date: %s\n", report.Timestamp.Format(time.RFC3339)))
	sb.WriteString(fmt.Sprintf("Files Scanned: %d\n", report.FilesScanned))
	sb.WriteString(fmt.Sprintf("Total Vulnerabilities: %d\n\n", report.Vulnerabilities))

	sb.WriteString("Severity Breakdown:\n")
	sb.WriteString(fmt.Sprintf("  CRITICAL: %d\n", report.Critical))
	sb.WriteString(fmt.Sprintf("  HIGH:     %d\n", report.High))
	sb.WriteString(fmt.Sprintf("  MEDIUM:   %d\n", report.Medium))
	sb.WriteString(fmt.Sprintf("  LOW:      %d\n\n", report.Low))

	if len(report.Results) > 0 {
		sb.WriteString("Detailed Findings:\n")
		sb.WriteString("───────────────────────────────────────────────────────────\n")

		for i, result := range report.Results {
			sb.WriteString(fmt.Sprintf("\n%d. [%s] %s\n", i+1, result.Severity, result.VulnName))
			sb.WriteString(fmt.Sprintf("   File: %s:%d\n", result.FilePath, result.LineNumber))
			sb.WriteString(fmt.Sprintf("   Description: %s\n", result.Description))
			sb.WriteString(fmt.Sprintf("   Code: %s\n", result.CodeSnippet))
			sb.WriteString(fmt.Sprintf("   Remediation: %s\n", result.Remediation))
		}
	}

	sb.WriteString("\n═══════════════════════════════════════════════════════════\n")

	return sb.String()
}
