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

// Zoho CRM API mock server
type MockZohoCRMServer struct {
	server *httptest.Server
	calls  []MockCall
}

func NewMockZohoCRMServer() *MockZohoCRMServer {
	m := &MockZohoCRMServer{
		calls: make([]MockCall, 0),
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
		case r.Method == "POST" && r.URL.Path == "/crm/v2/Leads":
			handleCreateLead(w)
		case r.Method == "GET" && r.URL.Path == "/crm/v2/Leads":
			handleListLeads(w)
		case r.Method == "GET" && len(r.URL.Path) > len("/crm/v2/Leads/"):
			handleGetLead(w)
		case r.Method == "PUT":
			handleUpdateLead(w)
		case r.Method == "DELETE":
			handleDeleteLead(w)
		default:
			http.Error(w, "Not Found", http.StatusNotFound)
		}
	}))

	return m
}

func (m *MockZohoCRMServer) Close() {
	m.server.Close()
}

func (m *MockZohoCRMServer) GetCalls() []MockCall {
	return m.calls
}

func (m *MockZohoCRMServer) Reset() {
	m.calls = make([]MockCall, 0)
}

// Handler functions
func handleCreateLead(w http.ResponseWriter) {
	response := map[string]interface{}{
		"data": []map[string]interface{}{
			{
				"code": 202,
				"details": map[string]interface{}{
					"id": "4556996000000057047",
				},
				"message": "record added",
				"status": "success",
			},
		},
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(response)
}

func handleListLeads(w http.ResponseWriter) {
	response := map[string]interface{}{
		"data": []map[string]interface{}{
			{
				"id": "4556996000000057047",
				"Last_Name": "Doe",
				"First_Name": "John",
				"Email": "john@example.com",
				"Company": "Acme",
				"Created_Time": time.Now().Format(time.RFC3339),
			},
		},
		"info": map[string]interface{}{
			"per_page": 200,
			"count": 1,
			"page": 1,
			"has_more": false,
		},
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func handleGetLead(w http.ResponseWriter) {
	response := map[string]interface{}{
		"data": []map[string]interface{}{
			{
				"id": "4556996000000057047",
				"Last_Name": "Doe",
				"First_Name": "John",
				"Email": "john@example.com",
				"Company": "Acme",
				"Created_Time": time.Now().Format(time.RFC3339),
				"Modified_Time": time.Now().Format(time.RFC3339),
			},
		},
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func handleUpdateLead(w http.ResponseWriter) {
	response := map[string]interface{}{
		"data": []map[string]interface{}{
			{
				"code": 204,
				"details": map[string]interface{}{
					"id": "4556996000000057047",
				},
				"message": "record updated",
				"status": "success",
			},
		},
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func handleDeleteLead(w http.ResponseWriter) {
	response := map[string]interface{}{
		"data": []map[string]interface{}{
			{
				"code": 204,
				"details": map[string]interface{}{
					"id": "4556996000000057047",
				},
				"message": "record deleted",
				"status": "success",
			},
		},
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// ZohoCRMClient for testing
type ZohoCRMClient struct {
	baseURL    string
	authToken  string
	orgID      string
	client     *http.Client
}

func NewZohoCRMClient(baseURL, authToken, orgID string) *ZohoCRMClient {
	return &ZohoCRMClient{
		baseURL:   baseURL,
		authToken: authToken,
		orgID:     orgID,
		client: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

func (c *ZohoCRMClient) CreateLead(ctx context.Context, firstName, lastName, email, company string) (string, error) {
	payload := map[string]interface{}{
		"data": []map[string]interface{}{
			{
				"First_Name": firstName,
				"Last_Name": lastName,
				"Email": email,
				"Company": company,
			},
		},
	}

	body, _ := json.Marshal(payload)
	req, err := http.NewRequestWithContext(ctx, "POST", c.baseURL+"/crm/v2/Leads",
		bytes.NewReader(body))
	if err != nil {
		return "", err
	}

	req.Header.Set("Authorization", "Zoho-oauthtoken "+c.authToken)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	var result map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&result)

	data := result["data"].([]interface{})
	if len(data) > 0 {
		details := data[0].(map[string]interface{})["details"].(map[string]interface{})
		if id, ok := details["id"]; ok {
			return id.(string), nil
		}
	}

	return "", nil
}

func (c *ZohoCRMClient) GetLead(ctx context.Context, id string) (map[string]interface{}, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", c.baseURL+"/crm/v2/Leads/"+id,
		nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "Zoho-oauthtoken "+c.authToken)

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&result)

	data := result["data"].([]interface{})
	if len(data) > 0 {
		return data[0].(map[string]interface{}), nil
	}

	return nil, nil
}

func (c *ZohoCRMClient) ListLeads(ctx context.Context) ([]map[string]interface{}, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", c.baseURL+"/crm/v2/Leads",
		nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "Zoho-oauthtoken "+c.authToken)

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&result)

	data := result["data"].([]interface{})
	var leads []map[string]interface{}
	for _, d := range data {
		leads = append(leads, d.(map[string]interface{}))
	}

	return leads, nil
}

func (c *ZohoCRMClient) UpdateLead(ctx context.Context, id string, data map[string]interface{}) error {
	payload := map[string]interface{}{
		"data": []map[string]interface{}{
			data,
		},
	}

	body, _ := json.Marshal(payload)
	req, err := http.NewRequestWithContext(ctx, "PUT", c.baseURL+"/crm/v2/Leads/"+id,
		bytes.NewReader(body))
	if err != nil {
		return err
	}

	req.Header.Set("Authorization", "Zoho-oauthtoken "+c.authToken)
	req.Header.Set("Content-Type", "application/json")

	_, err = c.client.Do(req)
	return err
}

func (c *ZohoCRMClient) DeleteLead(ctx context.Context, id string) error {
	req, err := http.NewRequestWithContext(ctx, "DELETE", c.baseURL+"/crm/v2/Leads/"+id,
		nil)
	if err != nil {
		return err
	}

	req.Header.Set("Authorization", "Zoho-oauthtoken "+c.authToken)

	_, err = c.client.Do(req)
	return err
}

// Test cases
func TestZohoCRMCreateLead(t *testing.T) {
	server := NewMockZohoCRMServer()
	defer server.Close()

	client := NewZohoCRMClient(server.server.URL, "test-auth-token", "org-id")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	id, err := client.CreateLead(ctx, "John", "Doe", "john@example.com", "Acme")
	if err != nil {
		t.Fatalf("Failed to create lead: %v", err)
	}

	if id != "4556996000000057047" {
		t.Errorf("Expected lead ID '4556996000000057047', got '%s'", id)
	}
}

func TestZohoCRMGetLead(t *testing.T) {
	server := NewMockZohoCRMServer()
	defer server.Close()

	client := NewZohoCRMClient(server.server.URL, "test-auth-token", "org-id")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	lead, err := client.GetLead(ctx, "4556996000000057047")
	if err != nil {
		t.Fatalf("Failed to get lead: %v", err)
	}

	if lead == nil {
		t.Error("Lead is nil")
		return
	}

	if lead["Last_Name"] != "Doe" {
		t.Errorf("Expected last name 'Doe', got '%v'", lead["Last_Name"])
	}
}

func TestZohoCRMListLeads(t *testing.T) {
	server := NewMockZohoCRMServer()
	defer server.Close()

	client := NewZohoCRMClient(server.server.URL, "test-auth-token", "org-id")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	leads, err := client.ListLeads(ctx)
	if err != nil {
		t.Fatalf("Failed to list leads: %v", err)
	}

	if len(leads) == 0 {
		t.Error("Expected at least one lead")
	}
}

func TestZohoCRMUpdateLead(t *testing.T) {
	server := NewMockZohoCRMServer()
	defer server.Close()

	client := NewZohoCRMClient(server.server.URL, "test-auth-token", "org-id")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	updateData := map[string]interface{}{
		"Email": "newemail@example.com",
		"Company": "NewCompany",
	}

	err := client.UpdateLead(ctx, "4556996000000057047", updateData)
	if err != nil {
		t.Fatalf("Failed to update lead: %v", err)
	}
}

func TestZohoCRMDeleteLead(t *testing.T) {
	server := NewMockZohoCRMServer()
	defer server.Close()

	client := NewZohoCRMClient(server.server.URL, "test-auth-token", "org-id")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err := client.DeleteLead(ctx, "4556996000000057047")
	if err != nil {
		t.Fatalf("Failed to delete lead: %v", err)
	}
}

func TestZohoCRMLeadValidation(t *testing.T) {
	tests := []struct {
		name      string
		firstName string
		lastName  string
		email     string
		company   string
		valid     bool
	}{
		{"Valid Lead", "John", "Doe", "john@example.com", "Acme", true},
		{"Missing FirstName", "", "Doe", "john@example.com", "Acme", false},
		{"Missing LastName", "John", "", "john@example.com", "Acme", false},
		{"Missing Email", "John", "Doe", "", "Acme", false},
		{"Invalid Email", "John", "Doe", "invalid-email", "Acme", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			valid := validateZohoLead(tt.firstName, tt.lastName, tt.email, tt.company)
			if valid != tt.valid {
				t.Errorf("Expected valid=%v, got valid=%v", tt.valid, valid)
			}
		})
	}
}

func TestZohoCRMPagination(t *testing.T) {
	server := NewMockZohoCRMServer()
	defer server.Close()

	client := NewZohoCRMClient(server.server.URL, "test-auth-token", "org-id")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	leads, err := client.ListLeads(ctx)
	if err != nil {
		t.Fatalf("Failed to list leads: %v", err)
	}

	t.Logf("Retrieved %d leads", len(leads))
}

func TestZohoCRMBulkOperations(t *testing.T) {
	server := NewMockZohoCRMServer()
	defer server.Close()

	client := NewZohoCRMClient(server.server.URL, "test-auth-token", "org-id")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Create multiple leads
	for i := 0; i < 5; i++ {
		client.CreateLead(ctx, "Test"+string(rune(i)), "User", "test"+string(rune(i))+"@example.com", "TestCorp")
	}

	calls := server.GetCalls()
	t.Logf("Made %d API calls", len(calls))
}

func TestZohoCRMErrorHandling(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"code": 400,
			"message": "Invalid request body",
			"status": "error",
		})
	}))
	defer server.Close()

	client := NewZohoCRMClient(server.URL, "test-auth-token", "org-id")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err := client.CreateLead(ctx, "", "", "", "")
	// Should handle error gracefully
	t.Logf("Handled error response: %v", err)
}

func TestZohoCRMRateLimit(t *testing.T) {
	callCount := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callCount++
		if callCount > 10 {
			w.WriteHeader(http.StatusTooManyRequests)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"code": 429,
				"message": "Rate limit exceeded",
			})
			return
		}
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"data": []map[string]interface{}{
				{
					"id": "test",
					"First_Name": "Test",
				},
			},
		})
	}))
	defer server.Close()

	client := NewZohoCRMClient(server.URL, "test-auth-token", "org-id")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Make multiple requests
	for i := 0; i < 15; i++ {
		client.ListLeads(ctx)
	}

	t.Logf("Total API calls made: %d", callCount)
}

// Helper function to validate Zoho lead data
func validateZohoLead(firstName, lastName, email, company string) bool {
	if firstName == "" || lastName == "" || email == "" {
		return false
	}

	// Simple email validation
	if !bytes.Contains([]byte(email), []byte("@")) {
		return false
	}

	return true
}
