# Internationalization (i18n) Implementation Guide

Complete internationalization and localization support for the Trading Engine platform.

## Overview

The i18n system provides comprehensive multi-language support across frontend and backend:

- **10 Languages**: English, Spanish, French, German, Japanese, Chinese, Arabic, Russian, Portuguese, Italian
- **RTL Support**: Full right-to-left layout support for Arabic
- **Smart Detection**: Automatic language detection from headers, cookies, and user preferences
- **Professional Formatting**: Locale-aware numbers, currencies, dates, and times
- **Template System**: Multi-language email and SMS templates
- **Validation Tools**: Translation completeness checking and coverage reports

## Architecture

### Frontend (React + i18next)

```
clients/admin-dashboard/src/
├── i18n/
│   ├── config.ts          # i18next configuration
│   ├── hooks.ts           # Custom React hooks
│   ├── formatters.ts      # Number/date formatting
│   ├── pseudo.ts          # Testing utilities
│   └── index.ts           # Module exports
└── components/
    └── LanguageSelector.tsx
```

### Backend (Go)

```
backend/i18n/
├── i18n.go               # Core translator
├── formatter.go          # Locale formatting
├── middleware.go         # HTTP middleware
├── context.go            # Context integration
├── validator.go          # Translation validation
├── templates/
│   ├── email_templates.go
│   └── sms_templates.go
└── README.md
```

### Translation Files

```
locales/
├── en-US/               # English (Primary)
│   ├── common.json
│   ├── trading.json
│   ├── errors.json
│   ├── notifications.json
│   └── legal.json
├── es-ES/               # Spanish
├── fr-FR/               # French
├── de-DE/               # German
├── ja-JP/               # Japanese
├── zh-CN/               # Chinese (Simplified)
├── ar-SA/               # Arabic (RTL)
├── ru-RU/               # Russian
├── pt-BR/               # Portuguese (Brazil)
└── it-IT/               # Italian
```

## Quick Start

### Frontend Setup

1. **Install Dependencies**

```bash
cd clients/admin-dashboard
npm install i18next react-i18next i18next-browser-languagedetector i18next-http-backend
```

2. **Initialize i18n**

```typescript
// src/main.tsx
import './i18n/config';

// Your app initialization...
```

3. **Use Translations**

```typescript
import { useI18n } from './i18n/hooks';

function MyComponent() {
  const { t, formatters } = useI18n();

  return (
    <div>
      <h1>{t('common.app.name')}</h1>
      <p>{formatters.number.currency(1234.56, 'USD')}</p>
    </div>
  );
}
```

### Backend Setup

1. **Initialize i18n**

```go
import "backend/i18n"

func main() {
    if err := i18n.Init("en-US"); err != nil {
        log.Fatal(err)
    }
}
```

2. **Add Middleware**

```go
detector := i18n.NewLanguageDetector("en-US")
router.Use(detector.Middleware)
```

3. **Use Translations**

```go
func handleRequest(w http.ResponseWriter, r *http.Request) {
    message := i18n.TranslateContext(r.Context(), "errors.network.timeout")

    json.NewEncoder(w).Encode(map[string]string{
        "error": message,
    })
}
```

## Features

### 1. Language Detection

Automatic detection with priority order:

1. Query parameter (`?lang=es-ES`)
2. Cookie (`language=es-ES`)
3. Local storage (`i18nextLng`)
4. Accept-Language header
5. Browser settings
6. Default language

### 2. Number Formatting

```typescript
const { formatters } = useI18n();

// Currency
formatters.number.currency(1234.56, 'USD')  // $1,234.56
formatters.number.currency(1234.56, 'EUR')  // €1.234,56

// Decimals
formatters.number.decimal(3.14159, 2)       // 3.14

// Percentages
formatters.number.percentage(15.5, 2)       // 15.50%

// Compact
formatters.number.compact(1500000)          // 1.5M
```

### 3. Date/Time Formatting

```typescript
const { formatters } = useI18n();

// Date
formatters.date.date(new Date())           // 01/18/2026 (US)
                                           // 18/01/2026 (EU)

// Time
formatters.date.time(new Date())           // 03:45:00 PM (US)
                                           // 15:45:00 (EU)

// Relative
formatters.date.relative(pastDate)         // 2 hours ago

// Long format
formatters.date.longDate(new Date())       // January 18, 2026
```

### 4. Email Templates

```go
emailTemplates := templates.InitializeDefaultTemplates()

data := map[string]interface{}{
    "UserName": "John Doe",
    "AppName": "Trading Engine",
    "OrderID": "ORD-12345",
    "Price": "$1,234.56",
}

email, _ := emailTemplates.Render("en-US", "trade_confirmation", data)
// email.Subject, email.Body, email.BodyHTML
```

Available templates:
- `welcome` - Welcome email
- `password_reset` - Password reset
- `trade_confirmation` - Trade confirmation
- Custom templates can be added

### 5. SMS Templates

```go
smsTemplates := templates.InitializeDefaultSMSTemplates()

data := map[string]interface{}{
    "AppName": "Trading Engine",
    "Code": "123456",
    "ExpiryMinutes": 5,
}

message, _ := smsTemplates.Render("en-US", "2fa_code", data)
```

Available templates:
- `2fa_code` - Two-factor authentication
- `trade_alert` - Trade execution
- `price_alert` - Price alerts
- `margin_call` - Margin call warnings
- `withdrawal_approved` - Withdrawal confirmation
- `login_alert` - Login notifications

### 6. RTL Support

```typescript
const { isRTL } = useI18n();

<div dir={isRTL ? 'rtl' : 'ltr'}>
  {/* Content automatically mirrors for RTL languages */}
</div>
```

CSS automatically adjusts:
- Text alignment
- Padding/margin directions
- Flexbox/Grid directions
- Border radius
- Position properties

### 7. Pluralization

```typescript
const { formatters } = useI18n();

formatters.plural.format(count, {
  zero: 'No orders',
  one: '1 order',
  other: '{{count}} orders'
})
```

## Translation Keys

Use semantic, hierarchical keys:

```
namespace.category.item

Examples:
- common.actions.save
- trading.orders.buyOrder
- errors.network.timeout
- notifications.orders.orderFilled
- legal.termsOfService.title
```

### Translation File Example

```json
{
  "orders": {
    "title": "Orders",
    "newOrder": "New Order",
    "status": {
      "pending": "Pending",
      "filled": "Filled",
      "cancelled": "Cancelled"
    }
  },
  "validation": {
    "required": "This field is required",
    "min": "Minimum value is {{min}}"
  }
}
```

## Testing

### 1. Pseudo-Localization

Test i18n implementation before real translations:

```typescript
import { pseudoLocalize, createPseudoTranslations } from './i18n/pseudo';

// Transform text
pseudoLocalize('Hello World')
// Output: [Ĥéļļö Ŵöŕļð ········]

// Create pseudo translations
const pseudo = createPseudoTranslations(englishTranslations);
```

Benefits:
- Identifies hardcoded strings
- Tests UI layout with longer text
- Verifies proper i18n integration
- Catches text truncation issues

### 2. RTL Testing

```typescript
import { enableRTLTesting } from './i18n/pseudo';

enableRTLTesting();
// UI switches to RTL mode
```

### 3. Character Encoding

```typescript
import { testCharacterEncoding } from './i18n/pseudo';

const result = testCharacterEncoding();
console.log(result.supportsUnicode);      // true
console.log(result.supportedScripts);     // ['arabic', 'chinese', ...]
```

### 4. Hardcoded String Detection

```typescript
import { detectHardcodedStrings } from './i18n/pseudo';

const hardcoded = detectHardcodedStrings(document.body);
if (hardcoded.length > 0) {
  console.error('Found hardcoded strings:', hardcoded);
}
```

## Validation

### Translation Completeness

```go
validator := i18n.NewValidator("en-US", "./locales")

result, err := validator.Validate()

if !result.IsValid {
    for lang, keys := range result.MissingKeys {
        log.Printf("%s missing %d keys\n", lang, len(keys))
        for _, key := range keys {
            log.Printf("  - %s\n", key)
        }
    }
}

// Coverage report
for lang, coverage := range result.CoveragePercent {
    log.Printf("%s: %.2f%% complete\n", lang, coverage)
}
```

### Generate Missing Template

```go
template, err := validator.GenerateMissingTranslations("es-ES")
// Returns organized list of missing keys by namespace
```

## Best Practices

### 1. Never Hardcode Strings

❌ **Bad:**
```typescript
<button>Save</button>
```

✅ **Good:**
```typescript
<button>{t('common.actions.save')}</button>
```

### 2. Use Proper Namespaces

❌ **Bad:**
```typescript
t('save')
t('button.save')
```

✅ **Good:**
```typescript
t('common.actions.save')
t('trading.orders.newOrder')
```

### 3. Parameterize Dynamic Content

❌ **Bad:**
```typescript
t('order.message') + orderId
```

✅ **Good:**
```typescript
t('order.messageWithId', { orderId })
```

### 4. Professional Translations

- Use professional translators for production
- Avoid machine translations for critical content
- Have native speakers review
- Consider cultural context

### 5. Test Thoroughly

- Test with actual languages
- Use pseudo-localization
- Test RTL layouts
- Verify number/date formatting
- Check email/SMS templates

### 6. Performance

- Lazy load translations
- Cache formatted values
- Use Suspense for loading states
- Pre-load user's preferred language

## Legal Compliance

### Regional Requirements

Each region has specific legal requirements:

#### United States
- CFTC/NFA warnings
- FINRA disclosures
- SEC compliance
- State-specific licensing

#### European Union
- MiFID II compliance
- ESMA warnings
- GDPR privacy notices
- Negative balance protection

#### United Kingdom
- FCA authorization
- FSCS protection notices
- Client money segregation

#### Asia Pacific
- Local regulatory notices
- License information
- Jurisdiction restrictions

### Implementation

```typescript
// Legal notices per region
const legalNotice = t('legal.regionalNotices.us.cftcWarning');

// Terms of service
const terms = t('legal.termsOfService.title');

// Risk disclosure
const riskWarning = t('legal.riskDisclosure.generalWarning');
```

## Adding a New Language

1. **Create Translation Files**

```bash
mkdir -p locales/xx-YY
cp -r locales/en-US/* locales/xx-YY/
```

2. **Translate Content**

Get professional translations for all JSON files.

3. **Update Configuration**

```typescript
// Frontend: src/i18n/config.ts
'xx-YY': {
  code: 'xx-YY',
  name: 'Language Name',
  nativeName: 'Native Name',
  flag: '🏳️',
  rtl: false,
}
```

```go
// Backend: backend/i18n/i18n.go
"xx-YY": {
    Code: "xx-YY",
    Name: "Language Name",
    NativeName: "Native Name",
    RTL: false,
},
```

4. **Add Email Templates**

```go
et.RegisterTemplate("xx-YY", "welcome", &EmailTemplate{
    Subject: "...",
    Body: "...",
    BodyHTML: "...",
})
```

5. **Add SMS Templates**

```go
st.RegisterTemplate("xx-YY", "2fa_code", &SMSTemplate{
    Message: "...",
})
```

6. **Validate**

```go
result, _ := validator.Validate()
// Check coverage for new language
```

7. **Test**

- All UI screens
- Email templates
- SMS templates
- Number formatting
- Date formatting
- RTL (if applicable)

## Troubleshooting

### Missing Translations

**Problem**: Key not found warning

**Solution**:
- Check key path is correct
- Verify JSON file structure
- Run validation to find missing keys

### Wrong Format

**Problem**: Numbers/dates display incorrectly

**Solution**:
- Verify locale code is correct
- Check browser/system locale support
- Test with Intl.NumberFormat/DateTimeFormat

### RTL Issues

**Problem**: Layout broken in RTL mode

**Solution**:
- Use logical properties (`margin-inline-start` vs `margin-left`)
- Test with Arabic language
- Use CSS `dir` attribute properly

### Slow Loading

**Problem**: Translations slow to load

**Solution**:
- Enable lazy loading
- Pre-load preferred language
- Use Suspense boundaries
- Cache translations

## Support

For issues or questions:
- Check documentation
- Review example implementations
- Contact development team

## License

Part of the Trading Engine platform.
