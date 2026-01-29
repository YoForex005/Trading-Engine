# i18n File Structure

Complete file organization for the internationalization system.

## Directory Tree

```
trading-engine/
├── backend/
│   ├── i18n/
│   │   ├── i18n.go                      # Core translator
│   │   ├── formatter.go                 # Number/date/time formatting
│   │   ├── middleware.go                # HTTP middleware
│   │   ├── context.go                   # Context integration
│   │   ├── validator.go                 # Translation validation
│   │   ├── templates/
│   │   │   ├── email_templates.go       # Email templates
│   │   │   └── sms_templates.go         # SMS templates
│   │   └── README.md                    # Backend i18n documentation
│   └── go.mod                           # Already includes golang.org/x/text
│
├── clients/
│   └── admin-dashboard/
│       ├── src/
│       │   ├── i18n/
│       │   │   ├── config.ts            # i18next configuration
│       │   │   ├── formatters.ts        # Number/date formatters
│       │   │   ├── hooks.ts             # Custom React hooks
│       │   │   ├── pseudo.ts            # Testing utilities
│       │   │   └── index.ts             # Module exports
│       │   ├── components/
│       │   │   └── LanguageSelector.tsx # Language switcher
│       │   └── App.tsx                  # Example integration
│       └── package.json                 # i18next dependencies added
│
├── locales/                             # Translation files (public)
│   ├── en-US/                           # English (Primary) ✅
│   │   ├── common.json
│   │   ├── trading.json
│   │   ├── errors.json
│   │   ├── notifications.json
│   │   └── legal.json
│   ├── es-ES/                           # Spanish 🟡
│   │   └── common.json
│   ├── fr-FR/                           # French ⚪
│   ├── de-DE/                           # German ⚪
│   ├── ja-JP/                           # Japanese ⚪
│   ├── zh-CN/                           # Chinese (Simplified) ⚪
│   ├── ar-SA/                           # Arabic (RTL) ⚪
│   ├── ru-RU/                           # Russian ⚪
│   ├── pt-BR/                           # Portuguese (Brazil) ⚪
│   └── it-IT/                           # Italian ⚪
│
├── docs/
│   └── i18n/
│       ├── README.md                    # Complete implementation guide
│       ├── IMPLEMENTATION_CHECKLIST.md  # Development checklist
│       ├── SUMMARY.md                   # Implementation summary
│       └── FILE_STRUCTURE.md            # This file
│
└── examples/
    └── i18n_example.go                  # Complete usage examples
```

## File Descriptions

### Backend Core (`backend/i18n/`)

#### i18n.go (Core Translator)
- `Translator` struct - Main translation engine
- `Init()` - Initialize global translator
- `T()` - Translate with parameters
- `IsRTL()` - RTL language detection
- Translation loading from embedded files
- Fallback mechanism
- Concurrent access support

**Key Functions:**
```go
Init(defaultLang string) error
T(lang, key string, params ...interface{}) string
IsRTL(lang string) bool
```

#### formatter.go (Formatters)
- `Formatter` struct - Locale-aware formatting
- `NumberFormatter` - Number, currency, percentage
- `DateFormatter` - Date, time, relative time
- Compact notation (1.2K, 1.5M)
- File size formatting

**Key Functions:**
```go
NewFormatter(lang string) *Formatter
FormatCurrency(value float64, currency string) string
FormatDate(t time.Time) string
FormatNumber(value float64, decimals int) string
```

#### middleware.go (HTTP Middleware)
- `LanguageDetector` - Language detection
- `Middleware()` - HTTP middleware
- Detection from query, cookie, header
- Context integration

**Key Functions:**
```go
NewLanguageDetector(defaultLang string) *LanguageDetector
DetectLanguage(r *http.Request) string
Middleware(next http.Handler) http.Handler
```

#### context.go (Context Support)
- Context key management
- Language storage/retrieval
- Convenience functions

**Key Functions:**
```go
WithLanguage(ctx context.Context, lang string) context.Context
FromContext(ctx context.Context) string
TranslateContext(ctx context.Context, key string, params ...interface{}) string
```

#### validator.go (Validation)
- Translation completeness checking
- Coverage reporting
- Missing key detection
- Template generation

**Key Functions:**
```go
NewValidator(baseLanguage, basePath string) *Validator
Validate() (*ValidationResult, error)
GenerateMissingTranslations(lang string) (map[string]interface{}, error)
```

### Backend Templates (`backend/i18n/templates/`)

#### email_templates.go
- HTML and plain text email templates
- Multi-language support
- Template rendering with data
- Pre-built templates: welcome, password reset, trade confirmation

**Templates:**
- `welcome` - New user welcome
- `password_reset` - Password reset request
- `trade_confirmation` - Trade execution notification

**Key Functions:**
```go
NewEmailTemplates() *EmailTemplates
RegisterTemplate(lang, name string, tmpl *EmailTemplate)
Render(lang, name string, data interface{}) (*EmailTemplate, error)
```

#### sms_templates.go
- Short message templates
- Character-limited formatting
- Multi-language support

**Templates:**
- `2fa_code` - Two-factor authentication
- `trade_alert` - Trade notifications
- `price_alert` - Price alerts
- `margin_call` - Margin warnings
- `withdrawal_approved` - Withdrawal confirmation
- `login_alert` - Login notifications

**Key Functions:**
```go
NewSMSTemplates() *SMSTemplates
Render(lang, name string, data interface{}) (string, error)
```

### Frontend Core (`clients/admin-dashboard/src/i18n/`)

#### config.ts (Configuration)
- i18next initialization
- Language metadata (code, name, RTL, flag)
- Detection options
- Backend configuration
- Interpolation settings

**Exports:**
```typescript
i18n: i18n instance
SUPPORTED_LANGUAGES: Language metadata
SupportedLanguage: Type definition
```

#### formatters.ts (Formatters)
- `NumberFormatter` - Number, currency, percentage, compact
- `DateFormatter` - Date, time, datetime, relative, long
- `PluralFormatter` - Pluralization
- `ListFormatter` - List formatting

**Classes:**
```typescript
class NumberFormatter
class DateFormatter
class PluralFormatter
class ListFormatter
createFormatters(locale: SupportedLanguage)
```

#### hooks.ts (React Hooks)
- `useI18n()` - Main hook with formatters
- `useCurrency()` - Currency formatting
- `useDate()` - Date formatting
- `useNumber()` - Number formatting
- `usePlural()` - Pluralization

**Hooks:**
```typescript
useI18n() => { t, i18n, formatters, changeLanguage, isRTL, ... }
useCurrency() => (value, currency) => string
useDate() => { date, time, dateTime, relative, ... }
useNumber() => { format, decimal, percentage, compact }
```

#### pseudo.ts (Testing)
- Pseudo-localization
- Hardcoded string detection
- RTL testing utilities
- Character encoding tests

**Functions:**
```typescript
pseudoLocalize(text: string, options?) => string
createPseudoTranslations(translations) => Record<string, any>
detectHardcodedStrings(node: Element) => string[]
enableRTLTesting() => void
testCharacterEncoding() => { supportsUnicode, supportedScripts }
```

### Components

#### LanguageSelector.tsx
- Dropdown language selector
- Flag icons
- Native language names
- Keyboard navigation
- Click-outside to close

**Props:** None (uses i18n context)

### Translation Files (`locales/`)

#### File Structure per Language
```
locales/{lang}/
├── common.json          # UI elements, actions, status
├── trading.json         # Trading terms, orders, positions
├── errors.json          # Error messages
├── notifications.json   # System notifications
└── legal.json          # Legal content, terms, policies
```

#### Key Namespaces

**common.json:**
- app (name, description, version)
- navigation (menu items)
- actions (save, cancel, delete, etc.)
- status (loading, success, error)
- time (today, yesterday, etc.)
- validation (required, email, etc.)
- confirmation (delete confirm, etc.)
- pagination
- accessibility

**trading.json:**
- orders (types, status, actions)
- positions (long, short, P&L)
- instruments (forex, crypto, stocks)
- charts (timeframes, indicators)
- risk (stop loss, take profit, margin)

**errors.json:**
- network (timeout, offline, server)
- authentication (invalid credentials, session expired)
- trading (insufficient balance, invalid order)
- validation (invalid email, phone, etc.)
- general (unknown error, operation failed)

**notifications.json:**
- orders (placed, filled, cancelled)
- positions (opened, closed, liquidated)
- account (deposit, withdrawal, KYC)
- market (price alerts, volatility)
- system (maintenance, updates)
- compliance (document required, trading restricted)

**legal.json:**
- termsOfService (acceptance, risks, restrictions)
- privacyPolicy (data collection, usage, rights)
- riskDisclosure (leverage, volatility, losses)
- aml (KYC, verification, compliance)
- regionalNotices (US, EU, UK, Asia)

## File Sizes

```
Backend:
├── i18n.go              ~8 KB   (core translator)
├── formatter.go         ~6 KB   (formatters)
├── middleware.go        ~4 KB   (HTTP middleware)
├── context.go           ~2 KB   (context support)
├── validator.go         ~6 KB   (validation)
├── email_templates.go   ~15 KB  (email templates)
└── sms_templates.go     ~8 KB   (SMS templates)
Total: ~49 KB

Frontend:
├── config.ts            ~3 KB   (configuration)
├── formatters.ts        ~8 KB   (formatters)
├── hooks.ts             ~4 KB   (React hooks)
├── pseudo.ts            ~6 KB   (testing)
├── index.ts             ~1 KB   (exports)
└── LanguageSelector.tsx ~3 KB   (component)
Total: ~25 KB

Translation Files (per language):
├── common.json          ~3 KB
├── trading.json         ~4 KB
├── errors.json          ~3 KB
├── notifications.json   ~3 KB
└── legal.json           ~5 KB
Total per language: ~18 KB
Total all languages (10): ~180 KB

Documentation:
├── README.md                    ~15 KB
├── IMPLEMENTATION_CHECKLIST.md  ~8 KB
├── SUMMARY.md                   ~10 KB
└── FILE_STRUCTURE.md            ~6 KB
Total: ~39 KB

Examples:
└── i18n_example.go              ~12 KB

Grand Total: ~305 KB
```

## Dependencies

### Backend
```go
// go.mod
golang.org/x/text v0.32.0  // Already included
```

### Frontend
```json
// package.json
{
  "i18next": "^23.7.16",
  "react-i18next": "^14.0.0",
  "i18next-browser-languagedetector": "^7.2.0",
  "i18next-http-backend": "^2.4.2"
}
```

## Installation

### Backend
```bash
# Dependencies already in go.mod
cd backend
go mod download
```

### Frontend
```bash
cd clients/admin-dashboard
npm install i18next react-i18next i18next-browser-languagedetector i18next-http-backend
```

## Usage

### Backend Initialization
```go
// main.go
import "backend/i18n"

func main() {
    if err := i18n.Init("en-US"); err != nil {
        log.Fatal(err)
    }
    
    // Add middleware
    detector := i18n.NewLanguageDetector("en-US")
    router.Use(detector.Middleware)
}
```

### Frontend Initialization
```typescript
// main.tsx
import './i18n/config';

// App.tsx
import { I18nextProvider } from 'react-i18next';
import i18n from './i18n/config';

<I18nextProvider i18n={i18n}>
  <App />
</I18nextProvider>
```

## Status Legend

- ✅ Complete and tested
- 🟡 Partially complete
- ⚪ Not started
- 📋 Planned

## Maintenance

### Adding New Translations
1. Create JSON files in `locales/{lang}/`
2. Follow existing structure
3. Run validation
4. Test thoroughly

### Adding New Keys
1. Add to English (en-US) first
2. Use semantic naming
3. Run validation to find missing translations
4. Update all languages

### Updating Templates
1. Update Go template code
2. Test rendering with sample data
3. Update all language variants
4. Test email/SMS delivery

---

**Last Updated**: 2026-01-18
**Maintained By**: Development Team
