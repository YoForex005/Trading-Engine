package templates

import (
	"bytes"
	"fmt"
	"html/template"
)

// EmailTemplate represents an email template
type EmailTemplate struct {
	Subject  string
	Body     string
	BodyHTML string
}

// EmailTemplates stores all email templates by language
type EmailTemplates struct {
	templates map[string]map[string]*EmailTemplate
}

// NewEmailTemplates creates a new email templates collection
func NewEmailTemplates() *EmailTemplates {
	return &EmailTemplates{
		templates: make(map[string]map[string]*EmailTemplate),
	}
}

// RegisterTemplate registers an email template
func (et *EmailTemplates) RegisterTemplate(lang, name string, tmpl *EmailTemplate) {
	if et.templates[lang] == nil {
		et.templates[lang] = make(map[string]*EmailTemplate)
	}
	et.templates[lang][name] = tmpl
}

// Render renders an email template with data
func (et *EmailTemplates) Render(lang, name string, data interface{}) (*EmailTemplate, error) {
	tmpl, ok := et.templates[lang][name]
	if !ok {
		// Fallback to English
		tmpl, ok = et.templates["en-US"][name]
		if !ok {
			return nil, fmt.Errorf("template not found: %s", name)
		}
	}

	result := &EmailTemplate{}

	// Render subject
	subjectTmpl, err := template.New("subject").Parse(tmpl.Subject)
	if err != nil {
		return nil, fmt.Errorf("failed to parse subject template: %w", err)
	}
	var subjectBuf bytes.Buffer
	if err := subjectTmpl.Execute(&subjectBuf, data); err != nil {
		return nil, fmt.Errorf("failed to execute subject template: %w", err)
	}
	result.Subject = subjectBuf.String()

	// Render plain text body
	bodyTmpl, err := template.New("body").Parse(tmpl.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to parse body template: %w", err)
	}
	var bodyBuf bytes.Buffer
	if err := bodyTmpl.Execute(&bodyBuf, data); err != nil {
		return nil, fmt.Errorf("failed to execute body template: %w", err)
	}
	result.Body = bodyBuf.String()

	// Render HTML body
	if tmpl.BodyHTML != "" {
		htmlTmpl, err := template.New("html").Parse(tmpl.BodyHTML)
		if err != nil {
			return nil, fmt.Errorf("failed to parse HTML template: %w", err)
		}
		var htmlBuf bytes.Buffer
		if err := htmlTmpl.Execute(&htmlBuf, data); err != nil {
			return nil, fmt.Errorf("failed to execute HTML template: %w", err)
		}
		result.BodyHTML = htmlBuf.String()
	}

	return result, nil
}

// InitializeDefaultTemplates initializes default email templates
func InitializeDefaultTemplates() *EmailTemplates {
	et := NewEmailTemplates()

	// Welcome Email - English
	et.RegisterTemplate("en-US", "welcome", &EmailTemplate{
		Subject: "Welcome to {{.AppName}}!",
		Body: `Dear {{.UserName}},

Welcome to {{.AppName}}! We're excited to have you on board.

Your account has been successfully created. You can now start trading on our platform.

To get started:
1. Complete your profile verification
2. Fund your account
3. Start trading

If you have any questions, please don't hesitate to contact our support team.

Best regards,
The {{.AppName}} Team`,
		BodyHTML: `<!DOCTYPE html>
<html>
<head>
    <meta charset="utf-8">
</head>
<body style="font-family: Arial, sans-serif; line-height: 1.6; color: #333;">
    <div style="max-width: 600px; margin: 0 auto; padding: 20px;">
        <h1 style="color: #2563eb;">Welcome to {{.AppName}}!</h1>
        <p>Dear {{.UserName}},</p>
        <p>Welcome to {{.AppName}}! We're excited to have you on board.</p>
        <p>Your account has been successfully created. You can now start trading on our platform.</p>
        <h3>To get started:</h3>
        <ol>
            <li>Complete your profile verification</li>
            <li>Fund your account</li>
            <li>Start trading</li>
        </ol>
        <p>If you have any questions, please don't hesitate to contact our support team.</p>
        <p>Best regards,<br>The {{.AppName}} Team</p>
    </div>
</body>
</html>`,
	})

	// Welcome Email - Spanish
	et.RegisterTemplate("es-ES", "welcome", &EmailTemplate{
		Subject: "¡Bienvenido a {{.AppName}}!",
		Body: `Estimado/a {{.UserName}},

¡Bienvenido/a a {{.AppName}}! Estamos emocionados de tenerte con nosotros.

Tu cuenta ha sido creada exitosamente. Ahora puedes comenzar a operar en nuestra plataforma.

Para comenzar:
1. Completa la verificación de tu perfil
2. Fondea tu cuenta
3. Comienza a operar

Si tienes alguna pregunta, no dudes en contactar a nuestro equipo de soporte.

Saludos cordiales,
El equipo de {{.AppName}}`,
		BodyHTML: `<!DOCTYPE html>
<html>
<head>
    <meta charset="utf-8">
</head>
<body style="font-family: Arial, sans-serif; line-height: 1.6; color: #333;">
    <div style="max-width: 600px; margin: 0 auto; padding: 20px;">
        <h1 style="color: #2563eb;">¡Bienvenido a {{.AppName}}!</h1>
        <p>Estimado/a {{.UserName}},</p>
        <p>¡Bienvenido/a a {{.AppName}}! Estamos emocionados de tenerte con nosotros.</p>
        <p>Tu cuenta ha sido creada exitosamente. Ahora puedes comenzar a operar en nuestra plataforma.</p>
        <h3>Para comenzar:</h3>
        <ol>
            <li>Completa la verificación de tu perfil</li>
            <li>Fondea tu cuenta</li>
            <li>Comienza a operar</li>
        </ol>
        <p>Si tienes alguna pregunta, no dudes en contactar a nuestro equipo de soporte.</p>
        <p>Saludos cordiales,<br>El equipo de {{.AppName}}</p>
    </div>
</body>
</html>`,
	})

	// Password Reset - English
	et.RegisterTemplate("en-US", "password_reset", &EmailTemplate{
		Subject: "Password Reset Request",
		Body: `Dear {{.UserName}},

We received a request to reset your password for your {{.AppName}} account.

To reset your password, please use the following code:

{{.ResetCode}}

This code will expire in {{.ExpiryMinutes}} minutes.

If you did not request this password reset, please ignore this email or contact support if you have concerns.

Best regards,
The {{.AppName}} Team`,
		BodyHTML: `<!DOCTYPE html>
<html>
<head>
    <meta charset="utf-8">
</head>
<body style="font-family: Arial, sans-serif; line-height: 1.6; color: #333;">
    <div style="max-width: 600px; margin: 0 auto; padding: 20px;">
        <h1 style="color: #2563eb;">Password Reset Request</h1>
        <p>Dear {{.UserName}},</p>
        <p>We received a request to reset your password for your {{.AppName}} account.</p>
        <p>To reset your password, please use the following code:</p>
        <div style="background-color: #f3f4f6; padding: 20px; text-align: center; margin: 20px 0;">
            <h2 style="margin: 0; font-size: 32px; letter-spacing: 5px;">{{.ResetCode}}</h2>
        </div>
        <p>This code will expire in {{.ExpiryMinutes}} minutes.</p>
        <p>If you did not request this password reset, please ignore this email or contact support if you have concerns.</p>
        <p>Best regards,<br>The {{.AppName}} Team</p>
    </div>
</body>
</html>`,
	})

	// Trade Confirmation - English
	et.RegisterTemplate("en-US", "trade_confirmation", &EmailTemplate{
		Subject: "Trade Confirmation - {{.OrderID}}",
		Body: `Dear {{.UserName}},

Your trade has been executed:

Order ID: {{.OrderID}}
Symbol: {{.Symbol}}
Side: {{.Side}}
Quantity: {{.Quantity}}
Price: {{.Price}}
Total Value: {{.TotalValue}}
Status: {{.Status}}
Execution Time: {{.ExecutionTime}}

Thank you for trading with {{.AppName}}.

Best regards,
The {{.AppName}} Team`,
		BodyHTML: `<!DOCTYPE html>
<html>
<head>
    <meta charset="utf-8">
</head>
<body style="font-family: Arial, sans-serif; line-height: 1.6; color: #333;">
    <div style="max-width: 600px; margin: 0 auto; padding: 20px;">
        <h1 style="color: #2563eb;">Trade Confirmation</h1>
        <p>Dear {{.UserName}},</p>
        <p>Your trade has been executed:</p>
        <table style="width: 100%; border-collapse: collapse; margin: 20px 0;">
            <tr style="border-bottom: 1px solid #e5e7eb;">
                <td style="padding: 10px; font-weight: bold;">Order ID:</td>
                <td style="padding: 10px;">{{.OrderID}}</td>
            </tr>
            <tr style="border-bottom: 1px solid #e5e7eb;">
                <td style="padding: 10px; font-weight: bold;">Symbol:</td>
                <td style="padding: 10px;">{{.Symbol}}</td>
            </tr>
            <tr style="border-bottom: 1px solid #e5e7eb;">
                <td style="padding: 10px; font-weight: bold;">Side:</td>
                <td style="padding: 10px;">{{.Side}}</td>
            </tr>
            <tr style="border-bottom: 1px solid #e5e7eb;">
                <td style="padding: 10px; font-weight: bold;">Quantity:</td>
                <td style="padding: 10px;">{{.Quantity}}</td>
            </tr>
            <tr style="border-bottom: 1px solid #e5e7eb;">
                <td style="padding: 10px; font-weight: bold;">Price:</td>
                <td style="padding: 10px;">{{.Price}}</td>
            </tr>
            <tr style="border-bottom: 1px solid #e5e7eb;">
                <td style="padding: 10px; font-weight: bold;">Total Value:</td>
                <td style="padding: 10px;">{{.TotalValue}}</td>
            </tr>
            <tr style="border-bottom: 1px solid #e5e7eb;">
                <td style="padding: 10px; font-weight: bold;">Status:</td>
                <td style="padding: 10px;">{{.Status}}</td>
            </tr>
            <tr>
                <td style="padding: 10px; font-weight: bold;">Execution Time:</td>
                <td style="padding: 10px;">{{.ExecutionTime}}</td>
            </tr>
        </table>
        <p>Thank you for trading with {{.AppName}}.</p>
        <p>Best regards,<br>The {{.AppName}} Team</p>
    </div>
</body>
</html>`,
	})

	return et
}
