package templates

import (
	"bytes"
	"fmt"
	"text/template"
)

// SMSTemplate represents an SMS template
type SMSTemplate struct {
	Message string
}

// SMSTemplates stores all SMS templates by language
type SMSTemplates struct {
	templates map[string]map[string]*SMSTemplate
}

// NewSMSTemplates creates a new SMS templates collection
func NewSMSTemplates() *SMSTemplates {
	return &SMSTemplates{
		templates: make(map[string]map[string]*SMSTemplate),
	}
}

// RegisterTemplate registers an SMS template
func (st *SMSTemplates) RegisterTemplate(lang, name string, tmpl *SMSTemplate) {
	if st.templates[lang] == nil {
		st.templates[lang] = make(map[string]*SMSTemplate)
	}
	st.templates[lang][name] = tmpl
}

// Render renders an SMS template with data
func (st *SMSTemplates) Render(lang, name string, data interface{}) (string, error) {
	tmpl, ok := st.templates[lang][name]
	if !ok {
		// Fallback to English
		tmpl, ok = st.templates["en-US"][name]
		if !ok {
			return "", fmt.Errorf("template not found: %s", name)
		}
	}

	msgTmpl, err := template.New("message").Parse(tmpl.Message)
	if err != nil {
		return "", fmt.Errorf("failed to parse SMS template: %w", err)
	}

	var buf bytes.Buffer
	if err := msgTmpl.Execute(&buf, data); err != nil {
		return "", fmt.Errorf("failed to execute SMS template: %w", err)
	}

	return buf.String(), nil
}

// InitializeDefaultSMSTemplates initializes default SMS templates
func InitializeDefaultSMSTemplates() *SMSTemplates {
	st := NewSMSTemplates()

	// Two-Factor Authentication - English
	st.RegisterTemplate("en-US", "2fa_code", &SMSTemplate{
		Message: "{{.AppName}}: Your verification code is {{.Code}}. Valid for {{.ExpiryMinutes}} minutes. Do not share this code.",
	})

	// Two-Factor Authentication - Spanish
	st.RegisterTemplate("es-ES", "2fa_code", &SMSTemplate{
		Message: "{{.AppName}}: Tu código de verificación es {{.Code}}. Válido por {{.ExpiryMinutes}} minutos. No compartas este código.",
	})

	// Two-Factor Authentication - French
	st.RegisterTemplate("fr-FR", "2fa_code", &SMSTemplate{
		Message: "{{.AppName}}: Votre code de vérification est {{.Code}}. Valable {{.ExpiryMinutes}} minutes. Ne partagez pas ce code.",
	})

	// Trade Alert - English
	st.RegisterTemplate("en-US", "trade_alert", &SMSTemplate{
		Message: "{{.AppName}}: Trade executed - {{.Side}} {{.Quantity}} {{.Symbol}} @ {{.Price}}",
	})

	// Trade Alert - Spanish
	st.RegisterTemplate("es-ES", "trade_alert", &SMSTemplate{
		Message: "{{.AppName}}: Operación ejecutada - {{.Side}} {{.Quantity}} {{.Symbol}} @ {{.Price}}",
	})

	// Price Alert - English
	st.RegisterTemplate("en-US", "price_alert", &SMSTemplate{
		Message: "{{.AppName}}: Price Alert - {{.Symbol}} reached {{.Price}}",
	})

	// Price Alert - Spanish
	st.RegisterTemplate("es-ES", "price_alert", &SMSTemplate{
		Message: "{{.AppName}}: Alerta de Precio - {{.Symbol}} alcanzó {{.Price}}",
	})

	// Margin Call - English
	st.RegisterTemplate("en-US", "margin_call", &SMSTemplate{
		Message: "{{.AppName}}: URGENT - Margin call. Current margin level: {{.MarginLevel}}%. Please add funds or close positions.",
	})

	// Margin Call - Spanish
	st.RegisterTemplate("es-ES", "margin_call", &SMSTemplate{
		Message: "{{.AppName}}: URGENTE - Llamada de margen. Nivel de margen actual: {{.MarginLevel}}%. Por favor agrega fondos o cierra posiciones.",
	})

	// Withdrawal Approved - English
	st.RegisterTemplate("en-US", "withdrawal_approved", &SMSTemplate{
		Message: "{{.AppName}}: Your withdrawal of {{.Amount}} has been approved and will be processed within {{.ProcessingDays}} business days.",
	})

	// Withdrawal Approved - Spanish
	st.RegisterTemplate("es-ES", "withdrawal_approved", &SMSTemplate{
		Message: "{{.AppName}}: Tu retiro de {{.Amount}} ha sido aprobado y será procesado en {{.ProcessingDays}} días hábiles.",
	})

	// Login Alert - English
	st.RegisterTemplate("en-US", "login_alert", &SMSTemplate{
		Message: "{{.AppName}}: New login detected from {{.Location}}. If this wasn't you, please secure your account immediately.",
	})

	// Login Alert - Spanish
	st.RegisterTemplate("es-ES", "login_alert", &SMSTemplate{
		Message: "{{.AppName}}: Nuevo inicio de sesión detectado desde {{.Location}}. Si no fuiste tú, asegura tu cuenta inmediatamente.",
	})

	return st
}
