package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"time"

	"backend/i18n"
	"backend/i18n/templates"
)

func main() {
	// Initialize i18n system
	if err := i18n.Init("en-US"); err != nil {
		log.Fatal(err)
	}

	// Example 1: Basic Translation
	fmt.Println("=== Basic Translation ===")
	englishText := i18n.T("en-US", "common.actions.save")
	spanishText := i18n.T("es-ES", "common.actions.save")
	fmt.Printf("English: %s\n", englishText)
	fmt.Printf("Spanish: %s\n", spanishText)

	// Example 2: Translation with Parameters
	fmt.Println("\n=== Translation with Parameters ===")
	minValue := 10
	message := i18n.T("en-US", "validation.min", minValue)
	fmt.Printf("Validation: %s\n", message)

	// Example 3: Number Formatting
	fmt.Println("\n=== Number Formatting ===")
	formatterUS := i18n.NewFormatter("en-US")
	formatterES := i18n.NewFormatter("es-ES")

	amount := 1234567.89

	fmt.Printf("US Currency: %s\n", formatterUS.FormatCurrency(amount, "USD"))
	fmt.Printf("ES Currency: %s\n", formatterES.FormatCurrency(amount, "EUR"))
	fmt.Printf("US Number: %s\n", formatterUS.FormatNumber(amount, 2))
	fmt.Printf("ES Number: %s\n", formatterES.FormatNumber(amount, 2))
	fmt.Printf("Percentage: %s\n", formatterUS.FormatPercentage(15.5, 2))
	fmt.Printf("Compact: %s\n", formatterUS.FormatCompactNumber(amount))

	// Example 4: Date/Time Formatting
	fmt.Println("\n=== Date/Time Formatting ===")
	now := time.Now()

	fmt.Printf("US Date: %s\n", formatterUS.FormatDate(now))
	fmt.Printf("ES Date: %s\n", formatterES.FormatDate(now))
	fmt.Printf("US Time: %s\n", formatterUS.FormatTime(now))
	fmt.Printf("ES Time: %s\n", formatterES.FormatTime(now))
	fmt.Printf("DateTime: %s\n", formatterUS.FormatDateTime(now))
	fmt.Printf("Long Date: %s\n", formatterUS.FormatLongDate(now))

	pastTime := now.Add(-2 * time.Hour)
	fmt.Printf("Relative: %s\n", formatterUS.FormatRelativeTime(pastTime))

	// Example 5: Email Templates
	fmt.Println("\n=== Email Templates ===")
	emailTemplates := templates.InitializeDefaultTemplates()

	emailData := map[string]interface{}{
		"UserName": "John Doe",
		"AppName":  "Trading Engine",
	}

	welcomeEmail, err := emailTemplates.Render("en-US", "welcome", emailData)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Subject: %s\n", welcomeEmail.Subject)
	fmt.Printf("Body Preview: %s...\n", welcomeEmail.Body[:100])

	// Spanish welcome email
	welcomeEmailES, _ := emailTemplates.Render("es-ES", "welcome", emailData)
	fmt.Printf("\nSpanish Subject: %s\n", welcomeEmailES.Subject)

	// Trade confirmation email
	tradeData := map[string]interface{}{
		"UserName":      "John Doe",
		"AppName":       "Trading Engine",
		"OrderID":       "ORD-12345",
		"Symbol":        "EURUSD",
		"Side":          "Buy",
		"Quantity":      "100,000",
		"Price":         "1.0950",
		"TotalValue":    "$109,500",
		"Status":        "Filled",
		"ExecutionTime": "2026-01-18 15:30:00 UTC",
	}

	tradeEmail, _ := emailTemplates.Render("en-US", "trade_confirmation", tradeData)
	fmt.Printf("\nTrade Email Subject: %s\n", tradeEmail.Subject)

	// Example 6: SMS Templates
	fmt.Println("\n=== SMS Templates ===")
	smsTemplates := templates.InitializeDefaultSMSTemplates()

	smsData := map[string]interface{}{
		"AppName":       "Trading Engine",
		"Code":          "123456",
		"ExpiryMinutes": 5,
	}

	smsMessage, err := smsTemplates.Render("en-US", "2fa_code", smsData)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("SMS: %s\n", smsMessage)

	// Spanish SMS
	smsMessageES, _ := smsTemplates.Render("es-ES", "2fa_code", smsData)
	fmt.Printf("SMS (Spanish): %s\n", smsMessageES)

	// Example 7: Context-based Translation
	fmt.Println("\n=== Context-based Translation ===")
	ctx := context.Background()
	ctx = i18n.WithLanguage(ctx, "es-ES")

	contextMessage := i18n.TranslateContext(ctx, "common.actions.cancel")
	fmt.Printf("From Context: %s\n", contextMessage)

	// Example 8: HTTP Middleware
	fmt.Println("\n=== HTTP Server Example ===")
	detector := i18n.NewLanguageDetector("en-US")

	http.HandleFunc("/api/test", func(w http.ResponseWriter, r *http.Request) {
		lang := i18n.FromContext(r.Context())
		message := i18n.T(lang, "common.app.name")

		formatter := i18n.FormatContext(r.Context())
		formattedPrice := formatter.FormatCurrency(1234.56, "USD")

		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintf(w, `{"message": "%s", "price": "%s"}`, message, formattedPrice)
	})

	http.Handle("/", detector.Middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Root endpoint"))
	})))

	fmt.Println("HTTP server would start on :8080")
	fmt.Println("Example endpoint: http://localhost:8080/api/test")
	fmt.Println("Try with: ?lang=es-ES or Accept-Language header")

	// Example 9: Error Handling with i18n
	fmt.Println("\n=== Error Handling ===")
	errorMessages := map[string]string{
		"en-US": i18n.T("en-US", "errors.trading.insufficientBalance"),
		"es-ES": i18n.T("es-ES", "errors.trading.insufficientBalance"),
		"fr-FR": i18n.T("fr-FR", "errors.trading.insufficientBalance"),
	}

	for lang, msg := range errorMessages {
		fmt.Printf("%s: %s\n", lang, msg)
	}

	// Example 10: RTL Language Check
	fmt.Println("\n=== RTL Detection ===")
	languages := []string{"en-US", "es-ES", "ar-SA", "fr-FR"}
	for _, lang := range languages {
		isRTL := i18n.IsRTL(lang)
		fmt.Printf("%s is RTL: %v\n", lang, isRTL)
	}

	// Example 11: File Size Formatting
	fmt.Println("\n=== File Size Formatting ===")
	sizes := []int64{1024, 1048576, 1073741824, 1099511627776}
	for _, size := range sizes {
		fmt.Printf("%d bytes = %s\n", size, formatterUS.FormatFileSize(size))
	}

	// Example 12: Compact Number Formatting
	fmt.Println("\n=== Compact Numbers ===")
	numbers := []float64{999, 1500, 1500000, 1500000000}
	for _, num := range numbers {
		fmt.Printf("%.0f = %s\n", num, formatterUS.FormatCompactNumber(num))
	}

	fmt.Println("\n=== Translation Validation ===")
	// Example 13: Validation
	validator := i18n.NewValidator("en-US", "./locales")
	result, err := validator.Validate()
	if err != nil {
		log.Printf("Validation error: %v\n", err)
	} else {
		fmt.Printf("Total keys: %d\n", result.TotalKeys)
		fmt.Printf("Valid: %v\n", result.IsValid)

		for lang, coverage := range result.CoveragePercent {
			fmt.Printf("%s coverage: %.2f%%\n", lang, coverage)
		}

		if len(result.MissingKeys) > 0 {
			fmt.Println("\nMissing translations:")
			for lang, keys := range result.MissingKeys {
				fmt.Printf("%s: %d missing keys\n", lang, len(keys))
			}
		}
	}
}

// Example HTTP Handler with full i18n support
func exampleHandler(w http.ResponseWriter, r *http.Request) {
	lang := i18n.FromContext(r.Context())
	formatter := i18n.FormatContext(r.Context())

	// Simulate a trading response
	response := map[string]interface{}{
		"message": i18n.T(lang, "trading.orders.orderPlaced"),
		"order": map[string]interface{}{
			"id":       "ORD-12345",
			"symbol":   "EURUSD",
			"side":     i18n.T(lang, "trading.orders.buy"),
			"quantity": 100000,
			"price":    formatter.FormatNumber(1.0950, 4),
			"total":    formatter.FormatCurrency(109500, "USD"),
			"time":     formatter.FormatDateTime(time.Now()),
		},
	}

	// In production, use proper JSON encoding
	fmt.Fprintf(w, "%+v\n", response)
}

// Example: Custom error with i18n
type TradingError struct {
	Code           string
	TranslationKey string
	Params         map[string]interface{}
}

func (e *TradingError) Error() string {
	return e.Code
}

func (e *TradingError) Localize(lang string) string {
	if len(e.Params) > 0 {
		// Convert params to interface{} slice for formatting
		params := make([]interface{}, 0, len(e.Params))
		for _, v := range e.Params {
			params = append(params, v)
		}
		return i18n.T(lang, e.TranslationKey, params...)
	}
	return i18n.T(lang, e.TranslationKey)
}

// Example usage:
// err := &TradingError{
//     Code: "INSUFFICIENT_BALANCE",
//     TranslationKey: "errors.trading.insufficientBalance",
// }
// localizedMessage := err.Localize("es-ES")
