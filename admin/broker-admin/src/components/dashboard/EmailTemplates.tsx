/**
 * Email Template Management
 * Editor for broker communication email templates
 */

'use client';

import React, { useState, useMemo } from 'react';
import {
  Mail,
  Save,
  RotateCcw,
  Copy,
  Trash2,
  Eye,
  EyeOff,
  Code,
  Type,
  Clock,
} from 'lucide-react';

type TemplateType =
  | 'welcome'
  | 'kyc_approved'
  | 'kyc_rejected'
  | 'deposit_confirmed'
  | 'withdrawal_approved'
  | 'margin_call_warning'
  | 'password_reset'
  | 'twofa_enabled'
  | 'account_suspended';

interface EmailTemplate {
  id: TemplateType;
  name: string;
  subject: string;
  body: string;
  lastModified: Date;
  variables: string[];
}

const TEMPLATE_VARIABLES = [
  { key: '{{client_name}}', description: 'Client full name' },
  { key: '{{account_id}}', description: 'Trading account number' },
  { key: '{{amount}}', description: 'Transaction amount' },
  { key: '{{currency}}', description: 'Currency code' },
  { key: '{{date}}', description: 'Current date/time' },
  { key: '{{support_email}}', description: 'Support email address' },
  { key: '{{broker_name}}', description: 'Broker company name' },
];

const INITIAL_TEMPLATES: EmailTemplate[] = [
  {
    id: 'welcome',
    name: 'Welcome Email',
    subject: 'Welcome to {{broker_name}} - Your Trading Account is Ready!',
    body: `Dear {{client_name}},

Welcome to {{broker_name}}! We're thrilled to have you join our trading community.

Your trading account #{{account_id}} has been successfully created and is now active. You can start trading immediately by logging into your account.

<strong>Getting Started:</strong>
1. Fund your account - Minimum deposit: 100 {{currency}}
2. Download our trading platform
3. Explore our educational resources
4. Contact our support team anytime at {{support_email}}

We're committed to providing you with the best trading experience. If you have any questions, our support team is available 24/7.

Happy Trading!
The {{broker_name}} Team

<small>This email was sent on {{date}}</small>`,
    lastModified: new Date('2026-02-01'),
    variables: ['client_name', 'broker_name', 'account_id', 'currency', 'support_email', 'date'],
  },
  {
    id: 'kyc_approved',
    name: 'KYC Approved',
    subject: 'KYC Verification Approved - Account #{{account_id}}',
    body: `Dear {{client_name}},

Great news! Your identity verification has been <strong>successfully approved</strong>.

<strong>Account Details:</strong>
- Account Number: {{account_id}}
- Verification Status: Approved
- Approval Date: {{date}}

Your account is now fully verified and you have access to all trading features, including:
- Higher deposit/withdrawal limits
- Access to all trading instruments
- Priority customer support
- Exclusive promotions and bonuses

You can now deposit funds and start trading without restrictions.

If you have any questions, please contact us at {{support_email}}.

Best regards,
{{broker_name}} Compliance Team`,
    lastModified: new Date('2026-01-28'),
    variables: ['client_name', 'account_id', 'date', 'support_email', 'broker_name'],
  },
  {
    id: 'kyc_rejected',
    name: 'KYC Rejected',
    subject: 'Additional Documents Required - Account #{{account_id}}',
    body: `Dear {{client_name}},

Thank you for submitting your verification documents for account #{{account_id}}.

Unfortunately, we were unable to verify your identity with the documents provided. This may be due to:
- Document quality issues (blurry, cropped, or unclear)
- Expired identification documents
- Incomplete information
- Document type not accepted

<strong>Next Steps:</strong>
Please resubmit your documents ensuring they meet our requirements:
1. Clear, color images or scans
2. All four corners visible
3. Documents must be valid (not expired)
4. Personal information must be clearly visible

You can upload new documents directly in your account dashboard.

If you need assistance or have questions, please contact our support team at {{support_email}}.

Best regards,
{{broker_name}} Compliance Team

<small>Request Date: {{date}}</small>`,
    lastModified: new Date('2026-01-28'),
    variables: ['client_name', 'account_id', 'support_email', 'broker_name', 'date'],
  },
  {
    id: 'deposit_confirmed',
    name: 'Deposit Confirmed',
    subject: 'Deposit Confirmation - {{amount}} {{currency}}',
    body: `Dear {{client_name}},

Your deposit has been <strong>successfully processed</strong> and credited to your account.

<strong>Transaction Details:</strong>
- Amount: {{amount}} {{currency}}
- Account: #{{account_id}}
- Date: {{date}}
- Status: Confirmed

Your new account balance has been updated and is available for trading immediately.

<strong>What's Next?</strong>
- Start trading on our platform
- Explore our trading instruments
- Set up risk management tools
- Join our trading community

Thank you for choosing {{broker_name}}. Happy trading!

Questions? Contact us anytime at {{support_email}}.

Best regards,
{{broker_name}} Finance Team`,
    lastModified: new Date('2026-02-05'),
    variables: ['client_name', 'amount', 'currency', 'account_id', 'date', 'broker_name', 'support_email'],
  },
  {
    id: 'withdrawal_approved',
    name: 'Withdrawal Approved',
    subject: 'Withdrawal Approved - {{amount}} {{currency}}',
    body: `Dear {{client_name}},

Your withdrawal request has been approved and is being processed.

<strong>Withdrawal Details:</strong>
- Amount: {{amount}} {{currency}}
- Account: #{{account_id}}
- Approval Date: {{date}}
- Estimated Processing: 1-3 business days

The funds will be transferred to your registered payment method. You will receive a confirmation email once the transfer is complete.

<strong>Important Notes:</strong>
- Processing times may vary depending on your payment method
- Ensure your payment details are up to date
- Contact us if you don't receive funds within the expected timeframe

If you have any questions about this withdrawal, please contact our finance team at {{support_email}}.

Best regards,
{{broker_name}} Finance Team

<small>Reference: WD-{{account_id}}-{{date}}</small>`,
    lastModified: new Date('2026-02-03'),
    variables: ['client_name', 'amount', 'currency', 'account_id', 'date', 'support_email', 'broker_name'],
  },
  {
    id: 'margin_call_warning',
    name: 'Margin Call Warning',
    subject: 'URGENT: Margin Call Warning - Account #{{account_id}}',
    body: `<strong style="color: #ef4444;">URGENT: MARGIN CALL WARNING</strong>

Dear {{client_name}},

Your trading account #{{account_id}} has reached critical margin levels.

<strong>Current Status:</strong>
- Margin Level: Below 50%
- Risk Status: <span style="color: #ef4444;">HIGH RISK</span>
- Date: {{date}}

<strong style="color: #ef4444;">IMMEDIATE ACTION REQUIRED:</strong>
To avoid automatic position closure, you must take one of the following actions:

1. <strong>Deposit Additional Funds</strong>
   - Add funds to increase your margin level
   - Recommended amount: {{amount}} {{currency}}

2. <strong>Close Some Positions</strong>
   - Reduce your open positions to free up margin
   - Focus on positions with the highest margin requirements

3. <strong>Reduce Position Sizes</strong>
   - Scale down your trading volume

<strong>What Happens if You Don't Act?</strong>
If your margin level drops below 30%, our system will automatically close your positions to protect your account from further losses (Stop Out).

<strong>Need Help?</strong>
Our support team is available 24/7 to assist you:
- Email: {{support_email}}
- Live Chat: Available in your trading platform
- Phone: Check your account dashboard for contact numbers

Please address this matter immediately to avoid forced position closures.

Best regards,
{{broker_name}} Risk Management Team

<small style="color: #666;">This is an automated alert sent on {{date}}</small>`,
    lastModified: new Date('2026-02-08'),
    variables: ['client_name', 'account_id', 'date', 'amount', 'currency', 'support_email', 'broker_name'],
  },
  {
    id: 'password_reset',
    name: 'Password Reset',
    subject: 'Password Reset Request - {{broker_name}}',
    body: `Dear {{client_name}},

We received a request to reset the password for your {{broker_name}} trading account.

<strong>Account Information:</strong>
- Account: #{{account_id}}
- Request Date: {{date}}

If you made this request, please use the link below to create a new password:

<a href="#" style="display: inline-block; padding: 12px 24px; background-color: #3B82F6; color: white; text-decoration: none; border-radius: 4px; margin: 16px 0;">Reset My Password</a>

This link will expire in 24 hours for security reasons.

<strong style="color: #ef4444;">Didn't request a password reset?</strong>
If you didn't make this request, please ignore this email. Your password will remain unchanged and your account is secure.

For your security, we recommend:
- Using a strong, unique password
- Enabling two-factor authentication (2FA)
- Never sharing your password with anyone
- Regularly updating your password

If you have concerns about your account security, please contact us immediately at {{support_email}}.

Best regards,
{{broker_name}} Security Team

<small>For security reasons, this link expires on {{date}}</small>`,
    lastModified: new Date('2026-02-04'),
    variables: ['client_name', 'broker_name', 'account_id', 'date', 'support_email'],
  },
  {
    id: 'twofa_enabled',
    name: '2FA Enabled',
    subject: 'Two-Factor Authentication Enabled - Account #{{account_id}}',
    body: `Dear {{client_name}},

Two-Factor Authentication (2FA) has been <strong>successfully enabled</strong> on your account.

<strong>Security Update:</strong>
- Account: #{{account_id}}
- Security Feature: 2FA Enabled
- Activation Date: {{date}}

Your account is now protected with an additional layer of security. From now on, you'll need to provide a verification code from your authenticator app in addition to your password when logging in.

<strong>Important Reminders:</strong>
- Keep your authenticator app secure
- Save your backup codes in a safe place
- Never share your 2FA codes with anyone
- Contact support if you lose access to your authenticator

<strong>What if I lose my device?</strong>
If you lose access to your authenticator app, use your backup codes to log in. You can then disable and re-enable 2FA with a new device.

If you did NOT enable 2FA on your account, please contact our security team immediately at {{support_email}}.

Thank you for taking steps to secure your account.

Best regards,
{{broker_name}} Security Team

<small>This security update was made on {{date}}</small>`,
    lastModified: new Date('2026-01-30'),
    variables: ['client_name', 'account_id', 'date', 'support_email', 'broker_name'],
  },
  {
    id: 'account_suspended',
    name: 'Account Suspended',
    subject: 'Important: Account Suspension Notice - Account #{{account_id}}',
    body: `Dear {{client_name}},

We regret to inform you that your trading account #{{account_id}} has been <strong>temporarily suspended</strong>.

<strong>Suspension Details:</strong>
- Account: #{{account_id}}
- Status: Suspended
- Suspension Date: {{date}}

<strong>Reason for Suspension:</strong>
This action was taken due to [specific reason - compliance review, suspicious activity, document verification required, etc.].

<strong>What This Means:</strong>
While your account is suspended:
- You cannot place new trades
- Existing positions remain open (unless otherwise notified)
- You can still view your account balance and history
- Deposits and withdrawals are temporarily disabled

<strong>How to Resolve This:</strong>
To reactivate your account, please:
1. Contact our compliance team at {{support_email}}
2. Provide any requested documentation
3. Address the specific concerns outlined by our team

<strong>Need Immediate Assistance?</strong>
Our compliance team is here to help you resolve this matter quickly:
- Email: {{support_email}}
- Response Time: Within 24 hours

We take account security and regulatory compliance seriously. We appreciate your understanding and cooperation in resolving this matter.

Best regards,
{{broker_name}} Compliance Team

<small>Suspension Date: {{date}} | Account: #{{account_id}}</small>`,
    lastModified: new Date('2026-02-02'),
    variables: ['client_name', 'account_id', 'date', 'support_email', 'broker_name'],
  },
];

export default function EmailTemplates() {
  const [templates, setTemplates] = useState<EmailTemplate[]>(INITIAL_TEMPLATES);
  const [selectedTemplate, setSelectedTemplate] = useState<TemplateType>('welcome');
  const [editedSubject, setEditedSubject] = useState('');
  const [editedBody, setEditedBody] = useState('');
  const [showPreview, setShowPreview] = useState(false);
  const [hasChanges, setHasChanges] = useState(false);

  const currentTemplate = useMemo(
    () => templates.find((t) => t.id === selectedTemplate),
    [templates, selectedTemplate]
  );

  // Initialize editor when template changes
  React.useEffect(() => {
    if (currentTemplate) {
      setEditedSubject(currentTemplate.subject);
      setEditedBody(currentTemplate.body);
      setHasChanges(false);
    }
  }, [currentTemplate]);

  const handleSubjectChange = (value: string) => {
    setEditedSubject(value);
    setHasChanges(true);
  };

  const handleBodyChange = (value: string) => {
    setEditedBody(value);
    setHasChanges(true);
  };

  const handleSave = () => {
    setTemplates((prev) =>
      prev.map((t) =>
        t.id === selectedTemplate
          ? { ...t, subject: editedSubject, body: editedBody, lastModified: new Date() }
          : t
      )
    );
    setHasChanges(false);
  };

  const handleReset = () => {
    if (currentTemplate) {
      setEditedSubject(currentTemplate.subject);
      setEditedBody(currentTemplate.body);
      setHasChanges(false);
    }
  };

  const handleDuplicate = () => {
    alert('Duplicate template feature - would create a new custom template based on this one');
  };

  const handleDelete = () => {
    if (confirm(`Are you sure you want to delete the "${currentTemplate?.name}" template?`)) {
      alert('Delete template feature - would remove this template');
    }
  };

  const insertVariable = (variable: string) => {
    setEditedBody((prev) => prev + ' ' + variable);
    setHasChanges(true);
  };

  // Render preview with sample data
  const renderPreview = (text: string) => {
    return text
      .replace(/{{client_name}}/g, 'John Smith')
      .replace(/{{account_id}}/g, '1234567')
      .replace(/{{amount}}/g, '1,000.00')
      .replace(/{{currency}}/g, 'USD')
      .replace(/{{date}}/g, new Date().toLocaleString())
      .replace(/{{support_email}}/g, 'support@broker.com')
      .replace(/{{broker_name}}/g, 'RTX5 Trading');
  };

  return (
    <div className="flex h-full bg-[#1e1e1e] text-zinc-300">
      {/* Template List Sidebar */}
      <div className="w-64 border-r border-zinc-700 bg-[#252528] flex flex-col">
        <div className="p-4 border-b border-zinc-700">
          <h3 className="text-sm font-semibold text-zinc-200 flex items-center gap-2">
            <Mail size={16} />
            Email Templates
          </h3>
        </div>
        <div className="flex-1 overflow-y-auto custom-scrollbar">
          {templates.map((template) => (
            <button
              key={template.id}
              onClick={() => setSelectedTemplate(template.id)}
              className={`w-full px-4 py-3 text-left border-b border-zinc-800 hover:bg-zinc-700/30 transition-colors ${
                selectedTemplate === template.id ? 'bg-zinc-700/50 border-l-2 border-l-blue-500' : ''
              }`}
            >
              <div className="text-sm font-medium text-zinc-200">{template.name}</div>
              <div className="text-xs text-zinc-500 mt-1 flex items-center gap-1">
                <Clock size={10} />
                {template.lastModified.toLocaleDateString()}
              </div>
            </button>
          ))}
        </div>
      </div>

      {/* Template Editor */}
      <div className="flex-1 flex flex-col overflow-hidden">
        {/* Header with Actions */}
        <div className="px-4 py-3 border-b border-zinc-700 bg-[#252528] flex items-center justify-between">
          <div>
            <h2 className="text-lg font-semibold text-zinc-200">{currentTemplate?.name}</h2>
            <div className="text-xs text-zinc-500 mt-0.5">
              Last modified: {currentTemplate?.lastModified.toLocaleString()}
            </div>
          </div>
          <div className="flex items-center gap-2">
            <button
              onClick={() => setShowPreview(!showPreview)}
              className="px-3 py-1.5 bg-zinc-700 hover:bg-zinc-600 text-zinc-300 rounded text-sm flex items-center gap-2 transition-colors"
            >
              {showPreview ? <EyeOff size={14} /> : <Eye size={14} />}
              {showPreview ? 'Edit' : 'Preview'}
            </button>
            <button
              onClick={handleReset}
              disabled={!hasChanges}
              className="px-3 py-1.5 bg-zinc-700 hover:bg-zinc-600 disabled:opacity-50 disabled:cursor-not-allowed text-zinc-300 rounded text-sm flex items-center gap-2 transition-colors"
            >
              <RotateCcw size={14} />
              Reset
            </button>
            <button
              onClick={handleSave}
              disabled={!hasChanges}
              className="px-3 py-1.5 bg-blue-600 hover:bg-blue-700 disabled:opacity-50 disabled:cursor-not-allowed text-white rounded text-sm flex items-center gap-2 transition-colors"
            >
              <Save size={14} />
              Save
            </button>
            <button
              onClick={handleDuplicate}
              className="px-3 py-1.5 bg-zinc-700 hover:bg-zinc-600 text-zinc-300 rounded text-sm flex items-center gap-2 transition-colors"
            >
              <Copy size={14} />
              Duplicate
            </button>
            <button
              onClick={handleDelete}
              className="px-3 py-1.5 bg-rose-600/20 hover:bg-rose-600/30 text-rose-400 border border-rose-600/30 rounded text-sm flex items-center gap-2 transition-colors"
            >
              <Trash2 size={14} />
              Delete
            </button>
          </div>
        </div>

        {showPreview ? (
          /* Preview Pane */
          <div className="flex-1 overflow-y-auto custom-scrollbar p-6">
            <div className="max-w-3xl mx-auto bg-white text-gray-900 rounded-lg shadow-lg p-8">
              <div className="border-b border-gray-200 pb-4 mb-4">
                <div className="text-sm text-gray-600 mb-2">Subject:</div>
                <div className="text-xl font-semibold">{renderPreview(editedSubject)}</div>
              </div>
              <div
                className="prose prose-sm max-w-none"
                dangerouslySetInnerHTML={{
                  __html: renderPreview(editedBody).replace(/\n/g, '<br />'),
                }}
              />
            </div>
          </div>
        ) : (
          /* Editor Pane */
          <div className="flex-1 flex flex-col overflow-hidden">
            {/* Variable Toolbar */}
            <div className="px-4 py-2 border-b border-zinc-700 bg-[#1e1e1e]">
              <div className="text-xs font-semibold text-zinc-400 mb-2 uppercase tracking-wide">
                Insert Variable:
              </div>
              <div className="flex flex-wrap gap-2">
                {TEMPLATE_VARIABLES.map((variable) => (
                  <button
                    key={variable.key}
                    onClick={() => insertVariable(variable.key)}
                    className="px-2 py-1 bg-zinc-700 hover:bg-zinc-600 text-zinc-300 rounded text-xs transition-colors"
                    title={variable.description}
                  >
                    <Code size={12} className="inline mr-1" />
                    {variable.key}
                  </button>
                ))}
              </div>
            </div>

            {/* Subject Line Editor */}
            <div className="px-4 py-3 border-b border-zinc-700">
              <label className="block text-xs font-semibold text-zinc-400 mb-2 uppercase tracking-wide">
                <Type size={12} className="inline mr-1" />
                Subject Line:
              </label>
              <input
                type="text"
                value={editedSubject}
                onChange={(e) => handleSubjectChange(e.target.value)}
                className="w-full px-3 py-2 bg-[#252528] border border-zinc-700 rounded text-sm text-zinc-300 focus:outline-none focus:border-blue-500"
                placeholder="Email subject line..."
              />
            </div>

            {/* Body Editor */}
            <div className="flex-1 p-4 overflow-hidden flex flex-col">
              <label className="block text-xs font-semibold text-zinc-400 mb-2 uppercase tracking-wide">
                <Mail size={12} className="inline mr-1" />
                Email Body:
              </label>
              <textarea
                value={editedBody}
                onChange={(e) => handleBodyChange(e.target.value)}
                className="flex-1 w-full px-3 py-2 bg-[#252528] border border-zinc-700 rounded text-sm text-zinc-300 font-mono leading-relaxed focus:outline-none focus:border-blue-500 resize-none custom-scrollbar"
                placeholder="Email body content... Use HTML tags for formatting."
              />
              <div className="text-xs text-zinc-500 mt-2">
                Tip: You can use HTML tags like &lt;strong&gt;, &lt;a href=""&gt;, &lt;br /&gt;, etc. for formatting
              </div>
            </div>
          </div>
        )}
      </div>
    </div>
  );
}
