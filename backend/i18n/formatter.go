package i18n

import (
	"fmt"
	"time"

	"golang.org/x/text/currency"
	"golang.org/x/text/language"
	"golang.org/x/text/message"
)

// Formatter provides locale-aware formatting
type Formatter struct {
	lang    language.Tag
	printer *message.Printer
}

// NewFormatter creates a new formatter for the given language
func NewFormatter(lang string) *Formatter {
	tag := language.MustParse(lang)
	return &Formatter{
		lang:    tag,
		printer: message.NewPrinter(tag),
	}
}

// FormatNumber formats a number with locale-specific separators
func (f *Formatter) FormatNumber(value float64, decimals int) string {
	return f.printer.Sprintf("%.*f", decimals, value)
}

// FormatCurrency formats a currency value
func (f *Formatter) FormatCurrency(value float64, currencyCode string) string {
	cur := currency.MustParseISO(currencyCode)
	// Use printer with currency unit - golang.org/x/text doesn't have number.Currency
	return f.printer.Sprintf("%s %.2f", cur.String(), value)
}

// FormatPercentage formats a percentage value
func (f *Formatter) FormatPercentage(value float64, decimals int) string {
	return f.printer.Sprintf("%.*f%%", decimals, value)
}

// FormatDate formats a date according to locale
func (f *Formatter) FormatDate(t time.Time) string {
	// Format based on locale
	switch f.lang.String() {
	case "en-US":
		return t.Format("01/02/2006")
	case "en-GB", "fr-FR", "de-DE", "es-ES", "it-IT", "pt-BR":
		return t.Format("02/01/2006")
	case "ja-JP", "zh-CN":
		return t.Format("2006/01/02")
	default:
		return t.Format("2006-01-02")
	}
}

// FormatDateTime formats a date and time according to locale
func (f *Formatter) FormatDateTime(t time.Time) string {
	dateStr := f.FormatDate(t)
	timeStr := f.FormatTime(t)
	return fmt.Sprintf("%s %s", dateStr, timeStr)
}

// FormatTime formats a time according to locale
func (f *Formatter) FormatTime(t time.Time) string {
	// US uses 12-hour format, others use 24-hour
	if f.lang.String() == "en-US" {
		return t.Format("03:04:05 PM")
	}
	return t.Format("15:04:05")
}

// FormatRelativeTime formats a relative time (e.g., "2 hours ago")
func (f *Formatter) FormatRelativeTime(t time.Time) string {
	now := time.Now()
	diff := now.Sub(t)

	if diff < time.Minute {
		seconds := int(diff.Seconds())
		return f.printer.Sprintf("%d seconds ago", seconds)
	} else if diff < time.Hour {
		minutes := int(diff.Minutes())
		return f.printer.Sprintf("%d minutes ago", minutes)
	} else if diff < 24*time.Hour {
		hours := int(diff.Hours())
		return f.printer.Sprintf("%d hours ago", hours)
	} else if diff < 30*24*time.Hour {
		days := int(diff.Hours() / 24)
		return f.printer.Sprintf("%d days ago", days)
	} else if diff < 365*24*time.Hour {
		months := int(diff.Hours() / 24 / 30)
		return f.printer.Sprintf("%d months ago", months)
	} else {
		years := int(diff.Hours() / 24 / 365)
		return f.printer.Sprintf("%d years ago", years)
	}
}

// FormatCompactNumber formats a number in compact notation (1.2K, 1.5M, etc.)
func (f *Formatter) FormatCompactNumber(value float64) string {
	abs := value
	if abs < 0 {
		abs = -abs
	}

	if abs < 1000 {
		return f.FormatNumber(value, 0)
	} else if abs < 1000000 {
		return f.printer.Sprintf("%.1fK", value/1000)
	} else if abs < 1000000000 {
		return f.printer.Sprintf("%.1fM", value/1000000)
	} else {
		return f.printer.Sprintf("%.1fB", value/1000000000)
	}
}

// FormatFileSize formats a file size in bytes
func (f *Formatter) FormatFileSize(bytes int64) string {
	const unit = 1024
	if bytes < unit {
		return f.printer.Sprintf("%d B", bytes)
	}

	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}

	units := []string{"KB", "MB", "GB", "TB", "PB"}
	return f.printer.Sprintf("%.1f %s", float64(bytes)/float64(div), units[exp])
}
