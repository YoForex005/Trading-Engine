# i18n Quick Start Guide

Get internationalization up and running in 5 minutes.

## Prerequisites

- Go 1.21+
- Node.js 18+
- npm or yarn

## Backend Setup (2 minutes)

### 1. Initialize i18n

Add to your `main.go`:

```go
import "backend/i18n"

func main() {
    // Initialize i18n with default language
    if err := i18n.Init("en-US"); err != nil {
        log.Fatal(err)
    }

    // Your existing code...
}
```

### 2. Add Middleware

```go
import "backend/i18n"

func setupRouter() *http.ServeMux {
    router := http.NewServeMux()
    
    // Create language detector
    detector := i18n.NewLanguageDetector("en-US")
    
    // Add middleware
    http.Handle("/", detector.Middleware(router))
    
    return router
}
```

### 3. Use in Handlers

```go
func myHandler(w http.ResponseWriter, r *http.Request) {
    // Get language from context
    lang := i18n.FromContext(r.Context())
    
    // Translate
    message := i18n.T(lang, "common.actions.save")
    
    // Format numbers
    formatter := i18n.FormatContext(r.Context())
    price := formatter.FormatCurrency(1234.56, "USD")
    
    // Send response
    json.NewEncoder(w).Encode(map[string]string{
        "message": message,
        "price": price,
    })
}
```

## Frontend Setup (3 minutes)

### 1. Install Dependencies

```bash
cd clients/admin-dashboard
npm install i18next react-i18next i18next-browser-languagedetector i18next-http-backend
```

### 2. Initialize i18n

The configuration is already created in `src/i18n/config.ts`.

Import it in your entry point (`main.tsx` or `index.tsx`):

```typescript
import './i18n/config';
import React from 'react';
import ReactDOM from 'react-dom/client';
import App from './App';

ReactDOM.createRoot(document.getElementById('root')!).render(
  <React.StrictMode>
    <App />
  </React.StrictMode>
);
```

### 3. Wrap App with Provider

Update `App.tsx`:

```typescript
import { Suspense } from 'react';
import { I18nextProvider } from 'react-i18next';
import i18n from './i18n/config';
import { LanguageSelector } from './components/LanguageSelector';

function App() {
  return (
    <I18nextProvider i18n={i18n}>
      <Suspense fallback={<div>Loading...</div>}>
        <div className="app">
          <header>
            <LanguageSelector />
          </header>
          
          <main>
            {/* Your app content */}
          </main>
        </div>
      </Suspense>
    </I18nextProvider>
  );
}

export default App;
```

### 4. Use in Components

```typescript
import { useI18n } from './i18n/hooks';

function MyComponent() {
  const { t, formatters } = useI18n();

  return (
    <div>
      <h1>{t('common.app.name')}</h1>
      <button>{t('common.actions.save')}</button>
      <p>Price: {formatters.number.currency(1234.56, 'USD')}</p>
      <p>Date: {formatters.date.date(new Date())}</p>
    </div>
  );
}
```

## Testing (1 minute)

### Backend Test

```bash
cd backend
go run examples/i18n_example.go
```

You should see formatted output for different languages.

### Frontend Test

```bash
cd clients/admin-dashboard
npm run dev
```

Visit http://localhost:5173 and:
1. Click the language selector
2. Switch between languages
3. Verify text changes

## Common Use Cases

### 1. Display Translated Text

```typescript
// Simple translation
{t('common.actions.save')}

// With parameters
{t('validation.min', { min: 10 })}

// With namespace
{t('trading.orders.newOrder')}
```

### 2. Format Currency

```typescript
const { formatters } = useI18n();

// Format as USD
formatters.number.currency(1234.56, 'USD')
// Output: $1,234.56

// Format as EUR (Spanish locale)
formatters.number.currency(1234.56, 'EUR')
// Output: 1.234,56 €
```

### 3. Format Dates

```typescript
const { formatters } = useI18n();

// Short date
formatters.date.date(new Date())
// US: 01/18/2026
// EU: 18/01/2026

// Relative time
formatters.date.relative(twoHoursAgo)
// Output: 2 hours ago
```

### 4. Change Language

```typescript
const { changeLanguage } = useI18n();

<button onClick={() => changeLanguage('es-ES')}>
  Español
</button>
```

### 5. Backend Email

```go
import "backend/i18n/templates"

emailTemplates := templates.InitializeDefaultTemplates()

data := map[string]interface{}{
    "UserName": "John Doe",
    "AppName": "Trading Engine",
}

email, _ := emailTemplates.Render("en-US", "welcome", data)

// Send email using email.Subject, email.Body, email.BodyHTML
```

### 6. Backend SMS

```go
import "backend/i18n/templates"

smsTemplates := templates.InitializeDefaultSMSTemplates()

data := map[string]interface{}{
    "AppName": "Trading Engine",
    "Code": "123456",
    "ExpiryMinutes": 5,
}

message, _ := smsTemplates.Render("en-US", "2fa_code", data)

// Send SMS with message
```

## Translation Keys Reference

### Common Actions
```typescript
t('common.actions.save')      // Save
t('common.actions.cancel')    // Cancel
t('common.actions.delete')    // Delete
t('common.actions.edit')      // Edit
```

### Trading
```typescript
t('trading.orders.buyOrder')    // Buy Order
t('trading.orders.sellOrder')   // Sell Order
t('trading.positions.long')     // Long
t('trading.positions.short')    // Short
```

### Errors
```typescript
t('errors.network.timeout')                      // Request timed out
t('errors.trading.insufficientBalance')          // Insufficient balance
t('errors.authentication.invalidCredentials')    // Invalid credentials
```

### Notifications
```typescript
t('notifications.orders.orderPlaced')    // Order placed successfully
t('notifications.orders.orderFilled')    // Order filled at {{price}}
```

## Language Detection

Languages are detected in this order:

1. **Query parameter**: `?lang=es-ES`
2. **Cookie**: `language=es-ES`
3. **Local storage**: `i18nextLng`
4. **Accept-Language header**: Browser settings
5. **Default**: `en-US`

## Validation

Check translation completeness:

```go
validator := i18n.NewValidator("en-US", "./locales")
result, _ := validator.Validate()

fmt.Printf("Total keys: %d\n", result.TotalKeys)
fmt.Printf("Valid: %v\n", result.IsValid)

for lang, coverage := range result.CoveragePercent {
    fmt.Printf("%s: %.2f%%\n", lang, coverage)
}
```

## Troubleshooting

### Backend

**Issue**: Translations not found

**Solution**: Ensure `locales/` directory is in the correct location and files are properly formatted JSON.

**Issue**: Wrong number format

**Solution**: Check locale code matches exactly (e.g., `en-US` not `en`).

### Frontend

**Issue**: `i18n is not initialized`

**Solution**: Import `./i18n/config` in your entry point before rendering.

**Issue**: Translations not updating

**Solution**: Check Suspense fallback is working, and files are in `public/locales/`.

**Issue**: Language not persisting

**Solution**: Ensure local storage is enabled and not blocked.

## Next Steps

1. **Add More Translations**: Complete all 10 languages
2. **Update Components**: Replace hardcoded strings with i18n
3. **Add Email Templates**: Create all required email templates
4. **Test RTL**: Implement and test Arabic layout
5. **Production**: Deploy with proper CDN for translation files

## Resources

- [Full Documentation](./README.md)
- [Implementation Checklist](./IMPLEMENTATION_CHECKLIST.md)
- [Examples](../../examples/i18n_example.go)
- [i18next Documentation](https://www.i18next.com/)
- [React i18next](https://react.i18next.com/)

## Support

For issues:
1. Check documentation
2. Run validation tests
3. Review example code
4. Contact development team

---

**Setup Time**: ~5 minutes
**Status**: Ready to use
**Last Updated**: 2026-01-18
