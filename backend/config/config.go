package config

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"

	"github.com/joho/godotenv"
)

// Config holds all application configuration
type Config struct {
	// Server
	Port        string
	Environment string

	// Database
	Database DatabaseConfig

	// Redis
	Redis RedisConfig

	// JWT
	JWT JWTConfig

	// Admin
	Admin AdminConfig

	// Default Account Settings
	DefaultAccount DefaultAccountConfig

	// Broker Settings
	Broker BrokerConfig

	// LP Settings
	LP LPConfig

	// CORS
	CORS CORSConfig

	// Encryption
	Encryption EncryptionConfig

	// FIX API Provisioning
	FIX FIXConfig

	// Compliance Settings
	Compliance ComplianceConfig
}

type FIXConfig struct {
	ProvisioningEnabled   bool
	ProvisioningStorePath string
	MasterPassword        string
}

type ComplianceConfig struct {
	Enabled              bool
	AuditRetentionYears  int
	ReportArchivePath    string
	AutoArchiveEnabled   bool
	TamperProofEnabled   bool
	AdminOnlyAccess      bool
	MiFIDIIEnabled       bool
	SECRule606Enabled    bool
}

type DatabaseConfig struct {
	Host     string
	Port     string
	Name     string
	User     string
	Password string
	SSLMode  string
}

type RedisConfig struct {
	Host     string
	Port     string
	Password string
}

type JWTConfig struct {
	Secret string
	Expiry string
}

type AdminConfig struct {
	Email       string
	IPWhitelist []string
	Password    string // Bcrypt hashed password
}

type DefaultAccountConfig struct {
	Balance  float64
	Leverage int
	Currency string
}

type BrokerConfig struct {
	Name              string
	DisplayName       string
	PriceFeedLP       string
	PriceFeedName     string
	ExecutionMode     string
	DefaultLeverage   int
	DefaultBalance    float64
	MarginMode        string
	MaxTicksPerSymbol int
}

type LPConfig struct {
	OandaAPIKey      string
	OandaAccountID   string
	BinanceAPIKey    string
	BinanceSecretKey string
}

type CORSConfig struct {
	AllowedOrigins []string
}

type EncryptionConfig struct {
	MasterKey string
}

// Load loads configuration from environment variables
func Load() (*Config, error) {
	// Try to load .env file (ignore error if not found)
	_ = godotenv.Load()

	cfg := &Config{
		Port:        getEnv("PORT", "7999"),
		Environment: getEnv("ENVIRONMENT", "development"),

		Database: DatabaseConfig{
			Host:     getEnv("DB_HOST", "localhost"),
			Port:     getEnv("DB_PORT", "5432"),
			Name:     getEnv("DB_NAME", "trading_engine"),
			User:     getEnv("DB_USER", "postgres"),
			Password: getEnv("DB_PASSWORD", ""),
			SSLMode:  getEnv("DB_SSL_MODE", "disable"),
		},

		Redis: RedisConfig{
			Host:     getEnv("REDIS_HOST", "localhost"),
			Port:     getEnv("REDIS_PORT", "6379"),
			Password: getEnv("REDIS_PASSWORD", ""),
		},

		JWT: JWTConfig{
			Secret: getEnv("JWT_SECRET", ""),
			Expiry: getEnv("JWT_EXPIRY", "24h"),
		},

		Admin: AdminConfig{
			Email:       getEnv("ADMIN_EMAIL", "admin@example.com"),
			IPWhitelist: getEnvAsSlice("ADMIN_IP_WHITELIST", []string{"127.0.0.1", "::1"}, ","),
			Password:    getEnv("ADMIN_PASSWORD_HASH", ""),
		},

		DefaultAccount: DefaultAccountConfig{
			Balance:  getEnvAsFloat("DEFAULT_ACCOUNT_BALANCE", 10000.0),
			Leverage: getEnvAsInt("DEFAULT_ACCOUNT_LEVERAGE", 100),
			Currency: getEnv("DEFAULT_ACCOUNT_CURRENCY", "USD"),
		},

		Broker: BrokerConfig{
			Name:              getEnv("BROKER_NAME", "RTX Trading"),
			DisplayName:       getEnv("BROKER_DISPLAY_NAME", "YoForex"),
			PriceFeedLP:       getEnv("PRICE_FEED_LP", "OANDA"),
			PriceFeedName:     getEnv("PRICE_FEED_NAME", "YoForex LP"),
			ExecutionMode:     getEnv("EXECUTION_MODE", "BBOOK"),
			DefaultLeverage:   getEnvAsInt("DEFAULT_LEVERAGE", 100),
			DefaultBalance:    getEnvAsFloat("DEFAULT_BALANCE", 5000.0),
			MarginMode:        getEnv("MARGIN_MODE", "HEDGING"),
			MaxTicksPerSymbol: getEnvAsInt("MAX_TICKS_PER_SYMBOL", 50000),
		},

		LP: LPConfig{
			OandaAPIKey:      getEnv("OANDA_API_KEY", ""),
			OandaAccountID:   getEnv("OANDA_ACCOUNT_ID", ""),
			BinanceAPIKey:    getEnv("BINANCE_API_KEY", ""),
			BinanceSecretKey: getEnv("BINANCE_SECRET_KEY", ""),
		},

		CORS: CORSConfig{
			AllowedOrigins: getEnvAsSlice("ALLOWED_ORIGINS", []string{"http://localhost:3000"}, ","),
		},

		Encryption: EncryptionConfig{
			MasterKey: getEnv("MASTER_ENCRYPTION_KEY", ""),
		},

		FIX: FIXConfig{
			ProvisioningEnabled:   getEnvAsBool("FIX_PROVISIONING_ENABLED", false),
			ProvisioningStorePath: getEnv("FIX_PROVISIONING_STORE_PATH", "./data/fix_credentials"),
			MasterPassword:        getEnv("FIX_MASTER_PASSWORD", ""),
		},

		Compliance: ComplianceConfig{
			Enabled:             getEnvAsBool("COMPLIANCE_ENABLED", true),
			AuditRetentionYears: getEnvAsInt("AUDIT_RETENTION_YEARS", 7),
			ReportArchivePath:   getEnv("COMPLIANCE_ARCHIVE_PATH", "./data/compliance_reports"),
			AutoArchiveEnabled:  getEnvAsBool("COMPLIANCE_AUTO_ARCHIVE", true),
			TamperProofEnabled:  getEnvAsBool("COMPLIANCE_TAMPER_PROOF", true),
			AdminOnlyAccess:     getEnvAsBool("COMPLIANCE_ADMIN_ONLY", true),
			MiFIDIIEnabled:      getEnvAsBool("COMPLIANCE_MIFID_II", true),
			SECRule606Enabled:   getEnvAsBool("COMPLIANCE_SEC_RULE_606", true),
		},
	}

	// Validate required fields
	if err := cfg.Validate(); err != nil {
		return nil, err
	}

	return cfg, nil
}

// Validate checks if required configuration is present
func (c *Config) Validate() error {
	if c.Environment == "production" {
		if c.JWT.Secret == "" {
			return fmt.Errorf("JWT_SECRET is required in production")
		}
		if c.Encryption.MasterKey == "" {
			return fmt.Errorf("MASTER_ENCRYPTION_KEY is required in production")
		}
		if c.Admin.Password == "" {
			log.Println("WARNING: ADMIN_PASSWORD_HASH not set - admin login will use default password")
		}
	}

	return nil
}

// Helper functions
func getEnv(key string, defaultVal string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultVal
}

func getEnvAsInt(key string, defaultVal int) int {
	valueStr := getEnv(key, "")
	if value, err := strconv.Atoi(valueStr); err == nil {
		return value
	}
	return defaultVal
}

func getEnvAsFloat(key string, defaultVal float64) float64 {
	valueStr := getEnv(key, "")
	if value, err := strconv.ParseFloat(valueStr, 64); err == nil {
		return value
	}
	return defaultVal
}

func getEnvAsSlice(key string, defaultVal []string, sep string) []string {
	valueStr := getEnv(key, "")
	if valueStr == "" {
		return defaultVal
	}
	return strings.Split(valueStr, sep)
}

func getEnvAsBool(key string, defaultVal bool) bool {
	valueStr := getEnv(key, "")
	if valueStr == "" {
		return defaultVal
	}
	value, err := strconv.ParseBool(valueStr)
	if err != nil {
		return defaultVal
	}
	return value
}
