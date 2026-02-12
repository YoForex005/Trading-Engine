package notifications

import (
	"bytes"
	"fmt"
	"html/template"
	"strings"
)

// Email template constants
const (
	TemplateMarginWarning        = "margin_warning"
	TemplateOrderFilled          = "order_filled"
	TemplateDepositConfirmed     = "deposit_confirmed"
	TemplateWithdrawalProcessed  = "withdrawal_processed"
	TemplateLoginAlert           = "login_alert"
	TemplatePasswordChanged      = "password_changed"
	TemplateReportGenerated      = "report_generated"
)

// emailTemplates contains all email templates as Go html/template strings
var emailTemplates = map[string]string{
	TemplateMarginWarning: `
<!DOCTYPE html>
<html>
<head>
	<meta charset="UTF-8">
	<style>
		body { font-family: Arial, sans-serif; line-height: 1.6; color: #333; background: #f4f4f4; }
		.container { max-width: 600px; margin: 20px auto; background: white; border-radius: 8px; overflow: hidden; box-shadow: 0 2px 4px rgba(0,0,0,0.1); }
		.header { background: #FF5722; color: white; padding: 20px; text-align: center; }
		.content { padding: 30px; }
		.warning-box { background: #FFF3CD; border-left: 4px solid #FF5722; padding: 15px; margin: 20px 0; }
		.button { display: inline-block; padding: 12px 24px; background: #FF5722; color: white; text-decoration: none; border-radius: 5px; margin: 10px 0; }
		.footer { background: #f9f9f9; padding: 20px; text-align: center; font-size: 12px; color: #666; }
	</style>
</head>
<body>
	<div class="container">
		<div class="header">
			<h1>⚠️ Margin Warning</h1>
		</div>
		<div class="content">
			<div class="warning-box">
				<h3>Action Required</h3>
				<p><strong>Your margin level is at {{.level}}%.</strong></p>
				<p>Please deposit funds or reduce positions to avoid margin call.</p>
			</div>
			<p>Current Status:</p>
			<ul>
				<li>Margin Level: <strong>{{.level}}%</strong></li>
				<li>Required Level: <strong>100%</strong></li>
			</ul>
			<p>To protect your account, please take action:</p>
			<ul>
				<li>Deposit additional funds</li>
				<li>Close some open positions</li>
				<li>Reduce position sizes</li>
			</ul>
			<a href="{{.accountUrl}}" class="button">Manage Account</a>
		</div>
		<div class="footer">
			<p>RTX5 Trading Platform | This is an automated notification</p>
		</div>
	</div>
</body>
</html>`,

	TemplateOrderFilled: `
<!DOCTYPE html>
<html>
<head>
	<meta charset="UTF-8">
	<style>
		body { font-family: Arial, sans-serif; line-height: 1.6; color: #333; background: #f4f4f4; }
		.container { max-width: 600px; margin: 20px auto; background: white; border-radius: 8px; overflow: hidden; box-shadow: 0 2px 4px rgba(0,0,0,0.1); }
		.header { background: #2196F3; color: white; padding: 20px; text-align: center; }
		.content { padding: 30px; }
		.order-details { background: #f9f9f9; padding: 20px; border-radius: 5px; margin: 20px 0; }
		.detail-row { display: flex; justify-content: space-between; padding: 8px 0; border-bottom: 1px solid #e0e0e0; }
		.detail-row:last-child { border-bottom: none; }
		.label { font-weight: bold; color: #555; }
		.value { color: #333; }
		.footer { background: #f9f9f9; padding: 20px; text-align: center; font-size: 12px; color: #666; }
	</style>
</head>
<body>
	<div class="container">
		<div class="header">
			<h1>✅ Order Filled</h1>
		</div>
		<div class="content">
			<p>Your {{.type}} order has been successfully filled.</p>
			<div class="order-details">
				<div class="detail-row">
					<span class="label">Symbol:</span>
					<span class="value">{{.symbol}}</span>
				</div>
				<div class="detail-row">
					<span class="label">Type:</span>
					<span class="value">{{.type}}</span>
				</div>
				<div class="detail-row">
					<span class="label">Volume:</span>
					<span class="value">{{.volume}} lots</span>
				</div>
				<div class="detail-row">
					<span class="label">Fill Price:</span>
					<span class="value">{{.price}}</span>
				</div>
			</div>
		</div>
		<div class="footer">
			<p>RTX5 Trading Platform | This is an automated notification</p>
		</div>
	</div>
</body>
</html>`,

	TemplateDepositConfirmed: `
<!DOCTYPE html>
<html>
<head>
	<meta charset="UTF-8">
	<style>
		body { font-family: Arial, sans-serif; line-height: 1.6; color: #333; background: #f4f4f4; }
		.container { max-width: 600px; margin: 20px auto; background: white; border-radius: 8px; overflow: hidden; box-shadow: 0 2px 4px rgba(0,0,0,0.1); }
		.header { background: #4CAF50; color: white; padding: 20px; text-align: center; }
		.content { padding: 30px; }
		.success-box { background: #E8F5E9; border-left: 4px solid #4CAF50; padding: 15px; margin: 20px 0; }
		.amount { font-size: 24px; font-weight: bold; color: #4CAF50; }
		.footer { background: #f9f9f9; padding: 20px; text-align: center; font-size: 12px; color: #666; }
	</style>
</head>
<body>
	<div class="container">
		<div class="header">
			<h1>💰 Deposit Confirmed</h1>
		</div>
		<div class="content">
			<div class="success-box">
				<p class="amount">${{.amount}}</p>
				<p>Your deposit has been successfully confirmed and credited to your account.</p>
			</div>
			<p><strong>Payment Method:</strong> {{.method}}</p>
			<p>You can now use these funds for trading.</p>
		</div>
		<div class="footer">
			<p>RTX5 Trading Platform | This is an automated notification</p>
		</div>
	</div>
</body>
</html>`,

	TemplateWithdrawalProcessed: `
<!DOCTYPE html>
<html>
<head>
	<meta charset="UTF-8">
	<style>
		body { font-family: Arial, sans-serif; line-height: 1.6; color: #333; background: #f4f4f4; }
		.container { max-width: 600px; margin: 20px auto; background: white; border-radius: 8px; overflow: hidden; box-shadow: 0 2px 4px rgba(0,0,0,0.1); }
		.header { background: #2196F3; color: white; padding: 20px; text-align: center; }
		.content { padding: 30px; }
		.info-box { background: #E3F2FD; border-left: 4px solid #2196F3; padding: 15px; margin: 20px 0; }
		.amount { font-size: 24px; font-weight: bold; color: #2196F3; }
		.footer { background: #f9f9f9; padding: 20px; text-align: center; font-size: 12px; color: #666; }
	</style>
</head>
<body>
	<div class="container">
		<div class="header">
			<h1>📤 Withdrawal Processed</h1>
		</div>
		<div class="content">
			<div class="info-box">
				<p class="amount">${{.amount}}</p>
				<p>Your withdrawal has been processed successfully.</p>
			</div>
			<p>The funds will arrive in your account within 1-3 business days depending on your payment method.</p>
			<p>If you have any questions, please contact support.</p>
		</div>
		<div class="footer">
			<p>RTX5 Trading Platform | This is an automated notification</p>
		</div>
	</div>
</body>
</html>`,

	TemplateLoginAlert: `
<!DOCTYPE html>
<html>
<head>
	<meta charset="UTF-8">
	<style>
		body { font-family: Arial, sans-serif; line-height: 1.6; color: #333; background: #f4f4f4; }
		.container { max-width: 600px; margin: 20px auto; background: white; border-radius: 8px; overflow: hidden; box-shadow: 0 2px 4px rgba(0,0,0,0.1); }
		.header { background: #FF9800; color: white; padding: 20px; text-align: center; }
		.content { padding: 30px; }
		.alert-box { background: #FFF3E0; border-left: 4px solid #FF9800; padding: 15px; margin: 20px 0; }
		.button { display: inline-block; padding: 12px 24px; background: #FF5722; color: white; text-decoration: none; border-radius: 5px; margin: 10px 0; }
		.footer { background: #f9f9f9; padding: 20px; text-align: center; font-size: 12px; color: #666; }
	</style>
</head>
<body>
	<div class="container">
		<div class="header">
			<h1>🔐 New Login Detected</h1>
		</div>
		<div class="content">
			<div class="alert-box">
				<p>A new login to your account was detected.</p>
				<p><strong>IP Address:</strong> {{.ip}}</p>
				<p><strong>Time:</strong> {{.time}}</p>
			</div>
			<p>If this was you, no action is needed.</p>
			<p>If you don't recognize this activity, please secure your account immediately:</p>
			<ul>
				<li>Change your password</li>
				<li>Enable two-factor authentication</li>
				<li>Contact support</li>
			</ul>
			<a href="{{.securityUrl}}" class="button">Secure Account</a>
		</div>
		<div class="footer">
			<p>RTX5 Trading Platform | This is an automated security notification</p>
		</div>
	</div>
</body>
</html>`,

	TemplatePasswordChanged: `
<!DOCTYPE html>
<html>
<head>
	<meta charset="UTF-8">
	<style>
		body { font-family: Arial, sans-serif; line-height: 1.6; color: #333; background: #f4f4f4; }
		.container { max-width: 600px; margin: 20px auto; background: white; border-radius: 8px; overflow: hidden; box-shadow: 0 2px 4px rgba(0,0,0,0.1); }
		.header { background: #4CAF50; color: white; padding: 20px; text-align: center; }
		.content { padding: 30px; }
		.success-box { background: #E8F5E9; border-left: 4px solid #4CAF50; padding: 15px; margin: 20px 0; }
		.button { display: inline-block; padding: 12px 24px; background: #FF5722; color: white; text-decoration: none; border-radius: 5px; margin: 10px 0; }
		.footer { background: #f9f9f9; padding: 20px; text-align: center; font-size: 12px; color: #666; }
	</style>
</head>
<body>
	<div class="container">
		<div class="header">
			<h1>✅ Password Changed</h1>
		</div>
		<div class="content">
			<div class="success-box">
				<p>Your password was changed successfully.</p>
			</div>
			<p>If you made this change, no further action is needed.</p>
			<p>If you did NOT change your password, please contact support immediately:</p>
			<a href="{{.supportUrl}}" class="button">Contact Support</a>
		</div>
		<div class="footer">
			<p>RTX5 Trading Platform | This is an automated security notification</p>
		</div>
	</div>
</body>
</html>`,

	TemplateReportGenerated: `
<!DOCTYPE html>
<html>
<head>
	<meta charset="UTF-8">
	<style>
		body { font-family: Arial, sans-serif; line-height: 1.6; color: #333; background: #f4f4f4; }
		.container { max-width: 600px; margin: 20px auto; background: white; border-radius: 8px; overflow: hidden; box-shadow: 0 2px 4px rgba(0,0,0,0.1); }
		.header { background: #4CAF50; color: white; padding: 20px; text-align: center; }
		.content { padding: 30px; }
		.report-box { background: #f9f9f9; padding: 20px; border-radius: 5px; margin: 20px 0; }
		.detail-row { display: flex; justify-content: space-between; padding: 8px 0; border-bottom: 1px solid #e0e0e0; }
		.detail-row:last-child { border-bottom: none; }
		.label { font-weight: bold; color: #555; }
		.value { color: #333; }
		.button { display: inline-block; padding: 12px 24px; background: #4CAF50; color: white; text-decoration: none; border-radius: 5px; margin: 10px 0; }
		.footer { background: #f9f9f9; padding: 20px; text-align: center; font-size: 12px; color: #666; }
	</style>
</head>
<body>
	<div class="container">
		<div class="header">
			<h1>📊 Automated Report Generated</h1>
		</div>
		<div class="content">
			<p>A new automated report has been generated and is ready for review.</p>
			<div class="report-box">
				<div class="detail-row">
					<span class="label">Report Title:</span>
					<span class="value">{{.reportTitle}}</span>
				</div>
				<div class="detail-row">
					<span class="label">Report Type:</span>
					<span class="value">{{.reportType}}</span>
				</div>
				<div class="detail-row">
					<span class="label">Period:</span>
					<span class="value">{{.reportPeriod}}</span>
				</div>
				<div class="detail-row">
					<span class="label">Generated At:</span>
					<span class="value">{{.generatedAt}}</span>
				</div>
				{{if .totalTrades}}
				<div class="detail-row">
					<span class="label">Total Trades:</span>
					<span class="value">{{.totalTrades}}</span>
				</div>
				{{end}}
				{{if .totalVolume}}
				<div class="detail-row">
					<span class="label">Total Volume:</span>
					<span class="value">{{.totalVolume}}</span>
				</div>
				{{end}}
				{{if .totalPnL}}
				<div class="detail-row">
					<span class="label">Total P&L:</span>
					<span class="value">{{.totalPnL}}</span>
				</div>
				{{end}}
				{{if .totalRevenue}}
				<div class="detail-row">
					<span class="label">Total Revenue:</span>
					<span class="value">{{.totalRevenue}}</span>
				</div>
				{{end}}
				{{if .avgMarginUtilization}}
				<div class="detail-row">
					<span class="label">Avg Margin Utilization:</span>
					<span class="value">{{.avgMarginUtilization}}%</span>
				</div>
				{{end}}
			</div>
			<p>You can view the full report in the admin panel or download it as CSV.</p>
			<a href="{{.reportUrl}}" class="button">View Report</a>
		</div>
		<div class="footer">
			<p>RTX5 Trading Platform | Automated Reporting System</p>
		</div>
	</div>
</body>
</html>`,
}

// RenderTemplate renders an email template with the provided data
func RenderTemplate(templateName string, data map[string]interface{}) (string, error) {
	// Get template string
	templateStr, exists := emailTemplates[templateName]
	if !exists {
		return "", fmt.Errorf("template not found: %s", templateName)
	}

	// Parse template
	tmpl, err := template.New(templateName).Parse(templateStr)
	if err != nil {
		return "", fmt.Errorf("failed to parse template: %w", err)
	}

	// Render template
	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return "", fmt.Errorf("failed to execute template: %w", err)
	}

	return buf.String(), nil
}

// GetTemplateSubject returns the default subject for a template
func GetTemplateSubject(templateName string) string {
	subjects := map[string]string{
		TemplateMarginWarning:       "⚠️ Margin Warning - Action Required",
		TemplateOrderFilled:         "✅ Order Filled",
		TemplateDepositConfirmed:    "💰 Deposit Confirmed",
		TemplateWithdrawalProcessed: "📤 Withdrawal Processed",
		TemplateLoginAlert:          "🔐 New Login Detected",
		TemplatePasswordChanged:     "✅ Password Changed",
		TemplateReportGenerated:     "📊 Automated Report Generated",
	}

	subject, exists := subjects[templateName]
	if !exists {
		return "Notification from RTX5"
	}
	return subject
}

// GetPlainTextFromHTML converts HTML email to plain text (simplified)
func GetPlainTextFromHTML(html string) string {
	// Remove HTML tags and format for plain text
	text := html

	// Convert common HTML elements to plain text
	text = strings.ReplaceAll(text, "</p>", "\n\n")
	text = strings.ReplaceAll(text, "<br>", "\n")
	text = strings.ReplaceAll(text, "<br/>", "\n")
	text = strings.ReplaceAll(text, "</li>", "\n")
	text = strings.ReplaceAll(text, "</div>", "\n")

	// Strip all HTML tags
	inTag := false
	var result strings.Builder
	for _, r := range text {
		if r == '<' {
			inTag = true
			continue
		}
		if r == '>' {
			inTag = false
			continue
		}
		if !inTag {
			result.WriteRune(r)
		}
	}

	// Clean up whitespace
	cleaned := strings.TrimSpace(result.String())

	// Remove excessive newlines
	for strings.Contains(cleaned, "\n\n\n") {
		cleaned = strings.ReplaceAll(cleaned, "\n\n\n", "\n\n")
	}

	return cleaned
}
