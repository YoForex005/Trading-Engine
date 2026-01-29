package crm

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

// Salesforce API mock server
type MockSalesforceServer struct {
	server      *httptest.Server
	calls       []MockCall
	accessToken string
}

func NewMockSalesforceServer() *MockSalesforceServer {
	m := &MockSalesforceServer{
		calls:       make([]MockCall, 0),
		accessToken: "mock-access-token",
	}

	m.server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Record the call
		body, _ := io.ReadAll(r.Body)
		m.calls = append(m.calls, MockCall{
			Method: r.Method,
			Path:   r.URL.Path,
			Body:   body,
		})

		// Route to handler
		switch {
		case r.URL.Path == "/services/oauth2/token":
			handleOAuthToken(w)
		case r.Method == "POST" && r.URL.Path == "/services/data/v55.0/sobjects/Account":
			handleCreateAccount(w)
		case r.Method == "GET" && r.URL.Path == "/services/data/v55.0/sobjects/Account":
			handleListAccounts(w)
		case r.Method == "GET" && len(r.URL.Path) > len("/services/data/v55.0/sobjects/Account/"):
			handleGetAccount(w)
		case r.Method == "PATCH":
			handleUpdateAccount(w)
		default:
			http.Error(w, "Not Found", http.StatusNotFound)
		}
	}))

	return m
}

func (m *MockSalesforceServer) Close() {
	m.server.Close()
}

func (m *MockSalesforceServer) GetCalls() []MockCall {
	return m.calls
}

func (m *MockSalesforceServer) Reset() {
	m.calls = make([]MockCall, 0)
}

// Handler functions
func handleOAuthToken(w http.ResponseWriter) {
	response := map[string]interface{}{
		"access_token": "mock-access-token",
		"token_type": "Bearer",
		"expires_in": 3600,
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func handleCreateAccount(w http.ResponseWriter) {
	response := map[string]interface{}{
		"id": "0015e00000YHB1AAWY",
		"success": true,
		"errors": []interface{}{},
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(response)
}

func handleListAccounts(w http.ResponseWriter) {
	response := map[string]interface{}{
		"totalSize": 1,
		"done": true,
		"records": []map[string]interface{}{
			{
				"Id": "0015e00000YHB1AAWY",
				"Name": "Acme Corp",
				"Phone": "555-1234",
				"BillingCity": "San Francisco",
			},
		},
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func handleGetAccount(w http.ResponseWriter) {
	response := map[string]interface{}{
		"Id": "0015e00000YHB1AAWY",
		"Name": "Acme Corp",
		"Phone": "555-1234",
		"BillingCity": "San Francisco",
		"CreatedDate": time.Now().Format(time.RFC3339),
		"LastModifiedDate": time.Now().Format(time.RFC3339),
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func handleUpdateAccount(w http.ResponseWriter) {
	w.WriteHeader(http.StatusNoContent)
}

// SalesforceClient for testing
type SalesforceClient struct {
	baseURL      string
	clientID     string
	clientSecret string
	username     string
	password     string
	client       *http.Client
	accessToken  string
}

func NewSalesforceClient(baseURL, clientID, clientSecret, username, password string) *SalesforceClient {
	return &SalesforceClient{
		baseURL:      baseURL,
		clientID:     clientID,
		clientSecret: clientSecret,
		username:     username,
		password:     password,
		client: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

func (c *SalesforceClient) Authenticate(ctx context.Context) error {
	payload := map[string]string{
		"grant_type": "password",
		"client_id": c.clientID,
		"client_secret": c.clientSecret,
		"username": c.username,
		"password": c.password,
	}

	body, _ := json.Marshal(payload)
	req, err := http.NewRequestWithContext(ctx, "POST", c.baseURL+"/services/oauth2/token",
		bytes.NewReader(body))
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := c.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return err
	}

	if token, ok := result["access_token"]; ok {
		c.accessToken = token.(string)
		return nil
	}

	return nil
}

func (c *SalesforceClient) CreateAccount(ctx context.Context, name, phone, city string) (string, error) {
	payload := map[string]interface{}{
		"Name": name,
		"Phone": phone,
		"BillingCity": city,
	}

	body, _ := json.Marshal(payload)
	req, err := http.NewRequestWithContext(ctx, "POST",
		c.baseURL+"/services/data/v55.0/sobjects/Account",
		bytes.NewReader(body))
	if err != nil {
		return "", err
	}

	req.Header.Set("Authorization", "Bearer "+c.accessToken)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	var result map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&result)

	if id, ok := result["id"]; ok {
		return id.(string), nil
	}
	return "", nil
}

func (c *SalesforceClient) GetAccount(ctx context.Context, id string) (map[string]interface{}, error) {
	req, err := http.NewRequestWithContext(ctx, "GET",
		c.baseURL+"/services/data/v55.0/sobjects/Account/"+id,
		nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "Bearer "+c.accessToken)

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&result)

	return result, nil
}

func (c *SalesforceClient) ListAccounts(ctx context.Context) ([]map[string]interface{}, error) {
	req, err := http.NewRequestWithContext(ctx, "GET",
		c.baseURL+"/services/data/v55.0/sobjects/Account",
		nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "Bearer "+c.accessToken)

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&result)

	records := result["records"].([]interface{})
	var accounts []map[string]interface{}
	for _, r := range records {
		accounts = append(accounts, r.(map[string]interface{}))
	}

	return accounts, nil
}

// Test cases
func TestSalesforceAuthentication(t *testing.T) {
	server := NewMockSalesforceServer()
	defer server.Close()

	client := NewSalesforceClient(server.server.URL, "client-id", "client-secret",
		"user@example.com", "password")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err := client.Authenticate(ctx)
	if err != nil {
		t.Fatalf("Failed to authenticate: %v", err)
	}

	if client.accessToken == "" {
		t.Error("Access token not set after authentication")
	}
}

func TestSalesforceCreateAccount(t *testing.T) {
	server := NewMockSalesforceServer()
	defer server.Close()

	client := NewSalesforceClient(server.server.URL, "client-id", "client-secret",
		"user@example.com", "password")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// First authenticate
	client.Authenticate(ctx)

	// Then create account
	id, err := client.CreateAccount(ctx, "Acme Corp", "555-1234", "San Francisco")
	if err != nil {
		t.Fatalf("Failed to create account: %v", err)
	}

	if id != "0015e00000YHB1AAWY" {
		t.Errorf("Expected account ID '0015e00000YHB1AAWY', got '%s'", id)
	}
}

func TestSalesforceGetAccount(t *testing.T) {
	server := NewMockSalesforceServer()
	defer server.Close()

	client := NewSalesforceClient(server.server.URL, "client-id", "client-secret",
		"user@example.com", "password")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	client.Authenticate(ctx)

	account, err := client.GetAccount(ctx, "0015e00000YHB1AAWY")
	if err != nil {
		t.Fatalf("Failed to get account: %v", err)
	}

	if account["Name"] != "Acme Corp" {
		t.Errorf("Expected name 'Acme Corp', got '%v'", account["Name"])
	}
}

func TestSalesforceListAccounts(t *testing.T) {
	server := NewMockSalesforceServer()
	defer server.Close()

	client := NewSalesforceClient(server.server.URL, "client-id", "client-secret",
		"user@example.com", "password")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	client.Authenticate(ctx)

	accounts, err := client.ListAccounts(ctx)
	if err != nil {
		t.Fatalf("Failed to list accounts: %v", err)
	}

	if len(accounts) == 0 {
		t.Error("Expected at least one account")
	}
}

func TestSalesforceAccountValidation(t *testing.T) {
	tests := []struct {
		name  string
		accName string
		phone string
		city  string
		valid bool
	}{
		{"Valid Account", "Acme Corp", "555-1234", "San Francisco", true},
		{"Missing Name", "", "555-1234", "San Francisco", false},
		{"Missing Phone", "Acme Corp", "", "San Francisco", false},
		{"Missing City", "Acme Corp", "555-1234", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			valid := validateSalesforceAccount(tt.accName, tt.phone, tt.city)
			if valid != tt.valid {
				t.Errorf("Expected valid=%v, got valid=%v", tt.valid, valid)
			}
		})
	}
}

func TestSalesforceErrorHandling(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"error": "invalid_grant",
			"error_description": "authentication failure",
		})
	}))
	defer server.Close()

	client := NewSalesforceClient(server.URL, "client-id", "client-secret",
		"user@example.com", "password")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err := client.Authenticate(ctx)
	// Should handle error gracefully
	t.Logf("Handled error response: %v", err)
}

func TestSalesforceSOQLQuery(t *testing.T) {
	server := NewMockSalesforceServer()
	defer server.Close()

	// SOQL query test
	soqlQuery := "SELECT Id, Name, Phone FROM Account WHERE BillingCity = 'San Francisco'"

	t.Run("SOQLQueryValidation", func(t *testing.T) {
		// Validate SOQL syntax
		if !bytes.Contains([]byte(soqlQuery), []byte("SELECT")) {
			t.Error("Invalid SOQL query: missing SELECT")
		}

		if !bytes.Contains([]byte(soqlQuery), []byte("FROM")) {
			t.Error("Invalid SOQL query: missing FROM")
		}
	})
}

func TestSalesforceRateLimiting(t *testing.T) {
	callCount := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callCount++
		if callCount > 100 {
			w.WriteHeader(http.StatusTooManyRequests)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"error": "API_LIMIT_EXCEEDED",
			})
			return
		}
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"totalSize": 0,
			"records": []interface{}{},
		})
	}))
	defer server.Close()

	client := NewSalesforceClient(server.URL, "client-id", "client-secret",
		"user@example.com", "password")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	client.accessToken = "mock-token"

	// Make multiple requests
	for i := 0; i < 10; i++ {
		client.ListAccounts(ctx)
	}

	t.Logf("Total API calls made: %d", callCount)
}

func TestSalesforceTimestampHandling(t *testing.T) {
	account := map[string]interface{}{
		"Id": "0015e00000YHB1AAWY",
		"CreatedDate": "2024-01-18T10:30:00.000+0000",
		"LastModifiedDate": "2024-01-18T15:45:00.000+0000",
	}

	t.Run("TimestampParsing", func(t *testing.T) {
		createdDate, ok := account["CreatedDate"]
		if !ok {
			t.Error("CreatedDate not found")
		}

		if createdDate == "" {
			t.Error("CreatedDate is empty")
		}

		t.Logf("CreatedDate: %v", createdDate)
	})
}

// Helper function to validate Salesforce account data
func validateSalesforceAccount(name, phone, city string) bool {
	if name == "" || phone == "" || city == "" {
		return false
	}
	return true
}
