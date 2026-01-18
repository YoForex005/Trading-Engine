package payments

import (
	"context"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"io"
	"net"
	"regexp"
	"strings"
	"time"
)

// SecurityService handles PCI DSS compliant security
type SecurityService struct {
	encryptionKey []byte
	fraudRules    *FraudRules
	ipReputation  IPReputationChecker
	repository    Repository
}

// NewSecurityService creates a new security service
func NewSecurityService(
	encryptionKey string,
	fraudRules *FraudRules,
	ipReputation IPReputationChecker,
	repository Repository,
) *SecurityService {
	// Derive 32-byte key from encryption key
	hash := sha256.Sum256([]byte(encryptionKey))

	return &SecurityService{
		encryptionKey: hash[:],
		fraudRules:    fraudRules,
		ipReputation:  ipReputation,
		repository:    repository,
	}
}

// FraudDetector implements fraud detection interface
type FraudDetector interface {
	CheckDeposit(ctx context.Context, req *PaymentRequest) (*FraudCheck, error)
	CheckWithdrawal(ctx context.Context, req *PaymentRequest) (*FraudCheck, error)
}

// FraudRules defines fraud detection rules
type FraudRules struct {
	MaxVelocityPerHour   int     // Max transactions per hour
	MaxVelocityPerDay    int     // Max transactions per day
	MaxAmountPerDay      float64 // Max total amount per day
	HighRiskCountries    []string
	BlockedCountries     []string
	MinTimeBeforeRetry   time.Duration
	SuspiciousIPPatterns []string
}

// DefaultFraudRules returns default fraud detection rules
func DefaultFraudRules() *FraudRules {
	return &FraudRules{
		MaxVelocityPerHour: 5,
		MaxVelocityPerDay:  20,
		MaxAmountPerDay:    50000,
		HighRiskCountries:  []string{"NG", "CN", "RU", "IR", "KP"},
		BlockedCountries:   []string{"IR", "KP", "SY", "CU"},
		MinTimeBeforeRetry: 5 * time.Minute,
		SuspiciousIPPatterns: []string{
			"tor-exit",
			"proxy",
			"vpn",
		},
	}
}

// CheckDeposit performs fraud checks on deposit requests
func (s *SecurityService) CheckDeposit(ctx context.Context, req *PaymentRequest) (*FraudCheck, error) {
	check := &FraudCheck{
		TransactionID: "",
		RiskScore:     0,
		RiskLevel:     "low",
		Flags:         []string{},
		Blocked:       false,
		Checks:        make(map[string]string),
	}

	// Check 1: Velocity check
	velocityRisk, err := s.checkVelocity(ctx, req.UserID, req.Method)
	if err != nil {
		return nil, err
	}
	check.RiskScore += velocityRisk
	check.Checks["velocity"] = fmt.Sprintf("%.0f", velocityRisk)

	if velocityRisk > 0 {
		check.Flags = append(check.Flags, "high_velocity")
	}

	// Check 2: IP reputation
	ipRisk, err := s.checkIPReputation(ctx, req.IPAddress)
	if err == nil {
		check.RiskScore += ipRisk
		check.Checks["ip_reputation"] = fmt.Sprintf("%.0f", ipRisk)

		if ipRisk > 0 {
			check.Flags = append(check.Flags, "suspicious_ip")
		}
	}

	// Check 3: Geo-blocking
	// TODO: Implement IP geolocation to extract country from req.IPAddress
	// For now, skip geo-blocking check
	geoRisk := 0.0 // s.checkGeoLocation(country)
	check.RiskScore += geoRisk
	check.Checks["geo_location"] = fmt.Sprintf("%.0f", geoRisk)

	if geoRisk >= 100 {
		check.Blocked = true
		check.Reason = "blocked country"
		check.Flags = append(check.Flags, "blocked_country")
	} else if geoRisk > 0 {
		check.Flags = append(check.Flags, "high_risk_country")
	}

	// Check 4: Amount anomaly detection
	amountRisk, err := s.checkAmountAnomaly(ctx, req.UserID, req.Amount)
	if err == nil {
		check.RiskScore += amountRisk
		check.Checks["amount_anomaly"] = fmt.Sprintf("%.0f", amountRisk)

		if amountRisk > 0 {
			check.Flags = append(check.Flags, "unusual_amount")
		}
	}

	// Check 5: Device fingerprinting
	if req.DeviceID != "" {
		deviceRisk, err := s.checkDeviceReputation(ctx, req.DeviceID)
		if err == nil {
			check.RiskScore += deviceRisk
			check.Checks["device"] = fmt.Sprintf("%.0f", deviceRisk)

			if deviceRisk > 0 {
				check.Flags = append(check.Flags, "suspicious_device")
			}
		}
	}

	// Calculate final risk level
	if check.RiskScore >= 80 {
		check.RiskLevel = "critical"
		check.Blocked = true
		check.Reason = "high fraud risk score"
	} else if check.RiskScore >= 60 {
		check.RiskLevel = "high"
	} else if check.RiskScore >= 30 {
		check.RiskLevel = "medium"
	}

	return check, nil
}

// CheckWithdrawal performs fraud checks on withdrawal requests
func (s *SecurityService) CheckWithdrawal(ctx context.Context, req *PaymentRequest) (*FraudCheck, error) {
	check := &FraudCheck{
		TransactionID: "",
		RiskScore:     0,
		RiskLevel:     "low",
		Flags:         []string{},
		Blocked:       false,
		Checks:        make(map[string]string),
	}

	// Check 1: Account age
	ageRisk, err := s.checkAccountAge(ctx, req.UserID)
	if err == nil {
		check.RiskScore += ageRisk
		check.Checks["account_age"] = fmt.Sprintf("%.0f", ageRisk)

		if ageRisk > 0 {
			check.Flags = append(check.Flags, "new_account")
		}
	}

	// Check 2: Velocity check
	velocityRisk, err := s.checkVelocity(ctx, req.UserID, req.Method)
	if err == nil {
		check.RiskScore += velocityRisk
		check.Checks["velocity"] = fmt.Sprintf("%.0f", velocityRisk)

		if velocityRisk > 0 {
			check.Flags = append(check.Flags, "high_velocity")
		}
	}

	// Check 3: IP change detection
	ipChangeRisk, err := s.checkIPChange(ctx, req.UserID, req.IPAddress)
	if err == nil {
		check.RiskScore += ipChangeRisk
		check.Checks["ip_change"] = fmt.Sprintf("%.0f", ipChangeRisk)

		if ipChangeRisk > 0 {
			check.Flags = append(check.Flags, "ip_changed")
		}
	}

	// Check 4: Withdrawal pattern
	patternRisk, err := s.checkWithdrawalPattern(ctx, req.UserID, req.Amount)
	if err == nil {
		check.RiskScore += patternRisk
		check.Checks["pattern"] = fmt.Sprintf("%.0f", patternRisk)

		if patternRisk > 0 {
			check.Flags = append(check.Flags, "unusual_pattern")
		}
	}

	// Calculate final risk level
	if check.RiskScore >= 70 {
		check.RiskLevel = "critical"
	} else if check.RiskScore >= 50 {
		check.RiskLevel = "high"
	} else if check.RiskScore >= 25 {
		check.RiskLevel = "medium"
	}

	return check, nil
}

// TokenizeCard tokenizes card information (PCI DSS requirement)
func (s *SecurityService) TokenizeCard(cardNumber, cvv string) (string, error) {
	// Validate card number
	if !s.validateCardNumber(cardNumber) {
		return "", fmt.Errorf("invalid card number")
	}

	// Generate token
	token := fmt.Sprintf("tok_%d", time.Now().UnixNano())

	// Encrypt sensitive data
	encrypted, err := s.encrypt(cardNumber + "|" + cvv)
	if err != nil {
		return "", err
	}

	// Store encrypted data (not shown - would be in secure vault)
	// In production, use a PCI DSS compliant vault service
	_ = encrypted

	return token, nil
}

// DetokenizeCard retrieves card information from token
func (s *SecurityService) DetokenizeCard(token string) (string, error) {
	// Retrieve encrypted data from vault
	// For demonstration purposes, returning error
	return "", fmt.Errorf("tokenization vault not implemented")
}

// EncryptSensitiveData encrypts sensitive payment data
func (s *SecurityService) encrypt(plaintext string) (string, error) {
	block, err := aes.NewCipher(s.encryptionKey)
	if err != nil {
		return "", err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", err
	}

	ciphertext := gcm.Seal(nonce, nonce, []byte(plaintext), nil)
	return base64.StdEncoding.EncodeToString(ciphertext), nil
}

// DecryptSensitiveData decrypts sensitive payment data
func (s *SecurityService) decrypt(ciphertext string) (string, error) {
	data, err := base64.StdEncoding.DecodeString(ciphertext)
	if err != nil {
		return "", err
	}

	block, err := aes.NewCipher(s.encryptionKey)
	if err != nil {
		return "", err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	nonceSize := gcm.NonceSize()
	if len(data) < nonceSize {
		return "", fmt.Errorf("ciphertext too short")
	}

	nonce, ciphertextBytes := data[:nonceSize], data[nonceSize:]
	plaintext, err := gcm.Open(nil, nonce, ciphertextBytes, nil)
	if err != nil {
		return "", err
	}

	return string(plaintext), nil
}

// Helper functions

func (s *SecurityService) checkVelocity(ctx context.Context, userID string, _ PaymentMethod) (float64, error) {
	now := time.Now()
	hourAgo := now.Add(-1 * time.Hour)
	dayAgo := now.Add(-24 * time.Hour)

	// Check transactions in last hour
	hourCount, err := s.repository.CountUserTransactions(ctx, userID, hourAgo, now)
	if err != nil {
		return 0, err
	}

	// Check transactions in last day
	dayCount, err := s.repository.CountUserTransactions(ctx, userID, dayAgo, now)
	if err != nil {
		return 0, err
	}

	risk := 0.0

	if hourCount > s.fraudRules.MaxVelocityPerHour {
		risk += 40.0
	} else if hourCount > s.fraudRules.MaxVelocityPerHour/2 {
		risk += 20.0
	}

	if dayCount > s.fraudRules.MaxVelocityPerDay {
		risk += 30.0
	} else if dayCount > s.fraudRules.MaxVelocityPerDay/2 {
		risk += 15.0
	}

	return risk, nil
}

func (s *SecurityService) checkIPReputation(ctx context.Context, ipAddress string) (float64, error) {
	if s.ipReputation == nil {
		return 0, nil
	}

	reputation, err := s.ipReputation.CheckIP(ctx, ipAddress)
	if err != nil {
		return 0, err
	}

	risk := 0.0

	if reputation.IsTor {
		risk += 30.0
	}
	if reputation.IsProxy || reputation.IsVPN {
		risk += 20.0
	}
	if reputation.IsDataCenter {
		risk += 10.0
	}
	if reputation.ReputationScore < 50 {
		risk += 20.0
	}

	return risk, nil
}

func (s *SecurityService) checkGeoLocation(country string) float64 {
	// Check blocked countries
	for _, blocked := range s.fraudRules.BlockedCountries {
		if country == blocked {
			return 100.0
		}
	}

	// Check high-risk countries
	for _, highRisk := range s.fraudRules.HighRiskCountries {
		if country == highRisk {
			return 30.0
		}
	}

	return 0.0
}

func (s *SecurityService) checkAmountAnomaly(ctx context.Context, userID string, amount float64) (float64, error) {
	// Get user's average transaction amount
	avgAmount, err := s.repository.GetUserAverageTransactionAmount(ctx, userID)
	if err != nil {
		return 0, err
	}

	if avgAmount == 0 {
		return 0, nil
	}

	// Check if amount is significantly higher than average
	ratio := amount / avgAmount

	risk := 0.0
	if ratio > 10 {
		risk = 30.0
	} else if ratio > 5 {
		risk = 20.0
	} else if ratio > 3 {
		risk = 10.0
	}

	return risk, nil
}

func (s *SecurityService) checkDeviceReputation(ctx context.Context, deviceID string) (float64, error) {
	// Check if device has been used for failed transactions
	failedCount, err := s.repository.GetDeviceFailedTransactionCount(ctx, deviceID)
	if err != nil {
		return 0, err
	}

	risk := 0.0
	if failedCount > 5 {
		risk = 40.0
	} else if failedCount > 2 {
		risk = 20.0
	}

	return risk, nil
}

func (s *SecurityService) checkAccountAge(ctx context.Context, userID string) (float64, error) {
	createdAt, err := s.repository.GetUserCreatedAt(ctx, userID)
	if err != nil {
		return 0, err
	}

	age := time.Since(createdAt)

	risk := 0.0
	if age < 24*time.Hour {
		risk = 40.0
	} else if age < 7*24*time.Hour {
		risk = 20.0
	} else if age < 30*24*time.Hour {
		risk = 10.0
	}

	return risk, nil
}

func (s *SecurityService) checkIPChange(ctx context.Context, userID string, ipAddress string) (float64, error) {
	lastIP, err := s.repository.GetUserLastIP(ctx, userID)
	if err != nil {
		return 0, err
	}

	if lastIP == "" || lastIP == ipAddress {
		return 0, nil
	}

	// Check if IPs are from different countries
	// For demonstration, checking if IP subnets are different
	if !s.sameSubnet(lastIP, ipAddress) {
		return 25.0, nil
	}

	return 0, nil
}

func (s *SecurityService) checkWithdrawalPattern(ctx context.Context, userID string, amount float64) (float64, error) {
	// Check deposit-to-withdrawal ratio
	totalDeposits, err := s.repository.GetUserTotalDeposits(ctx, userID)
	if err != nil {
		return 0, err
	}

	if totalDeposits == 0 {
		return 30.0, nil
	}

	totalWithdrawals, err := s.repository.GetUserTotalWithdrawals(ctx, userID)
	if err != nil {
		return 0, err
	}

	ratio := (totalWithdrawals + amount) / totalDeposits

	risk := 0.0
	if ratio > 1.0 {
		risk = 20.0
	} else if ratio > 0.9 {
		risk = 10.0
	}

	return risk, nil
}

func (s *SecurityService) validateCardNumber(cardNumber string) bool {
	// Remove spaces and dashes
	cardNumber = strings.ReplaceAll(cardNumber, " ", "")
	cardNumber = strings.ReplaceAll(cardNumber, "-", "")

	// Check if only digits
	matched, _ := regexp.MatchString("^[0-9]+$", cardNumber)
	if !matched {
		return false
	}

	// Luhn algorithm
	sum := 0
	alternate := false

	for i := len(cardNumber) - 1; i >= 0; i-- {
		digit := int(cardNumber[i] - '0')

		if alternate {
			digit *= 2
			if digit > 9 {
				digit -= 9
			}
		}

		sum += digit
		alternate = !alternate
	}

	return sum%10 == 0
}

func (s *SecurityService) sameSubnet(ip1, ip2 string) bool {
	ip1Addr := net.ParseIP(ip1)
	ip2Addr := net.ParseIP(ip2)

	if ip1Addr == nil || ip2Addr == nil {
		return false
	}

	// Check /24 subnet
	mask := net.CIDRMask(24, 32)
	return ip1Addr.Mask(mask).Equal(ip2Addr.Mask(mask))
}

// IPReputationChecker checks IP reputation
type IPReputationChecker interface {
	CheckIP(ctx context.Context, ip string) (*IPReputation, error)
}

// IPReputation represents IP reputation data
type IPReputation struct {
	IP              string  `json:"ip"`
	IsTor           bool    `json:"is_tor"`
	IsProxy         bool    `json:"is_proxy"`
	IsVPN           bool    `json:"is_vpn"`
	IsDataCenter    bool    `json:"is_datacenter"`
	ReputationScore float64 `json:"reputation_score"`
	Country         string  `json:"country"`
	City            string  `json:"city"`
}

// LimitsChecker checks transaction limits
type LimitsChecker interface {
	CheckDepositLimits(ctx context.Context, userID string, method PaymentMethod, amount float64) error
	CheckWithdrawalLimits(ctx context.Context, userID string, method PaymentMethod, amount float64) error
	GetLimits(ctx context.Context, userID string, method PaymentMethod) (*PaymentLimits, error)
}
