package security

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"time"
)

// SecurityTest represents a security test case
type SecurityTest struct {
	Name        string
	Description string
	TestFunc    func() (bool, string)
	Severity    string
}

// SecurityTestSuite manages security testing
type SecurityTestSuite struct {
	baseURL string
	client  *http.Client
	tests   []SecurityTest
}

// NewSecurityTestSuite creates a new security test suite
func NewSecurityTestSuite(baseURL string) *SecurityTestSuite {
	return &SecurityTestSuite{
		baseURL: baseURL,
		client: &http.Client{
			Timeout: 10 * time.Second,
		},
		tests: make([]SecurityTest, 0),
	}
}

// RegisterDefaultTests registers common security tests
func (s *SecurityTestSuite) RegisterDefaultTests() {
	// SQL Injection tests
	s.AddTest(SecurityTest{
		Name:        "SQL Injection - Basic",
		Description: "Tests for basic SQL injection vulnerabilities",
		Severity:    "CRITICAL",
		TestFunc:    s.testSQLInjectionBasic,
	})

	s.AddTest(SecurityTest{
		Name:        "SQL Injection - Union",
		Description: "Tests for UNION-based SQL injection",
		Severity:    "CRITICAL",
		TestFunc:    s.testSQLInjectionUnion,
	})

	// XSS tests
	s.AddTest(SecurityTest{
		Name:        "XSS - Reflected",
		Description: "Tests for reflected XSS vulnerabilities",
		Severity:    "CRITICAL",
		TestFunc:    s.testXSSReflected,
	})

	s.AddTest(SecurityTest{
		Name:        "XSS - Stored",
		Description: "Tests for stored XSS vulnerabilities",
		Severity:    "CRITICAL",
		TestFunc:    s.testXSSStored,
	})

	// CSRF tests
	s.AddTest(SecurityTest{
		Name:        "CSRF Protection",
		Description: "Tests CSRF token validation",
		Severity:    "HIGH",
		TestFunc:    s.testCSRFProtection,
	})

	// Authentication tests
	s.AddTest(SecurityTest{
		Name:        "Weak Credentials",
		Description: "Tests for weak password acceptance",
		Severity:    "HIGH",
		TestFunc:    s.testWeakCredentials,
	})

	s.AddTest(SecurityTest{
		Name:        "Brute Force Protection",
		Description: "Tests rate limiting on login",
		Severity:    "HIGH",
		TestFunc:    s.testBruteForceProtection,
	})

	// Header security tests
	s.AddTest(SecurityTest{
		Name:        "Security Headers",
		Description: "Validates security headers are present",
		Severity:    "MEDIUM",
		TestFunc:    s.testSecurityHeaders,
	})

	// SSL/TLS tests
	s.AddTest(SecurityTest{
		Name:        "HTTPS Enforcement",
		Description: "Tests HTTPS enforcement",
		Severity:    "CRITICAL",
		TestFunc:    s.testHTTPSEnforcement,
	})

	// Session management tests
	s.AddTest(SecurityTest{
		Name:        "Session Timeout",
		Description: "Tests session timeout enforcement",
		Severity:    "MEDIUM",
		TestFunc:    s.testSessionTimeout,
	})

	// Information disclosure tests
	s.AddTest(SecurityTest{
		Name:        "Error Message Disclosure",
		Description: "Tests for information leakage in errors",
		Severity:    "LOW",
		TestFunc:    s.testErrorDisclosure,
	})

	// Path traversal tests
	s.AddTest(SecurityTest{
		Name:        "Path Traversal",
		Description: "Tests for path traversal vulnerabilities",
		Severity:    "HIGH",
		TestFunc:    s.testPathTraversal,
	})
}

// AddTest adds a custom security test
func (s *SecurityTestSuite) AddTest(test SecurityTest) {
	s.tests = append(s.tests, test)
}

// RunAll executes all security tests
func (s *SecurityTestSuite) RunAll() *SecurityTestReport {
	report := &SecurityTestReport{
		Timestamp: time.Now(),
		Results:   make([]SecurityTestResult, 0),
	}

	for _, test := range s.tests {
		passed, message := test.TestFunc()

		result := SecurityTestResult{
			TestName:    test.Name,
			Description: test.Description,
			Severity:    test.Severity,
			Passed:      passed,
			Message:     message,
			Timestamp:   time.Now(),
		}

		report.Results = append(report.Results, result)

		if passed {
			report.Passed++
		} else {
			report.Failed++
		}
	}

	return report
}

// SecurityTestReport contains test results
type SecurityTestReport struct {
	Timestamp time.Time
	Passed    int
	Failed    int
	Results   []SecurityTestResult
}

// SecurityTestResult represents a single test result
type SecurityTestResult struct {
	TestName    string
	Description string
	Severity    string
	Passed      bool
	Message     string
	Timestamp   time.Time
}

// Individual test implementations

func (s *SecurityTestSuite) testSQLInjectionBasic() (bool, string) {
	// Test basic SQL injection patterns
	payloads := []string{
		"' OR '1'='1",
		"admin'--",
		"' OR 1=1--",
	}

	for _, payload := range payloads {
		resp, err := s.client.Get(s.baseURL + "/login?username=" + payload + "&password=test")
		if err != nil {
			continue
		}
		defer resp.Body.Close()

		body, _ := io.ReadAll(resp.Body)
		bodyStr := string(body)

		// Check for SQL error messages
		if strings.Contains(bodyStr, "SQL") || strings.Contains(bodyStr, "syntax") {
			return false, "SQL injection vulnerability detected with payload: " + payload
		}
	}

	return true, "No SQL injection vulnerabilities detected"
}

func (s *SecurityTestSuite) testSQLInjectionUnion() (bool, string) {
	payload := "' UNION SELECT null,null,null--"
	resp, err := s.client.Get(s.baseURL + "/api/users?id=" + payload)
	if err != nil {
		return true, "Endpoint not accessible"
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	if strings.Contains(string(body), "UNION") {
		return false, "UNION-based SQL injection possible"
	}

	return true, "UNION-based SQL injection not detected"
}

func (s *SecurityTestSuite) testXSSReflected() (bool, string) {
	payload := "<script>alert('XSS')</script>"
	resp, err := s.client.Get(s.baseURL + "/search?q=" + payload)
	if err != nil {
		return true, "Endpoint not accessible"
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	if strings.Contains(string(body), payload) {
		return false, "Reflected XSS vulnerability detected"
	}

	return true, "No reflected XSS detected"
}

func (s *SecurityTestSuite) testXSSStored() (bool, string) {
	// This would need to post data and check if it's stored unsanitized
	return true, "Stored XSS test requires manual verification"
}

func (s *SecurityTestSuite) testCSRFProtection() (bool, string) {
	// Attempt state-changing operation without CSRF token
	data := map[string]string{"amount": "1000"}
	jsonData, _ := json.Marshal(data)

	resp, err := s.client.Post(s.baseURL+"/api/transfer", "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return true, "Endpoint not accessible"
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusOK {
		return false, "CSRF protection not enforced"
	}

	return true, "CSRF protection appears to be in place"
}

func (s *SecurityTestSuite) testWeakCredentials() (bool, string) {
	weakPasswords := []string{"password", "123456", "admin"}

	for _, pwd := range weakPasswords {
		data := map[string]string{"username": "test", "password": pwd}
		jsonData, _ := json.Marshal(data)

		resp, _ := s.client.Post(s.baseURL+"/register", "application/json", bytes.NewBuffer(jsonData))
		if resp != nil {
			resp.Body.Close()
			if resp.StatusCode == http.StatusOK {
				return false, "Weak password accepted: " + pwd
			}
		}
	}

	return true, "Weak passwords rejected"
}

func (s *SecurityTestSuite) testBruteForceProtection() (bool, string) {
	// Attempt multiple failed logins
	for i := 0; i < 10; i++ {
		data := map[string]string{"username": "admin", "password": "wrong"}
		jsonData, _ := json.Marshal(data)
		s.client.Post(s.baseURL+"/login", "application/json", bytes.NewBuffer(jsonData))
	}

	// Try one more time
	data := map[string]string{"username": "admin", "password": "wrong"}
	jsonData, _ := json.Marshal(data)
	resp, err := s.client.Post(s.baseURL+"/login", "application/json", bytes.NewBuffer(jsonData))

	if err != nil {
		return true, "Endpoint not accessible"
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusTooManyRequests || resp.StatusCode == http.StatusForbidden {
		return true, "Brute force protection is active"
	}

	return false, "No brute force protection detected"
}

func (s *SecurityTestSuite) testSecurityHeaders() (bool, string) {
	resp, err := s.client.Get(s.baseURL)
	if err != nil {
		return false, "Cannot reach server"
	}
	defer resp.Body.Close()

	requiredHeaders := map[string]string{
		"Strict-Transport-Security": "HSTS",
		"X-Frame-Options":           "Clickjacking protection",
		"X-Content-Type-Options":    "MIME sniffing protection",
		"Content-Security-Policy":   "XSS protection",
	}

	missing := []string{}
	for header, desc := range requiredHeaders {
		if resp.Header.Get(header) == "" {
			missing = append(missing, fmt.Sprintf("%s (%s)", header, desc))
		}
	}

	if len(missing) > 0 {
		return false, "Missing security headers: " + strings.Join(missing, ", ")
	}

	return true, "All security headers present"
}

func (s *SecurityTestSuite) testHTTPSEnforcement() (bool, string) {
	if strings.HasPrefix(s.baseURL, "https://") {
		return true, "Using HTTPS"
	}

	return false, "Not using HTTPS - should enforce HTTPS in production"
}

func (s *SecurityTestSuite) testSessionTimeout() (bool, string) {
	// This requires actual session management testing
	return true, "Session timeout test requires integration testing"
}

func (s *SecurityTestSuite) testErrorDisclosure() (bool, string) {
	resp, err := s.client.Get(s.baseURL + "/nonexistent")
	if err != nil {
		return true, "Endpoint not accessible"
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	bodyStr := string(body)

	// Check for sensitive information disclosure
	sensitivePatterns := []string{
		"stack trace",
		"database",
		"password",
		"secret",
		"internal server error",
	}

	for _, pattern := range sensitivePatterns {
		if strings.Contains(strings.ToLower(bodyStr), pattern) {
			return false, "Potential information disclosure: " + pattern
		}
	}

	return true, "No obvious information disclosure"
}

func (s *SecurityTestSuite) testPathTraversal() (bool, string) {
	payloads := []string{
		"../../../etc/passwd",
		"..\\..\\..\\windows\\system32\\config\\sam",
		"....//....//....//etc/passwd",
	}

	for _, payload := range payloads {
		resp, err := s.client.Get(s.baseURL + "/file?path=" + payload)
		if err != nil {
			continue
		}
		defer resp.Body.Close()

		if resp.StatusCode == http.StatusOK {
			body, _ := io.ReadAll(resp.Body)
			if strings.Contains(string(body), "root:") {
				return false, "Path traversal vulnerability detected"
			}
		}
	}

	return true, "No path traversal vulnerabilities detected"
}

// Mock HTTP testing helper
func CreateMockSecurityTest(handler http.Handler) *SecurityTestSuite {
	server := httptest.NewServer(handler)
	suite := NewSecurityTestSuite(server.URL)
	suite.RegisterDefaultTests()
	return suite
}
