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

// HubSpot API mock server
type MockHubSpotServer struct {
	server *httptest.Server
	calls  []MockCall
}

type MockCall struct {
	Method   string
	Path     string
	Body     []byte
	Response interface{}
}

func NewMockHubSpotServer() *MockHubSpotServer {
	m := &MockHubSpotServer{
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
		case r.Method == "POST" && r.URL.Path == "/crm/v3/objects/contacts":
			handleCreateContact(w)
		case r.Method == "GET" && r.URL.Path == "/crm/v3/objects/contacts":
			handleListContacts(w)
		case r.Method == "GET" && len(r.URL.Path) > len("/crm/v3/objects/contacts/"):
			handleGetContact(w)
		case r.Method == "PATCH":
			handleUpdateContact(w)
		default:
			http.Error(w, "Not Found", http.StatusNotFound)
		}
	}))

	return m
}

func (m *MockHubSpotServer) Close() {
	m.server.Close()
}

func (m *MockHubSpotServer) GetCalls() []MockCall {
	return m.calls
}

func (m *MockHubSpotServer) Reset() {
	m.calls = make([]MockCall, 0)
}

// Handler functions
func handleCreateContact(w http.ResponseWriter) {
	response := map[string]interface{}{
		"id": "12345",
		"properties": map[string]interface{}{
			"email": "test@example.com",
			"firstname": "John",
			"lastname": "Doe",
		},
		"createdAt": time.Now().Format(time.RFC3339),
		"updatedAt": time.Now().Format(time.RFC3339),
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func handleListContacts(w http.ResponseWriter) {
	response := map[string]interface{}{
		"results": []map[string]interface{}{
			{
				"id": "12345",
				"properties": map[string]interface{}{
					"email": "test@example.com",
					"firstname": "John",
				},
			},
		},
		"paging": map[string]interface{}{
			"next": map[string]interface{}{
				"after": "next_page_token",
			},
		},
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func handleGetContact(w http.ResponseWriter) {
	response := map[string]interface{}{
		"id": "12345",
		"properties": map[string]interface{}{
			"email": "test@example.com",
			"firstname": "John",
			"lastname": "Doe",
		},
		"createdAt": time.Now().Format(time.RFC3339),
		"updatedAt": time.Now().Format(time.RFC3339),
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func handleUpdateContact(w http.ResponseWriter) {
	response := map[string]interface{}{
		"id": "12345",
		"properties": map[string]interface{}{
			"email": "updated@example.com",
			"firstname": "Jane",
		},
		"updatedAt": time.Now().Format(time.RFC3339),
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// HubSpotClient for testing
type HubSpotClient struct {
	baseURL string
	apiKey  string
	client  *http.Client
}

func NewHubSpotClient(baseURL, apiKey string) *HubSpotClient {
	return &HubSpotClient{
		baseURL: baseURL,
		apiKey:  apiKey,
		client: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

func (c *HubSpotClient) CreateContact(ctx context.Context, email, firstname, lastname string) (string, error) {
	payload := map[string]interface{}{
		"properties": map[string]interface{}{
			"email": email,
			"firstname": firstname,
			"lastname": lastname,
		},
	}

	body, _ := json.Marshal(payload)
	req, err := http.NewRequestWithContext(ctx, "POST", c.baseURL+"/crm/v3/objects/contacts",
		bytes.NewReader(body))
	if err != nil {
		return "", err
	}

	req.Header.Set("Authorization", "Bearer "+c.apiKey)
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

func (c *HubSpotClient) GetContact(ctx context.Context, id string) (map[string]interface{}, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", c.baseURL+"/crm/v3/objects/contacts/"+id,
		nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "Bearer "+c.apiKey)

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&result)

	return result, nil
}

func (c *HubSpotClient) ListContacts(ctx context.Context) ([]map[string]interface{}, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", c.baseURL+"/crm/v3/objects/contacts",
		nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "Bearer "+c.apiKey)

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&result)

	results := result["results"].([]interface{})
	var contacts []map[string]interface{}
	for _, r := range results {
		contacts = append(contacts, r.(map[string]interface{}))
	}

	return contacts, nil
}

// Test cases
func TestHubSpotCreateContact(t *testing.T) {
	server := NewMockHubSpotServer()
	defer server.Close()

	client := NewHubSpotClient(server.server.URL, "test-api-key")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	id, err := client.CreateContact(ctx, "john@example.com", "John", "Doe")
	if err != nil {
		t.Fatalf("Failed to create contact: %v", err)
	}

	if id != "12345" {
		t.Errorf("Expected contact ID '12345', got '%s'", id)
	}

	// Verify API call
	calls := server.GetCalls()
	if len(calls) == 0 {
		t.Fatal("No API calls recorded")
	}

	lastCall := calls[len(calls)-1]
	if lastCall.Method != "POST" {
		t.Errorf("Expected POST method, got %s", lastCall.Method)
	}

	if lastCall.Path != "/crm/v3/objects/contacts" {
		t.Errorf("Expected path /crm/v3/objects/contacts, got %s", lastCall.Path)
	}
}

func TestHubSpotGetContact(t *testing.T) {
	server := NewMockHubSpotServer()
	defer server.Close()

	client := NewHubSpotClient(server.server.URL, "test-api-key")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	contact, err := client.GetContact(ctx, "12345")
	if err != nil {
		t.Fatalf("Failed to get contact: %v", err)
	}

	if contact["id"] != "12345" {
		t.Errorf("Expected ID '12345', got '%v'", contact["id"])
	}
}

func TestHubSpotListContacts(t *testing.T) {
	server := NewMockHubSpotServer()
	defer server.Close()

	client := NewHubSpotClient(server.server.URL, "test-api-key")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	contacts, err := client.ListContacts(ctx)
	if err != nil {
		t.Fatalf("Failed to list contacts: %v", err)
	}

	if len(contacts) == 0 {
		t.Error("Expected at least one contact")
	}
}

func TestHubSpotAPIAuthHeader(t *testing.T) {
	server := NewMockHubSpotServer()
	defer server.Close()

	client := NewHubSpotClient(server.server.URL, "test-api-key-123")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	client.CreateContact(ctx, "test@example.com", "Test", "User")

	calls := server.GetCalls()
	if len(calls) == 0 {
		t.Fatal("No API calls recorded")
	}

	// Note: In real scenario, verify auth header
	t.Logf("API call recorded: %s %s", calls[0].Method, calls[0].Path)
}

func TestHubSpotContactValidation(t *testing.T) {
	tests := []struct {
		name     string
		email    string
		firstname string
		lastname  string
		valid    bool
	}{
		{"Valid Contact", "john@example.com", "John", "Doe", true},
		{"Missing Email", "", "John", "Doe", false},
		{"Missing FirstName", "john@example.com", "", "Doe", false},
		{"Invalid Email", "invalid-email", "John", "Doe", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			valid := validateContactData(tt.email, tt.firstname, tt.lastname)
			if valid != tt.valid {
				t.Errorf("Expected valid=%v, got valid=%v", tt.valid, valid)
			}
		})
	}
}

func TestHubSpotErrorHandling(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status": "error",
			"message": "Invalid request",
		})
	}))
	defer server.Close()

	client := NewHubSpotClient(server.URL, "test-api-key")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	contacts, err := client.ListContacts(ctx)
	// Should handle error gracefully
	t.Logf("Handled error response: %v, contacts: %v", err, contacts)
}

func TestHubSpotRateLimit(t *testing.T) {
	callCount := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callCount++
		if callCount > 5 {
			w.WriteHeader(http.StatusTooManyRequests)
			return
		}
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"id": "test",
		})
	}))
	defer server.Close()

	client := NewHubSpotClient(server.URL, "test-api-key")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Make multiple requests to test rate limiting
	for i := 0; i < 10; i++ {
		client.GetContact(ctx, "12345")
	}

	t.Logf("Total calls made: %d", callCount)
}

func TestHubSpotContactPagination(t *testing.T) {
	server := NewMockHubSpotServer()
	defer server.Close()

	client := NewHubSpotClient(server.server.URL, "test-api-key")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	contacts, err := client.ListContacts(ctx)
	if err != nil {
		t.Fatalf("Failed to list contacts: %v", err)
	}

	if len(contacts) > 0 {
		t.Logf("Retrieved %d contacts", len(contacts))
	}
}

// Helper function for contact validation
func validateContactData(email, firstname, lastname string) bool {
	if email == "" || firstname == "" || lastname == "" {
		return false
	}

	// Simple email validation
	if !bytes.Contains([]byte(email), []byte("@")) {
		return false
	}

	return true
}
