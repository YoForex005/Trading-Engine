# Internationalization (i18n) Package

Comprehensive internationalization and localization support for the Trading Engine backend.

## Features

- **Multi-language Support**: 10 languages including RTL support for Arabic
- **Translation Management**: JSON-based translation files with namespaces
- **Auto-detection**: Language detection from headers, cookies, and query parameters
- **Formatting**: Locale-aware number, currency, date/time formatting
- **Email Templates**: Multi-language HTML and plain text email templates
- **SMS Templates**: Short message templates for notifications
- **Validation**: Translation completeness checking and coverage reports
- **Context Support**: Request context integration for language tracking

## Supported Languages

- English (en-US) - Primary
- Spanish (es-ES)
- French (fr-FR)
- German (de-DE)
- Japanese (ja-JP)
- Chinese Simplified (zh-CN)
- Arabic (ar-SA) - RTL
- Russian (ru-RU)
- Portuguese (pt-BR)
- Italian (it-IT)

## Usage

### Initialize i18n

```go
import "backend/i18n"

func main() {
    if err := i18n.Init("en-US"); err != nil {
        log.Fatal(err)
    }
}
```

### Translate Text

```go
// Simple translation
text := i18n.T("en-US", "common.actions.save")

// With parameters
text := i18n.T("en-US", "validation.min", 10)

// From context
text := i18n.TranslateContext(ctx, "errors.trading.insufficientBalance")
```

### Format Numbers and Dates

```go
formatter := i18n.NewFormatter("en-US")

// Currency
fmt.Println(formatter.FormatCurrency(1234.56, "USD")) // $1,234.56

// Date
fmt.Println(formatter.FormatDate(time.Now())) // 01/18/2026

// Percentage
fmt.Println(formatter.FormatPercentage(15.5, 2)) // 15.50%

// Compact number
fmt.Println(formatter.FormatCompactNumber(1500000)) // 1.5M
```

### Language Detection Middleware

```go
detector := i18n.NewLanguageDetector("en-US")

router.Use(detector.Middleware)
```

### Email Templates

```go
import "backend/i18n/templates"

emailTemplates := templates.InitializeDefaultTemplates()

data := map[string]interface{}{
    "UserName": "John Doe",
    "AppName": "Trading Engine",
}

email, err := emailTemplates.Render("en-US", "welcome", data)
if err != nil {
    log.Fatal(err)
}

// Send email with email.Subject, email.Body, email.BodyHTML
```

### SMS Templates

```go
smsTemplates := templates.InitializeDefaultSMSTemplates()

data := map[string]interface{}{
    "AppName": "Trading Engine",
    "Code": "123456",
    "ExpiryMinutes": 5,
}

message, err := smsTemplates.Render("en-US", "2fa_code", data)
if err != nil {
    log.Fatal(err)
}

// Send SMS with message
```

### Validate Translations

```go
validator := i18n.NewValidator("en-US", "./locales")

result, err := validator.Validate()
if err != nil {
    log.Fatal(err)
}

if !result.IsValid {
    for lang, keys := range result.MissingKeys {
        log.Printf("Language %s missing %d keys\n", lang, len(keys))
    }
}

// Check coverage
for lang, coverage := range result.CoveragePercent {
    log.Printf("Language %s: %.2f%% complete\n", lang, coverage)
}
```

## Translation File Structure

```
locales/
├── en-US/
│   ├── common.json
│   ├── trading.json
│   ├── errors.json
│   ├── notifications.json
│   └── legal.json
├── es-ES/
│   ├── common.json
│   └── ...
└── ... (other languages)
```

### Example Translation File

```json
{
  "orders": {
    "title": "Orders",
    "newOrder": "New Order",
    "buyOrder": "Buy Order"
  },
  "validation": {
    "required": "This field is required",
    "min": "Minimum value is {{.Min}}"
  }
}
```

## Translation Keys

Translation keys use dot notation: `namespace.section.key`

Examples:
- `common.actions.save`
- `trading.orders.title`
- `errors.network.timeout`
- `notifications.orders.orderPlaced`
- `legal.termsOfService.title`

## Best Practices

1. **Use Translation Keys**: Never hardcode user-facing text
2. **Namespace Organization**: Group related translations
3. **Context Matters**: Use appropriate namespaces (common, trading, errors, etc.)
4. **Parameterization**: Use template variables for dynamic content
5. **Validation**: Run validation before releases
6. **Fallback**: Always provide English translations
7. **Professional Translation**: Use professional translators for production
8. **Test Coverage**: Test with different locales
9. **RTL Support**: Test with Arabic for RTL layout

## Adding a New Language

1. Create directory: `locales/xx-YY/`
2. Copy English JSON files
3. Translate all keys
4. Add to `SupportedLanguages` in `i18n.go`
5. Add email templates
6. Add SMS templates
7. Run validation
8. Test thoroughly

## API Response Localization

```go
func HandleError(w http.ResponseWriter, r *http.Request, err error) {
    lang := i18n.FromContext(r.Context())

    errorMessage := i18n.T(lang, "errors.general.unknownError")
    if tradingErr, ok := err.(*TradingError); ok {
        errorMessage = i18n.T(lang, tradingErr.TranslationKey)
    }

    json.NewEncoder(w).Encode(map[string]string{
        "error": errorMessage,
    })
}
```

## Testing

```go
func TestTranslation(t *testing.T) {
    i18n.Init("en-US")

    text := i18n.T("en-US", "common.actions.save")
    assert.Equal(t, "Save", text)

    textES := i18n.T("es-ES", "common.actions.save")
    assert.Equal(t, "Guardar", textES)
}
```

## Performance

- Translation files are embedded and loaded at startup
- No file I/O during runtime
- Concurrent access via sync.RWMutex
- Cached printers for each language
- Minimal memory footprint

## License

Part of the Trading Engine backend.
