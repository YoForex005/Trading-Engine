package auth

import (
	"strings"
	"sync"
	"testing"

	"github.com/epic1st/rtx/backend/internal/core"
	"golang.org/x/crypto/bcrypt"
)

// TestNewService tests service initialization
func TestNewService(t *testing.T) {
	engine := core.NewEngine()
	service := NewService(engine, "", "test-jwt-secret-for-testing-only")

	if service == nil {
		t.Fatal("NewService() returned nil")
	}

	if service.engine == nil {
		t.Error("engine not set")
	}

	if service.adminHash == nil {
		t.Error("adminHash not initialized")
	}

	// Verify admin hash can validate "password"
	err := bcrypt.CompareHashAndPassword(service.adminHash, []byte("password"))
	if err != nil {
		t.Error("Admin hash should validate 'password'")
	}
}

// TestAdminLogin tests admin authentication
func TestAdminLogin(t *testing.T) {
	tests := []struct {
		name        string
		username    string
		password    string
		wantErr     bool
		wantRole    string
		wantUserID  string
	}{
		{
			name:       "Valid admin login",
			username:   "admin",
			password:   "password",
			wantErr:    false,
			wantRole:   "ADMIN",
			wantUserID: "0",
		},
		{
			name:     "Invalid admin password",
			username: "admin",
			password: "wrongpassword",
			wantErr:  true,
		},
		{
			name:     "Empty admin password",
			username: "admin",
			password: "",
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			engine := core.NewEngine()
			service := NewService(engine, "", "test-jwt-secret-for-testing-only")

			token, user, err := service.Login(tt.username, tt.password)

			if (err != nil) != tt.wantErr {
				t.Errorf("Login() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.wantErr {
				if token != "" {
					t.Error("Token should be empty on error")
				}
				if user != nil {
					t.Error("User should be nil on error")
				}
				return
			}

			// Validate successful login
			if token == "" {
				t.Error("Token should not be empty")
			}

			if user == nil {
				t.Fatal("User should not be nil")
			}

			if user.ID != tt.wantUserID {
				t.Errorf("User ID = %s, want %s", user.ID, tt.wantUserID)
			}

			if user.Username != tt.username {
				t.Errorf("Username = %s, want %s", user.Username, tt.username)
			}

			if user.Role != tt.wantRole {
				t.Errorf("Role = %s, want %s", user.Role, tt.wantRole)
			}
		})
	}
}

// TestClientLogin tests client authentication
func TestClientLogin(t *testing.T) {
	engine := core.NewEngine()
	service := NewService(engine, "", "test-jwt-secret-for-testing-only")

	// Create test account with hashed password
	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("clientpass123"), bcrypt.DefaultCost)
	account := engine.CreateAccount("user1", "trader1", string(hashedPassword), false)

	tests := []struct {
		name       string
		username   string
		password   string
		wantErr    bool
		wantRole   string
		wantUserID string
	}{
		{
			name:       "Valid client login by account ID",
			username:   "1",
			password:   "clientpass123",
			wantErr:    false,
			wantRole:   "TRADER",
			wantUserID: "1",
		},
		{
			name:     "Invalid client password",
			username: "1",
			password: "wrongpass",
			wantErr:  true,
		},
		{
			name:     "Non-existent account ID",
			username: "999999",
			password: "anypass",
			wantErr:  true,
		},
		{
			name:     "Non-existent username",
			username: "nonexistent",
			password: "anypass",
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			token, user, err := service.Login(tt.username, tt.password)

			if (err != nil) != tt.wantErr {
				t.Errorf("Login() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.wantErr {
				if token != "" {
					t.Error("Token should be empty on error")
				}
				if user != nil {
					t.Error("User should be nil on error")
				}
				return
			}

			// Validate successful login
			if token == "" {
				t.Error("Token should not be empty")
			}

			if user == nil {
				t.Fatal("User should not be nil")
			}

			if user.ID != tt.wantUserID {
				t.Errorf("User ID = %s, want %s", user.ID, tt.wantUserID)
			}

			if user.Role != tt.wantRole {
				t.Errorf("Role = %s, want %s", user.Role, tt.wantRole)
			}

			// Verify username matches account
			if user.Username != account.Username {
				t.Errorf("Username = %s, want %s", user.Username, account.Username)
			}
		})
	}
}

// TestPlaintextPasswordAutoUpgrade tests automatic bcrypt upgrade
func TestPlaintextPasswordAutoUpgrade(t *testing.T) {
	engine := core.NewEngine()
	service := NewService(engine, "", "test-jwt-secret-for-testing-only")

	// Create account with plaintext password (legacy mode)
	account := engine.CreateAccount("user1", "trader1", "plainpass", false)

	// First login should succeed and upgrade password
	token, user, err := service.Login("1", "plainpass")

	if err != nil {
		t.Fatalf("Login() error = %v", err)
	}

	if token == "" || user == nil {
		t.Fatal("Login should succeed with plaintext password")
	}

	// Verify password was upgraded to bcrypt
	updatedAccount, _ := engine.GetAccount(account.ID)
	if !strings.HasPrefix(updatedAccount.Password, "$2") {
		t.Error("Password should be upgraded to bcrypt hash")
	}

	// Second login should still work with same password
	token2, user2, err := service.Login("1", "plainpass")

	if err != nil {
		t.Errorf("Login after upgrade error = %v", err)
	}

	if token2 == "" || user2 == nil {
		t.Error("Login should succeed after password upgrade")
	}
}

// TestJWTTokenGeneration tests JWT token creation
func TestJWTTokenGeneration(t *testing.T) {
	engine := core.NewEngine()
	service := NewService(engine, "", "test-jwt-secret-for-testing-only")

	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("testpass"), bcrypt.DefaultCost)
	engine.CreateAccount("user1", "trader1", string(hashedPassword), false)

	token, user, err := service.Login("1", "testpass")

	if err != nil {
		t.Fatalf("Login() error = %v", err)
	}

	// JWT tokens should have 3 parts separated by dots
	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		t.Errorf("JWT token parts = %d, want 3 (header.payload.signature)", len(parts))
	}

	// Verify token is not empty
	if len(token) < 20 {
		t.Errorf("Token length = %d, seems too short for valid JWT", len(token))
	}

	// Verify user data is returned
	if user.ID == "" {
		t.Error("User ID should not be empty")
	}
}

// TestLoginErrorConsistency tests error messages for security
func TestLoginErrorConsistency(t *testing.T) {
	engine := core.NewEngine()
	service := NewService(engine, "", "test-jwt-secret-for-testing-only")

	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("testpass"), bcrypt.DefaultCost)
	engine.CreateAccount("user1", "trader1", string(hashedPassword), false)

	// Test various failure scenarios
	tests := []struct {
		name     string
		username string
		password string
	}{
		{"Non-existent user", "999999", "anypass"},
		{"Wrong password", "1", "wrongpass"},
		{"Admin wrong password", "admin", "wrongadminpass"},
	}

	var errorMessages []string

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, _, err := service.Login(tt.username, tt.password)

			if err == nil {
				t.Error("Expected error, got nil")
				return
			}

			errorMessages = append(errorMessages, err.Error())
		})
	}

	// Verify all errors are the same (prevents user enumeration)
	firstError := errorMessages[0]
	for i, errMsg := range errorMessages {
		if errMsg != firstError {
			t.Errorf("Error message %d = '%s', expected consistent error message '%s'", i, errMsg, firstError)
		}
	}
}

// TestConcurrentLogins tests thread-safety
func TestConcurrentLogins(t *testing.T) {
	engine := core.NewEngine()
	service := NewService(engine, "", "test-jwt-secret-for-testing-only")

	// Create multiple accounts
	for i := 1; i <= 10; i++ {
		hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("testpass"), bcrypt.DefaultCost)
		engine.CreateAccount("user1", "trader1", string(hashedPassword), false)
	}

	var wg sync.WaitGroup
	successCount := 0
	var mu sync.Mutex

	// Concurrent valid logins
	for i := 1; i <= 10; i++ {
		wg.Add(1)
		go func(accountID int) {
			defer wg.Done()

			token, user, err := service.Login(string(rune('0'+accountID)), "testpass")

			mu.Lock()
			defer mu.Unlock()

			if err == nil && token != "" && user != nil {
				successCount++
			}
		}(i)
	}

	wg.Wait()

	if successCount != 10 {
		t.Errorf("Successful logins = %d, want 10", successCount)
	}
}

// TestConcurrentAdminLogins tests concurrent admin authentication
func TestConcurrentAdminLogins(t *testing.T) {
	engine := core.NewEngine()
	service := NewService(engine, "", "test-jwt-secret-for-testing-only")

	var wg sync.WaitGroup
	successCount := 0
	var mu sync.Mutex

	// Concurrent admin logins
	for i := 0; i < 20; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()

			token, user, err := service.Login("admin", "password")

			mu.Lock()
			defer mu.Unlock()

			if err == nil && token != "" && user != nil && user.Role == "ADMIN" {
				successCount++
			}
		}()
	}

	wg.Wait()

	if successCount != 20 {
		t.Errorf("Successful admin logins = %d, want 20", successCount)
	}
}

// TestPasswordHashSecurity tests bcrypt security
func TestPasswordHashSecurity(t *testing.T) {
	engine := core.NewEngine()
	svc := NewService(engine, "", "test-jwt-secret-for-testing-only")

	// Create account
	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("securepass123"), bcrypt.DefaultCost)
	account := engine.CreateAccount("user1", "trader1", string(hashedPassword), false)

	// Use svc variable
	_ = svc

	// Verify password is properly hashed
	retrievedAccount, _ := engine.GetAccount(account.ID)

	if retrievedAccount.Password == "securepass123" {
		t.Error("Password should be hashed, not stored as plaintext")
	}

	if !strings.HasPrefix(retrievedAccount.Password, "$2") {
		t.Error("Password should be bcrypt hash (starts with $2)")
	}

	// Verify hash validates correct password
	err := bcrypt.CompareHashAndPassword([]byte(retrievedAccount.Password), []byte("securepass123"))
	if err != nil {
		t.Error("Password hash should validate correct password")
	}

	// Verify hash rejects wrong password
	err = bcrypt.CompareHashAndPassword([]byte(retrievedAccount.Password), []byte("wrongpass"))
	if err == nil {
		t.Error("Password hash should reject wrong password")
	}
}

// TestLoginWithEmptyCredentials tests edge cases
func TestLoginWithEmptyCredentials(t *testing.T) {
	engine := core.NewEngine()
	service := NewService(engine, "", "test-jwt-secret-for-testing-only")

	tests := []struct {
		name     string
		username string
		password string
	}{
		{"Empty username and password", "", ""},
		{"Empty username", "", "somepass"},
		{"Empty password", "admin", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			token, user, err := service.Login(tt.username, tt.password)

			if err == nil {
				t.Error("Expected error for empty credentials, got nil")
			}

			if token != "" {
				t.Error("Token should be empty on error")
			}

			if user != nil {
				t.Error("User should be nil on error")
			}
		})
	}
}

// TestAccountIDParsing tests numeric ID parsing
func TestAccountIDParsing(t *testing.T) {
	engine := core.NewEngine()
	service := NewService(engine, "", "test-jwt-secret-for-testing-only")

	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("testpass"), bcrypt.DefaultCost)
	engine.CreateAccount("user1", "trader1", string(hashedPassword), false)

	tests := []struct {
		name     string
		username string
		wantErr  bool
	}{
		{"Valid numeric ID", "1", false},
		{"Invalid numeric format", "abc", true},
		{"Invalid numeric format with special chars", "12@34", true},
		{"Zero ID", "0", true}, // Account 0 doesn't exist (admin is special case)
		{"Negative ID", "-1", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, _, err := service.Login(tt.username, "testpass")

			if (err != nil) != tt.wantErr {
				t.Errorf("Login() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// TestMultipleAccountsPerUser tests user with multiple accounts
func TestMultipleAccountsPerUser(t *testing.T) {
	engine := core.NewEngine()
	service := NewService(engine, "", "test-jwt-secret-for-testing-only")

	// Create multiple accounts for same user
	hashedPass1, _ := bcrypt.GenerateFromPassword([]byte("pass1"), bcrypt.DefaultCost)
	hashedPass2, _ := bcrypt.GenerateFromPassword([]byte("pass2"), bcrypt.DefaultCost)

	account1 := engine.CreateAccount("user1", "trader1", string(hashedPass1), false)
	account2 := engine.CreateAccount("user1", "trader2", string(hashedPass2), true)

	// Login to first account
	token1, user1, err1 := service.Login("1", "pass1")
	if err1 != nil {
		t.Fatalf("Login to account 1 failed: %v", err1)
	}

	// Login to second account
	token2, user2, err2 := service.Login("2", "pass2")
	if err2 != nil {
		t.Fatalf("Login to account 2 failed: %v", err2)
	}

	// Verify different accounts
	if user1.ID == user2.ID {
		t.Error("Different accounts should have different IDs")
	}

	if token1 == token2 {
		t.Error("Different accounts should generate different tokens")
	}

	// Verify account-specific data
	if user1.ID != "1" {
		t.Errorf("Account 1 ID = %s, want 1", user1.ID)
	}

	if user2.ID != "2" {
		t.Errorf("Account 2 ID = %s, want 2", user2.ID)
	}

	// Both should be TRADER role
	if user1.Role != "TRADER" || user2.Role != "TRADER" {
		t.Error("Client accounts should have TRADER role")
	}

	// Verify usernames
	if user1.Username != account1.Username {
		t.Errorf("User1 username = %s, want %s", user1.Username, account1.Username)
	}

	if user2.Username != account2.Username {
		t.Errorf("User2 username = %s, want %s", user2.Username, account2.Username)
	}
}

// TestEdgeCases tests additional edge cases
func TestEdgeCases(t *testing.T) {
	t.Run("Login with very long password", func(t *testing.T) {
		engine := core.NewEngine()
		service := NewService(engine, "", "test-jwt-secret-for-testing-only")

		longPassword := strings.Repeat("a", 1000)
		hashedPassword, _ := bcrypt.GenerateFromPassword([]byte(longPassword), bcrypt.DefaultCost)
		engine.CreateAccount("user1", "trader1", string(hashedPassword), false)

		token, user, err := service.Login("1", longPassword)

		if err != nil {
			t.Errorf("Login with long password failed: %v", err)
		}

		if token == "" || user == nil {
			t.Error("Login should succeed with long password")
		}
	})

	t.Run("Login with special characters in password", func(t *testing.T) {
		engine := core.NewEngine()
		service := NewService(engine, "", "test-jwt-secret-for-testing-only")

		specialPassword := "P@$$w0rd!#%^&*()"
		hashedPassword, _ := bcrypt.GenerateFromPassword([]byte(specialPassword), bcrypt.DefaultCost)
		engine.CreateAccount("user1", "trader1", string(hashedPassword), false)

		token, user, err := service.Login("1", specialPassword)

		if err != nil {
			t.Errorf("Login with special chars failed: %v", err)
		}

		if token == "" || user == nil {
			t.Error("Login should succeed with special characters")
		}
	})

	t.Run("Admin login case sensitivity", func(t *testing.T) {
		engine := core.NewEngine()
		service := NewService(engine, "", "test-jwt-secret-for-testing-only")

		// Should fail with different case
		_, _, err := service.Login("Admin", "password")
		if err == nil {
			t.Error("Admin login should be case-sensitive")
		}

		_, _, err = service.Login("ADMIN", "password")
		if err == nil {
			t.Error("Admin login should be case-sensitive")
		}
	})
}
