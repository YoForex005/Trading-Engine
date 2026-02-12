// Centralized API configuration
// All API URLs, WebSocket endpoints, and configuration should be defined here

const API_BASE_URL = process.env.NEXT_PUBLIC_API_BASE_URL || 'http://localhost:7999';
const WS_BASE_URL = process.env.NEXT_PUBLIC_WS_BASE_URL || 'ws://localhost:7999';

export const API_CONFIG = {
    // Base URLs
    BASE_URL: API_BASE_URL,
    WS_BASE_URL: WS_BASE_URL,

    // HTTP API endpoints
    ADMIN_API_URL: `${API_BASE_URL}/api`,
    MARKET_DATA_URL: `${API_BASE_URL}/api`,

    // WebSocket endpoints
    MARKET_WS_URL: `${WS_BASE_URL}/ws`,
    ADMIN_WS_URL: `${WS_BASE_URL}/ws/admin`,

    // Specific endpoints
    SYMBOLS: `${API_BASE_URL}/api/symbols`,
    ACCOUNTS: `${API_BASE_URL}/api/accounts`,
    ORDERS: `${API_BASE_URL}/api/orders`,
    HISTORY: `${API_BASE_URL}/api/history`,

    // Auth endpoints
    AUTH_LOGIN: `${API_BASE_URL}/login`,
    AUTH_LOGOUT: `${API_BASE_URL}/logout`,
    AUTH_VERIFY: `${API_BASE_URL}/verify`,

    // Admin endpoints
    ADMIN_ACCOUNTS: `${API_BASE_URL}/api/admin/accounts`,
    ADMIN_ORDERS: `${API_BASE_URL}/api/admin/orders`,
    ADMIN_GROUPS: `${API_BASE_URL}/api/admin/groups`,

    // Admin user management endpoints
    ADMIN_USERS: `${API_BASE_URL}/admin/users`,
    ADMIN_USER_BY_ID: (id: string) => `${API_BASE_URL}/admin/users/${id}`,
    ADMIN_USER_DISABLE: (id: string) => `${API_BASE_URL}/admin/users/${id}/disable`,
    ADMIN_USER_ENABLE: (id: string) => `${API_BASE_URL}/admin/users/${id}/enable`,

    // Risk & Analytics endpoints
    ANALYTICS_EXPOSURE_CURRENT: `${API_BASE_URL}/api/analytics/exposure/current`,
    ANALYTICS_EXPOSURE_HEATMAP: `${API_BASE_URL}/api/analytics/exposure/heatmap`,
    ADMIN_POSITIONS: `${API_BASE_URL}/admin/positions`,

    // Chart data endpoints
    OHLC_DATA: (symbol: string, timeframe: string) =>
        `${API_BASE_URL}/api/ohlc/${symbol}/${timeframe}`,
    TICK_DATA: (symbol: string, date: string) =>
        `${API_BASE_URL}/api/ticks/${symbol}/${date}`,

    // 2FA / TOTP endpoints
    TOTP_SETUP: `${API_BASE_URL}/admin/2fa/setup`,
    TOTP_VERIFY: `${API_BASE_URL}/admin/2fa/verify`,
    TOTP_ENABLE: `${API_BASE_URL}/admin/2fa/enable`,
    TOTP_DISABLE: `${API_BASE_URL}/admin/2fa/disable`,
    TOTP_BACKUP_CODES: `${API_BASE_URL}/admin/2fa/backup-codes`,
    TOTP_VALIDATE: `${API_BASE_URL}/admin/2fa/validate`,

    // RBAC endpoints
    ADMIN_ROLE: `${API_BASE_URL}/admin/role`,
    ADMIN_PERMISSIONS: `${API_BASE_URL}/admin/permissions`,

    // Export / Reports endpoints
    EXPORT_TRADES: `${API_BASE_URL}/admin/export/trades`,
    EXPORT_POSITIONS: `${API_BASE_URL}/admin/export/positions`,
    EXPORT_ACCOUNTS: `${API_BASE_URL}/admin/export/accounts`,
    EXPORT_LEDGER: `${API_BASE_URL}/admin/export/ledger`,

    // Dashboard / Analytics endpoints
    ANALYTICS_OVERVIEW: `${API_BASE_URL}/admin/analytics/overview`,
    NOTIFICATIONS: `${API_BASE_URL}/api/notifications`,

    // Symbol Management endpoints
    SYMBOLS_LIST: `${API_BASE_URL}/api/symbols`,
    SYMBOL_SPREAD: (name: string) => `${API_BASE_URL}/api/symbols/${name}/spread`,
    SYMBOL_UPDATE_SPREAD: (name: string) => `${API_BASE_URL}/admin/symbols/${name}/spread`,
    SYMBOL_CREATE: `${API_BASE_URL}/admin/symbols`,
    SYMBOL_UPDATE_STATUS: (name: string) => `${API_BASE_URL}/admin/symbols/${name}/status`,

    // Audit Log endpoints
    AUDIT_LOGS: `${API_BASE_URL}/admin/audit/logs`,
    AUDIT_LOG_BY_ID: (id: string) => `${API_BASE_URL}/admin/audit/logs/${id}`,
    AUDIT_STATS: `${API_BASE_URL}/admin/audit/stats`,

    // Withdrawal Approval endpoints
    ADMIN_WITHDRAWALS_PENDING: `${API_BASE_URL}/admin/withdrawals/pending`,
    ADMIN_WITHDRAWALS_HISTORY: `${API_BASE_URL}/admin/withdrawals/history`,
    ADMIN_WITHDRAWALS_APPROVE: (id: string) => `${API_BASE_URL}/admin/withdrawals/${id}/approve`,
    ADMIN_WITHDRAWALS_REJECT: (id: string) => `${API_BASE_URL}/admin/withdrawals/${id}/reject`,

    // User Management endpoints
    ADMIN_USER_RESET_PASSWORD: (id: string) => `${API_BASE_URL}/admin/users/${id}/reset-password`,

    // Client Management endpoints
    CLIENT_LIST: `${API_BASE_URL}/admin/clients`,
    CLIENT_BY_ID: (id: string) => `${API_BASE_URL}/admin/clients/${id}`,
    CLIENT_SUSPEND: (id: string) => `${API_BASE_URL}/admin/clients/${id}/suspend`,
    CLIENT_ACTIVATE: (id: string) => `${API_BASE_URL}/admin/clients/${id}/activate`,
    CLIENT_SEND_NOTIFICATION: (id: string) => `${API_BASE_URL}/admin/clients/${id}/notification`,

    // Group Management endpoints
    ADMIN_GROUP_BY_ID: (id: string) => `${API_BASE_URL}/admin/groups/${id}`,
    ADMIN_GROUP_CREATE: `${API_BASE_URL}/admin/groups`,
    ADMIN_GROUP_UPDATE: (id: string) => `${API_BASE_URL}/admin/groups/${id}`,
    ADMIN_GROUP_DELETE: (id: string) => `${API_BASE_URL}/admin/groups/${id}`,
    ADMIN_GROUP_ASSIGN_CLIENT: `${API_BASE_URL}/admin/groups/assign`,

    // LP Health Monitoring endpoints
    LP_HEALTH_STATUS: `${API_BASE_URL}/admin/lp/health`,
    LP_HEALTH_DETAILS: (lpId: string) => `${API_BASE_URL}/admin/lp/health/${lpId}`,
    LP_HEALTH_METRICS: `${API_BASE_URL}/admin/lp/metrics`,

    // API Key Management endpoints
    API_KEYS_LIST: `${API_BASE_URL}/admin/api-keys`,
    API_KEYS_CREATE: `${API_BASE_URL}/admin/api-keys`,
    API_KEYS_REVOKE: (keyId: string) => `${API_BASE_URL}/admin/api-keys/${keyId}/revoke`,

    // Webhook Management endpoints
    WEBHOOKS_LIST: `${API_BASE_URL}/admin/webhooks`,
    WEBHOOKS_CREATE: `${API_BASE_URL}/admin/webhooks`,
    WEBHOOKS_UPDATE: (webhookId: string) => `${API_BASE_URL}/admin/webhooks/${webhookId}`,
    WEBHOOKS_DELETE: (webhookId: string) => `${API_BASE_URL}/admin/webhooks/${webhookId}`,
    WEBHOOKS_TOGGLE: (webhookId: string) => `${API_BASE_URL}/admin/webhooks/${webhookId}/toggle`,
    WEBHOOKS_TEST: (webhookId: string) => `${API_BASE_URL}/admin/webhooks/${webhookId}/test`,
    WEBHOOKS_RETRY: (webhookId: string, logId: string) => `${API_BASE_URL}/admin/webhooks/${webhookId}/retry/${logId}`,

    // Commission Report endpoints
    COMMISSION_REPORT: `${API_BASE_URL}/admin/commissions/report`,
    COMMISSION_EXPORT: `${API_BASE_URL}/admin/commissions/export`,
    COMMISSION_BY_CLIENT: (clientId: string) => `${API_BASE_URL}/admin/commissions/client/${clientId}`,
    COMMISSION_SUMMARY: `${API_BASE_URL}/admin/commissions/summary`,

    // Broker P&L Report endpoints
    BROKER_PNL_SUMMARY: `${API_BASE_URL}/admin/reports/pnl/summary`,
    BROKER_PNL_BY_SYMBOL_GROUP: `${API_BASE_URL}/admin/reports/pnl/symbol-groups`,
    BROKER_PNL_BY_SYMBOL: `${API_BASE_URL}/admin/reports/pnl/symbols`,
    BROKER_PNL_BY_CLIENT: `${API_BASE_URL}/admin/reports/pnl/clients`,
    BROKER_PNL_MONTHLY_REVENUE: `${API_BASE_URL}/admin/reports/pnl/monthly-revenue`,
    BROKER_PNL_DAILY_TREND: `${API_BASE_URL}/admin/reports/pnl/daily-trend`,
    BROKER_PNL_EXPORT: `${API_BASE_URL}/admin/reports/pnl/export`,

    // Email Template Management endpoints
    EMAIL_TEMPLATES_LIST: `${API_BASE_URL}/admin/email-templates`,
    EMAIL_TEMPLATE_BY_ID: (id: string) => `${API_BASE_URL}/admin/email-templates/${id}`,
    EMAIL_TEMPLATE_CREATE: `${API_BASE_URL}/admin/email-templates`,
    EMAIL_TEMPLATE_UPDATE: (id: string) => `${API_BASE_URL}/admin/email-templates/${id}`,
    EMAIL_TEMPLATE_DELETE: (id: string) => `${API_BASE_URL}/admin/email-templates/${id}`,
    EMAIL_TEMPLATE_SEND_TEST: (id: string) => `${API_BASE_URL}/admin/email-templates/${id}/test`,
    EMAIL_TEMPLATE_STATS: (id: string) => `${API_BASE_URL}/admin/email-templates/${id}/stats`,

    // IP Filter Management endpoints
    IP_FILTER_LIST: `${API_BASE_URL}/admin/security/ip-filters`,
    IP_FILTER_CREATE: `${API_BASE_URL}/admin/security/ip-filters`,
    IP_FILTER_DELETE: (ruleId: string) => `${API_BASE_URL}/admin/security/ip-filters/${ruleId}`,
    IP_FILTER_CHECK: `${API_BASE_URL}/admin/security/ip-filters/check`,
    IP_FILTER_IMPORT: `${API_BASE_URL}/admin/security/ip-filters/import`,
    IP_FILTER_STATS: `${API_BASE_URL}/admin/security/ip-filters/stats`,

    // Trade Copier / Social Trading endpoints
    TRADE_COPIER_LIST: `${API_BASE_URL}/admin/trade-copier/relations`,
    TRADE_COPIER_CREATE: `${API_BASE_URL}/admin/trade-copier/relations`,
    TRADE_COPIER_UPDATE: (relationId: string) => `${API_BASE_URL}/admin/trade-copier/relations/${relationId}`,
    TRADE_COPIER_DELETE: (relationId: string) => `${API_BASE_URL}/admin/trade-copier/relations/${relationId}`,
    TRADE_COPIER_TOGGLE: (relationId: string) => `${API_BASE_URL}/admin/trade-copier/relations/${relationId}/toggle`,
    TRADE_COPIER_LOG: (relationId: string) => `${API_BASE_URL}/admin/trade-copier/relations/${relationId}/log`,
    TRADE_COPIER_STATS: `${API_BASE_URL}/admin/trade-copier/stats`,

    // KYC/AML Verification endpoints
    KYC_APPLICATIONS_LIST: `${API_BASE_URL}/admin/kyc/applications`,
    KYC_APPLICATION_BY_ID: (applicationId: string) => `${API_BASE_URL}/admin/kyc/applications/${applicationId}`,
    KYC_APPLICATION_APPROVE: (applicationId: string) => `${API_BASE_URL}/admin/kyc/applications/${applicationId}/approve`,
    KYC_APPLICATION_REJECT: (applicationId: string) => `${API_BASE_URL}/admin/kyc/applications/${applicationId}/reject`,
    KYC_RUN_AML_CHECK: (applicationId: string) => `${API_BASE_URL}/admin/kyc/applications/${applicationId}/aml-check`,
    KYC_BULK_AML_CHECK: `${API_BASE_URL}/admin/kyc/bulk-aml-check`,
    KYC_STATS: `${API_BASE_URL}/admin/kyc/stats`,
    KYC_DOCUMENT: (applicationId: string, documentType: string) => `${API_BASE_URL}/admin/kyc/applications/${applicationId}/documents/${documentType}`,

    // Compliance Dashboard endpoints
    COMPLIANCE_SCORE: `${API_BASE_URL}/admin/compliance/score`,
    COMPLIANCE_KYC_STATS: `${API_BASE_URL}/admin/compliance/kyc/stats`,
    COMPLIANCE_TRANSACTIONS: `${API_BASE_URL}/admin/compliance/transactions`,
    COMPLIANCE_TRANSACTION_REVIEW: (txnId: string) => `${API_BASE_URL}/admin/compliance/transactions/${txnId}/review`,
    COMPLIANCE_REPORTS: `${API_BASE_URL}/admin/compliance/regulatory-reports`,
    COMPLIANCE_REPORT_GENERATE: `${API_BASE_URL}/admin/compliance/regulatory-reports/generate`,
    COMPLIANCE_RISK_DISTRIBUTION: `${API_BASE_URL}/admin/compliance/risk/distribution`,
    COMPLIANCE_SEND_REMINDER: (clientId: string) => `${API_BASE_URL}/admin/compliance/kyc/${clientId}/reminder`,

    // Notification Center endpoints
    NOTIFICATIONS_LIST: `${API_BASE_URL}/admin/notifications`,
    NOTIFICATIONS_UNREAD_COUNT: `${API_BASE_URL}/admin/notifications/unread-count`,
    NOTIFICATIONS_MARK_READ: (notificationId: string) => `${API_BASE_URL}/admin/notifications/${notificationId}/read`,
    NOTIFICATIONS_MARK_ALL_READ: `${API_BASE_URL}/admin/notifications/mark-all-read`,
    NOTIFICATIONS_DELETE: (notificationId: string) => `${API_BASE_URL}/admin/notifications/${notificationId}`,
    NOTIFICATIONS_CLEAR_ALL: `${API_BASE_URL}/admin/notifications/clear-all`,
    NOTIFICATIONS_SETTINGS: `${API_BASE_URL}/admin/notifications/settings`,
    NOTIFICATIONS_STATS: `${API_BASE_URL}/admin/notifications/stats`,
    NOTIFICATIONS_RULES: `${API_BASE_URL}/admin/notifications/rules`,
    NOTIFICATIONS_RULE_UPDATE: (ruleId: string) => `${API_BASE_URL}/admin/notifications/rules/${ruleId}`,
    NOTIFICATIONS_SEND: `${API_BASE_URL}/admin/notifications/send`,
    NOTIFICATIONS_BROADCAST: `${API_BASE_URL}/admin/notifications/broadcast`,

    // Deposit Management endpoints
    DEPOSITS_LIST: `${API_BASE_URL}/admin/deposits`,
    DEPOSIT_BY_ID: (depositId: string) => `${API_BASE_URL}/admin/deposits/${depositId}`,
    DEPOSIT_PROCESS: (depositId: string) => `${API_BASE_URL}/admin/deposits/${depositId}/process`,
    DEPOSIT_REJECT: (depositId: string) => `${API_BASE_URL}/admin/deposits/${depositId}/reject`,
    DEPOSITS_STATS: `${API_BASE_URL}/admin/deposits/stats`,

    // Internal Transfer endpoints
    TRANSFERS_LIST: `${API_BASE_URL}/admin/transfers`,
    TRANSFER_BY_ID: (transferId: string) => `${API_BASE_URL}/admin/transfers/${transferId}`,
    TRANSFER_CREATE: `${API_BASE_URL}/admin/transfers`,
    TRANSFER_STATS: `${API_BASE_URL}/admin/transfers/stats`,

    // Financial Operations Dashboard endpoints
    FINANCIAL_SUMMARY: `${API_BASE_URL}/admin/financial/summary`,
    FINANCIAL_DAILY_FLOW: `${API_BASE_URL}/admin/financial/daily-flow`,

    // Server Configuration endpoints
    SERVER_CONFIG_GET: `${API_BASE_URL}/admin/server/config`,
    SERVER_CONFIG_UPDATE: `${API_BASE_URL}/admin/server/config`,
    SERVER_INFO: `${API_BASE_URL}/admin/server/info`,
    SERVER_CONFIG_HISTORY: `${API_BASE_URL}/admin/server/config/history`,
    SERVER_BACKUP: `${API_BASE_URL}/admin/server/backup`,
    SERVER_LOGS: `${API_BASE_URL}/admin/server/logs`,

    // Dealing Desk endpoints
    DEALING_LIVE_ORDERS: `${API_BASE_URL}/admin/dealing/live-orders`,
    DEALING_PENDING_ORDERS: `${API_BASE_URL}/admin/dealing/pending-orders`,
    DEALING_ACCEPT_ORDER: (orderId: string) => `${API_BASE_URL}/admin/dealing/orders/${orderId}/accept`,
    DEALING_REJECT_ORDER: (orderId: string) => `${API_BASE_URL}/admin/dealing/orders/${orderId}/reject`,
    DEALING_REQUOTE_ORDER: (orderId: string) => `${API_BASE_URL}/admin/dealing/orders/${orderId}/requote`,
    DEALING_OPEN_POSITIONS: `${API_BASE_URL}/admin/dealing/positions`,
    DEALING_CLOSE_POSITION: (ticket: string) => `${API_BASE_URL}/admin/dealing/positions/${ticket}/close`,
    DEALING_STATS: `${API_BASE_URL}/admin/dealing/stats`,
    DEALING_HALT_TRADING: `${API_BASE_URL}/admin/dealing/halt-trading`,
    DEALING_DISABLE_SYMBOL: (symbol: string) => `${API_BASE_URL}/admin/dealing/symbols/${symbol}/disable`,
    DEALING_CLOSE_ALL_CLIENT: (clientId: string) => `${API_BASE_URL}/admin/dealing/clients/${clientId}/close-all`,

    // Trade History endpoints
    TRADE_HISTORY_LIST: `${API_BASE_URL}/admin/trades/history`,
    TRADE_HISTORY_BY_ACCOUNT: (accountId: string) => `${API_BASE_URL}/admin/trades/history/${accountId}`,
    TRADE_HISTORY_EXPORT_CSV: `${API_BASE_URL}/admin/trades/export/csv`,
    TRADE_HISTORY_EXPORT_PDF: `${API_BASE_URL}/admin/trades/export/pdf`,
    ACCOUNT_STATEMENT: (accountId: string) => `${API_BASE_URL}/admin/accounts/${accountId}/statement`,

    // System Health / Monitoring endpoints
    SYSTEM_HEALTH_METRICS: `${API_BASE_URL}/admin/system/health/metrics`,
    SYSTEM_HEALTH_SERVICES: `${API_BASE_URL}/admin/system/health/services`,
    SYSTEM_HEALTH_ALERTS: `${API_BASE_URL}/admin/system/health/alerts`,
    SYSTEM_HEALTH_CHART_DATA: `${API_BASE_URL}/admin/system/health/chart-data`,

    // Affiliate / IB Management endpoints
    AFFILIATES_LIST: `${API_BASE_URL}/admin/affiliates`,
    AFFILIATE_BY_ID: (affiliateId: string) => `${API_BASE_URL}/admin/affiliates/${affiliateId}`,
    AFFILIATE_APPROVE: (affiliateId: string) => `${API_BASE_URL}/admin/affiliates/${affiliateId}/approve`,
    AFFILIATE_SUSPEND: (affiliateId: string) => `${API_BASE_URL}/admin/affiliates/${affiliateId}/suspend`,
    AFFILIATE_ADJUST_TIER: (affiliateId: string) => `${API_BASE_URL}/admin/affiliates/${affiliateId}/tier`,
    AFFILIATE_PROCESS_PAYOUT: (affiliateId: string) => `${API_BASE_URL}/admin/affiliates/${affiliateId}/payout`,
    AFFILIATE_REFERRED_CLIENTS: (affiliateId: string) => `${API_BASE_URL}/admin/affiliates/${affiliateId}/clients`,
    AFFILIATE_COMMISSIONS: (affiliateId: string) => `${API_BASE_URL}/admin/affiliates/${affiliateId}/commissions`,
    AFFILIATE_PAYOUTS: (affiliateId: string) => `${API_BASE_URL}/admin/affiliates/${affiliateId}/payouts`,
    AFFILIATE_STATS: `${API_BASE_URL}/admin/affiliates/stats`,

    // Risk Rules Engine endpoints
    RISK_RULES_LIST: `${API_BASE_URL}/admin/risk/rules`,
    RISK_RULE_CREATE: `${API_BASE_URL}/admin/risk/rules`,
    RISK_RULE_UPDATE: (ruleId: string) => `${API_BASE_URL}/admin/risk/rules/${ruleId}`,
    RISK_RULE_DELETE: (ruleId: string) => `${API_BASE_URL}/admin/risk/rules/${ruleId}`,
    RISK_RULE_TOGGLE: (ruleId: string) => `${API_BASE_URL}/admin/risk/rules/${ruleId}/toggle`,
    RISK_TRIGGER_LOGS: `${API_BASE_URL}/admin/risk/trigger-logs`,
    RISK_RULES_STATS: `${API_BASE_URL}/admin/risk/stats`,

    // Risk Dashboard endpoints (Task #169)
    RISK_OVERVIEW: `${API_BASE_URL}/admin/risk-dashboard/overview`,
    RISK_EXPOSURE_BY_GROUP: `${API_BASE_URL}/admin/risk-dashboard/exposure/by-group`,
    RISK_RISKY_ACCOUNTS: `${API_BASE_URL}/admin/risk-dashboard/risky-accounts`,
    RISK_ALERTS: `${API_BASE_URL}/admin/risk-dashboard/alerts`,
    RISK_HEATMAP: `${API_BASE_URL}/admin/risk-dashboard/heatmap`,
    RISK_TREND: `${API_BASE_URL}/admin/risk-dashboard/trend`,
    RISK_CIRCUIT_BREAKERS: `${API_BASE_URL}/admin/risk-dashboard/circuit-breakers`,
    RISK_LIMITS: `${API_BASE_URL}/admin/risk-dashboard/limits`,

    // Announcement Management endpoints
    ANNOUNCEMENTS_LIST: `${API_BASE_URL}/admin/announcements`,
    ANNOUNCEMENT_BY_ID: (announcementId: string) => `${API_BASE_URL}/admin/announcements/${announcementId}`,
    ANNOUNCEMENT_CREATE: `${API_BASE_URL}/admin/announcements`,
    ANNOUNCEMENT_UPDATE: (announcementId: string) => `${API_BASE_URL}/admin/announcements/${announcementId}`,
    ANNOUNCEMENT_DELETE: (announcementId: string) => `${API_BASE_URL}/admin/announcements/${announcementId}`,
    ANNOUNCEMENT_PREVIEW: (announcementId: string) => `${API_BASE_URL}/admin/announcements/${announcementId}/preview`,
    ANNOUNCEMENTS_STATS: `${API_BASE_URL}/admin/announcements/stats`,

    // Liquidity Provider Configuration endpoints
    LP_CONFIG_LIST: `${API_BASE_URL}/admin/lp/config`,
    LP_CONFIG_BY_ID: (lpId: string) => `${API_BASE_URL}/admin/lp/config/${lpId}`,
    LP_CONFIG_UPDATE: (lpId: string) => `${API_BASE_URL}/admin/lp/config/${lpId}`,
    LP_CONFIG_TEST_CONNECTION: (lpId: string) => `${API_BASE_URL}/admin/lp/config/${lpId}/test`,
    LP_ROUTING_CONFIG: `${API_BASE_URL}/admin/lp/routing`,
    LP_ROUTING_UPDATE: (symbol: string) => `${API_BASE_URL}/admin/lp/routing/${symbol}`,
    LP_SPREAD_COMPARISON: `${API_BASE_URL}/admin/lp/spreads/comparison`,
    LP_STATS: `${API_BASE_URL}/admin/lp/stats`,

    // Market Hours / Sessions Configuration endpoints
    MARKET_HOURS_LIST: `${API_BASE_URL}/admin/market-hours`,
    MARKET_HOURS_BY_SYMBOL: (symbol: string) => `${API_BASE_URL}/admin/market-hours/${symbol}`,
    MARKET_HOURS_UPDATE: (symbol: string) => `${API_BASE_URL}/admin/market-hours/${symbol}`,
    MARKET_HOURS_HOLIDAYS: (symbol: string) => `${API_BASE_URL}/admin/market-hours/${symbol}/holidays`,
    MARKET_HOURS_ADD_HOLIDAY: (symbol: string) => `${API_BASE_URL}/admin/market-hours/${symbol}/holidays`,
    MARKET_HOURS_DELETE_HOLIDAY: (symbol: string, holidayId: string) => `${API_BASE_URL}/admin/market-hours/${symbol}/holidays/${holidayId}`,
    MARKET_HOURS_CURRENT_STATUS: `${API_BASE_URL}/admin/market-hours/status`,

    // Execution Report / Fill Analysis endpoints
    EXECUTION_REPORT_LIST: `${API_BASE_URL}/admin/reports/executions`,
    EXECUTION_REPORT_BY_SYMBOL: (symbol: string) => `${API_BASE_URL}/admin/reports/executions/symbol/${symbol}`,
    EXECUTION_REPORT_BY_LP: (lpId: string) => `${API_BASE_URL}/admin/reports/executions/lp/${lpId}`,
    EXECUTION_REPORT_QUALITY_METRICS: `${API_BASE_URL}/admin/reports/executions/quality`,
    EXECUTION_REPORT_HOURLY_HEATMAP: `${API_BASE_URL}/admin/reports/executions/hourly`,
    EXECUTION_REPORT_EXPORT: `${API_BASE_URL}/admin/reports/executions/export`,

    // Payment Gateway Configuration endpoints
    PAYMENT_GATEWAYS_LIST: `${API_BASE_URL}/admin/payment/gateways`,
    PAYMENT_GATEWAY_CONFIG: (gatewayId: string) => `${API_BASE_URL}/admin/payment/gateways/${gatewayId}/config`,
    PAYMENT_GATEWAY_UPDATE: (gatewayId: string) => `${API_BASE_URL}/admin/payment/gateways/${gatewayId}`,
    PAYMENT_GATEWAY_TOGGLE: (gatewayId: string) => `${API_BASE_URL}/admin/payment/gateways/${gatewayId}/toggle`,
    PAYMENT_GATEWAY_TEST: (gatewayId: string) => `${API_BASE_URL}/admin/payment/gateways/${gatewayId}/test`,
    PAYMENT_LIMITS_LIST: `${API_BASE_URL}/admin/payment/limits`,
    PAYMENT_LIMITS_UPDATE: (methodId: string) => `${API_BASE_URL}/admin/payment/limits/${methodId}`,
    PAYMENT_FEES_LIST: `${API_BASE_URL}/admin/payment/fees`,
    PAYMENT_FEES_UPDATE: (methodId: string) => `${API_BASE_URL}/admin/payment/fees/${methodId}`,
    PAYMENT_PENDING_TRANSACTIONS: `${API_BASE_URL}/admin/payment/transactions/pending`,
    PAYMENT_TRANSACTION_PROCESS: (txnId: string) => `${API_BASE_URL}/admin/payment/transactions/${txnId}/process`,
    PAYMENT_STATS: `${API_BASE_URL}/admin/payment/stats`,

    // Swap Rate Configuration endpoints
    SWAP_CONFIG_LIST: `${API_BASE_URL}/admin/swaps`,
    SWAP_CONFIG_BY_SYMBOL: (symbol: string) => `${API_BASE_URL}/admin/swaps/${symbol}`,
    SWAP_CONFIG_UPDATE: (symbol: string) => `${API_BASE_URL}/admin/swaps/${symbol}`,
    SWAP_CONFIG_BULK_UPDATE: `${API_BASE_URL}/admin/swaps/bulk-update`,
    SWAP_HISTORY: `${API_BASE_URL}/admin/swaps/history`,
    SWAP_STATS: `${API_BASE_URL}/admin/swaps/stats`,

    // Support Ticket System endpoints
    SUPPORT_TICKETS_LIST: `${API_BASE_URL}/admin/support/tickets`,
    SUPPORT_TICKET_BY_ID: (ticketId: string) => `${API_BASE_URL}/admin/support/tickets/${ticketId}`,
    SUPPORT_TICKET_CREATE: `${API_BASE_URL}/admin/support/tickets`,
    SUPPORT_TICKET_UPDATE: (ticketId: string) => `${API_BASE_URL}/admin/support/tickets/${ticketId}`,
    SUPPORT_TICKET_REPLY: (ticketId: string) => `${API_BASE_URL}/admin/support/tickets/${ticketId}/reply`,
    SUPPORT_TICKET_ASSIGN: (ticketId: string) => `${API_BASE_URL}/admin/support/tickets/${ticketId}/assign`,
    SUPPORT_TICKET_CLOSE: (ticketId: string) => `${API_BASE_URL}/admin/support/tickets/${ticketId}/close`,
    SUPPORT_CANNED_RESPONSES: `${API_BASE_URL}/admin/support/canned-responses`,
    SUPPORT_STATS: `${API_BASE_URL}/admin/support/stats`,

    // Communication Log endpoints
    COMMUNICATION_LOG_LIST: `${API_BASE_URL}/admin/communications`,
    COMMUNICATION_LOG_BY_ID: (commId: string) => `${API_BASE_URL}/admin/communications/${commId}`,
    COMMUNICATION_SEND: `${API_BASE_URL}/admin/communications/send`,
    COMMUNICATION_TEMPLATES: `${API_BASE_URL}/admin/communications/templates`,
    COMMUNICATION_TEMPLATE_BY_ID: (templateId: string) => `${API_BASE_URL}/admin/communications/templates/${templateId}`,
    COMMUNICATION_STATS: `${API_BASE_URL}/admin/communications/stats`,
    COMMUNICATION_DELIVERY_STATUS: (commId: string) => `${API_BASE_URL}/admin/communications/${commId}/status`,

    // Localization / Multi-Language endpoints
    LOCALIZATION_LANGUAGES_LIST: `${API_BASE_URL}/admin/localization/languages`,
    LOCALIZATION_LANGUAGE_TOGGLE: (languageCode: string) => `${API_BASE_URL}/admin/localization/languages/${languageCode}/toggle`,
    LOCALIZATION_DEFAULT_LANGUAGE: `${API_BASE_URL}/admin/localization/default-language`,
    LOCALIZATION_TRANSLATIONS_LIST: `${API_BASE_URL}/admin/localization/translations`,
    LOCALIZATION_TRANSLATION_UPDATE: (key: string) => `${API_BASE_URL}/admin/localization/translations/${key}`,
    LOCALIZATION_EXPORT: `${API_BASE_URL}/admin/localization/export`,
    LOCALIZATION_IMPORT: `${API_BASE_URL}/admin/localization/import`,
    LOCALIZATION_LOCALE_FORMATS: `${API_BASE_URL}/admin/localization/locale-formats`,
    LOCALIZATION_LOCALE_FORMAT_UPDATE: (languageCode: string) => `${API_BASE_URL}/admin/localization/locale-formats/${languageCode}`,

    // Regulatory Reporting endpoints
    REGULATORY_REPORTS_LIST: `${API_BASE_URL}/admin/regulatory/reports`,
    REGULATORY_REPORT_GENERATE: `${API_BASE_URL}/admin/regulatory/reports/generate`,
    REGULATORY_REPORT_BY_ID: (reportId: string) => `${API_BASE_URL}/admin/regulatory/reports/${reportId}`,
    REGULATORY_REPORT_DOWNLOAD: (reportId: string) => `${API_BASE_URL}/admin/regulatory/reports/${reportId}/download`,
    REGULATORY_REPORT_SUBMIT: (reportId: string) => `${API_BASE_URL}/admin/regulatory/reports/${reportId}/submit`,
    REGULATORY_REPORT_STATS: `${API_BASE_URL}/admin/regulatory/stats`,
    REGULATORY_DEADLINES: `${API_BASE_URL}/admin/regulatory/deadlines`,
    REGULATORY_TIMELINE: `${API_BASE_URL}/admin/regulatory/timeline`,

    // Audit Trail / Activity Dashboard endpoints
    AUDIT_TRAIL_LIST: `${API_BASE_URL}/admin/audit/entries`,
    AUDIT_TRAIL_BY_ID: (auditId: string) => `${API_BASE_URL}/admin/audit/entries/${auditId}`,
    AUDIT_TRAIL_STATS: `${API_BASE_URL}/admin/audit/stats`,
    AUDIT_TRAIL_EXPORT: `${API_BASE_URL}/admin/audit/export`,
    AUDIT_TRAIL_ADMINS: `${API_BASE_URL}/admin/audit/admins`,

    // API Keys & Webhooks endpoints
    API_KEY_BY_ID: (keyId: string) => `${API_BASE_URL}/admin/api-keys/${keyId}`,
    API_KEY_CREATE: `${API_BASE_URL}/admin/api-keys`,
    API_KEY_REVOKE: (keyId: string) => `${API_BASE_URL}/admin/api-keys/${keyId}/revoke`,
    API_KEY_STATS: `${API_BASE_URL}/admin/api-keys/stats`,
    WEBHOOK_BY_ID: (webhookId: string) => `${API_BASE_URL}/admin/webhooks/${webhookId}`,
    WEBHOOK_CREATE: `${API_BASE_URL}/admin/webhooks`,
    WEBHOOK_UPDATE: (webhookId: string) => `${API_BASE_URL}/admin/webhooks/${webhookId}`,
    WEBHOOK_DELETE: (webhookId: string) => `${API_BASE_URL}/admin/webhooks/${webhookId}`,
    WEBHOOK_TEST: (webhookId: string) => `${API_BASE_URL}/admin/webhooks/${webhookId}/test`,
    WEBHOOK_DELIVERIES: (webhookId: string) => `${API_BASE_URL}/admin/webhooks/${webhookId}/deliveries`,

    // Server Log Viewer endpoints
    SERVER_LOGS_LIST: `${API_BASE_URL}/admin/logs`,
    SERVER_LOGS_STREAM: `${API_BASE_URL}/admin/logs/stream`,
    SERVER_LOGS_STATS: `${API_BASE_URL}/admin/logs/stats`,
    SERVER_LOGS_ERROR_SUMMARY: `${API_BASE_URL}/admin/logs/errors/summary`,
    SERVER_LOGS_EXPORT: `${API_BASE_URL}/admin/logs/export`,

    // Platform Configuration endpoints
    PLATFORM_CONFIG_GET: `${API_BASE_URL}/admin/platform/config`,
    PLATFORM_CONFIG_UPDATE: `${API_BASE_URL}/admin/platform/config`,
    PLATFORM_CONFIG_RESET: (section: string) => `${API_BASE_URL}/admin/platform/config/${section}/reset`,
    PLATFORM_CONFIG_HISTORY: `${API_BASE_URL}/admin/platform/config/history`,

    // Margin Call Management endpoints
    MARGIN_CALL_AT_RISK_ACCOUNTS: `${API_BASE_URL}/admin/margin-calls/at-risk`,
    MARGIN_CALL_STATS: `${API_BASE_URL}/admin/margin-calls/stats`,
    MARGIN_CALL_NOTIFY: (accountId: string) => `${API_BASE_URL}/admin/margin-calls/${accountId}/notify`,
    MARGIN_CALL_CLOSE_POSITIONS: (accountId: string) => `${API_BASE_URL}/admin/margin-calls/${accountId}/close-positions`,
    MARGIN_CALL_SETTINGS: `${API_BASE_URL}/admin/margin-calls/settings`,
    MARGIN_CALL_HISTORY: `${API_BASE_URL}/admin/margin-calls/history`,
    MARGIN_CALL_DISTRIBUTION: `${API_BASE_URL}/admin/margin-calls/distribution`,

    // Trading Competitions endpoints
    COMPETITIONS_LIST: `${API_BASE_URL}/admin/competitions`,
    COMPETITION_BY_ID: (competitionId: string) => `${API_BASE_URL}/admin/competitions/${competitionId}`,
    COMPETITION_CREATE: `${API_BASE_URL}/admin/competitions`,
    COMPETITION_UPDATE: (competitionId: string) => `${API_BASE_URL}/admin/competitions/${competitionId}`,
    COMPETITION_DELETE: (competitionId: string) => `${API_BASE_URL}/admin/competitions/${competitionId}`,
    COMPETITION_PARTICIPANTS: (competitionId: string) => `${API_BASE_URL}/admin/competitions/${competitionId}/participants`,
    COMPETITION_LEADERBOARD: (competitionId: string) => `${API_BASE_URL}/admin/competitions/${competitionId}/leaderboard`,
    COMPETITION_DISQUALIFY: (competitionId: string, participantId: string) => `${API_BASE_URL}/admin/competitions/${competitionId}/participants/${participantId}/disqualify`,
    COMPETITION_STATS: (competitionId: string) => `${API_BASE_URL}/admin/competitions/${competitionId}/stats`,

    // Client Segmentation endpoints
    SEGMENTS_LIST: `${API_BASE_URL}/admin/segments`,
    SEGMENT_BY_ID: (segmentId: string) => `${API_BASE_URL}/admin/segments/${segmentId}`,
    SEGMENT_CREATE: `${API_BASE_URL}/admin/segments`,
    SEGMENT_UPDATE: (segmentId: string) => `${API_BASE_URL}/admin/segments/${segmentId}`,
    SEGMENT_DELETE: (segmentId: string) => `${API_BASE_URL}/admin/segments/${segmentId}`,
    SEGMENT_CLIENTS: (segmentId: string) => `${API_BASE_URL}/admin/segments/${segmentId}/clients`,
    SEGMENT_STATS: `${API_BASE_URL}/admin/segments/stats`,
    SEGMENT_OVERLAPS: `${API_BASE_URL}/admin/segments/overlaps`,
    SEGMENT_BULK_EMAIL: (segmentId: string) => `${API_BASE_URL}/admin/segments/${segmentId}/bulk-email`,
    SEGMENT_BULK_ACTION: (segmentId: string) => `${API_BASE_URL}/admin/segments/${segmentId}/bulk-action`,

    // White Label / Branding Configuration endpoints
    WHITE_LABEL_CONFIG_GET: `${API_BASE_URL}/admin/white-label/config`,
    WHITE_LABEL_CONFIG_UPDATE: `${API_BASE_URL}/admin/white-label/config`,
    WHITE_LABEL_PRESETS: `${API_BASE_URL}/admin/white-label/presets`,
    WHITE_LABEL_UPLOAD_LOGO: `${API_BASE_URL}/admin/white-label/upload/logo`,
    WHITE_LABEL_UPLOAD_FAVICON: `${API_BASE_URL}/admin/white-label/upload/favicon`,
    WHITE_LABEL_DNS_VERIFY: `${API_BASE_URL}/admin/white-label/dns/verify`,
    WHITE_LABEL_SSL_STATUS: `${API_BASE_URL}/admin/white-label/ssl/status`,

    // Client Activity / Session Log endpoints
    CLIENT_ACTIVITY_LOG: `${API_BASE_URL}/admin/activity/log`,
    CLIENT_ACTIVITY_STATS: `${API_BASE_URL}/admin/activity/stats`,
    CLIENT_ACTIVITY_BY_HOUR: `${API_BASE_URL}/admin/activity/by-hour`,
    CLIENT_ACTIVITY_GEO_DISTRIBUTION: `${API_BASE_URL}/admin/activity/geo-distribution`,
    CLIENT_ACTIVITY_SUSPICIOUS: `${API_BASE_URL}/admin/activity/suspicious`,
    CLIENT_SESSIONS_ACTIVE: `${API_BASE_URL}/admin/sessions/active`,
    CLIENT_SESSION_TERMINATE: (sessionId: string) => `${API_BASE_URL}/admin/sessions/${sessionId}/terminate`,

    // Promotion / Bonus Management endpoints
    PROMOTIONS_LIST: `${API_BASE_URL}/admin/promotions`,
    PROMOTION_BY_ID: (promotionId: string) => `${API_BASE_URL}/admin/promotions/${promotionId}`,
    PROMOTION_CREATE: `${API_BASE_URL}/admin/promotions`,
    PROMOTION_UPDATE: (promotionId: string) => `${API_BASE_URL}/admin/promotions/${promotionId}`,
    PROMOTION_DELETE: (promotionId: string) => `${API_BASE_URL}/admin/promotions/${promotionId}`,
    PROMOTION_CLAIMED_CLIENTS: (promotionId: string) => `${API_BASE_URL}/admin/promotions/${promotionId}/claimed`,
    PROMOTION_STATS: `${API_BASE_URL}/admin/promotions/stats`,
    PROMOTION_MONTHLY_PAYOUTS: `${API_BASE_URL}/admin/promotions/monthly-payouts`,

    // Trade Reconciliation / Discrepancy Report endpoints
    RECONCILIATION_DISCREPANCIES: `${API_BASE_URL}/admin/reconciliation/discrepancies`,
    RECONCILIATION_STATS: `${API_BASE_URL}/admin/reconciliation/stats`,
    RECONCILIATION_LP_COMPARISON: `${API_BASE_URL}/admin/reconciliation/lp-comparison`,
    RECONCILIATION_AUTO_MATCH: `${API_BASE_URL}/admin/reconciliation/auto-match`,
    RECONCILIATION_EXPORT: `${API_BASE_URL}/admin/reconciliation/export`,
    RECONCILIATION_SETTINGS: `${API_BASE_URL}/admin/reconciliation/settings`,

    // Client Notes / CRM endpoints
    CLIENT_NOTES_LIST: (clientId: string) => `${API_BASE_URL}/admin/clients/${clientId}/notes`,
    CLIENT_NOTE_CREATE: (clientId: string) => `${API_BASE_URL}/admin/clients/${clientId}/notes`,
    CLIENT_NOTE_UPDATE: (clientId: string, noteId: string) => `${API_BASE_URL}/admin/clients/${clientId}/notes/${noteId}`,
    CLIENT_NOTE_DELETE: (clientId: string, noteId: string) => `${API_BASE_URL}/admin/clients/${clientId}/notes/${noteId}`,
    CLIENT_NOTES_STATS: `${API_BASE_URL}/admin/clients/notes/stats`,
    CLIENT_NOTES_FOLLOW_UPS: `${API_BASE_URL}/admin/clients/notes/follow-ups`,
    CLIENT_NOTES_COMPLETE_FOLLOW_UP: (noteId: string) => `${API_BASE_URL}/admin/clients/notes/${noteId}/complete-followup`,
    CLIENT_SEARCH: `${API_BASE_URL}/admin/clients/search`,

    // Spread Monitor / Spread History endpoints
    SPREAD_MONITOR_LIVE: `${API_BASE_URL}/admin/spreads/live`,
    SPREAD_HISTORY: (symbol: string) => `${API_BASE_URL}/admin/spreads/history/${symbol}`,
    SPREAD_STATS: `${API_BASE_URL}/admin/spreads/stats`,
    SPREAD_HEATMAP: (symbol: string) => `${API_BASE_URL}/admin/spreads/heatmap/${symbol}`,
    SPREAD_LP_COMPARISON: (symbol: string) => `${API_BASE_URL}/admin/spreads/lp-comparison/${symbol}`,
    SPREAD_ALERTS_LIST: `${API_BASE_URL}/admin/spreads/alerts`,
    SPREAD_ALERT_CREATE: `${API_BASE_URL}/admin/spreads/alerts`,
    SPREAD_ALERT_UPDATE: (alertId: string) => `${API_BASE_URL}/admin/spreads/alerts/${alertId}`,
    SPREAD_ALERT_DELETE: (alertId: string) => `${API_BASE_URL}/admin/spreads/alerts/${alertId}`,

    // Deposit / Withdrawal Processing endpoints
    DEPOSIT_WITHDRAWAL_PENDING: `${API_BASE_URL}/admin/transactions/pending`,
    DEPOSIT_WITHDRAWAL_ALL: `${API_BASE_URL}/admin/transactions`,
    DEPOSIT_WITHDRAWAL_BY_ID: (txId: string) => `${API_BASE_URL}/admin/transactions/${txId}`,
    DEPOSIT_WITHDRAWAL_APPROVE: (txId: string) => `${API_BASE_URL}/admin/transactions/${txId}/approve`,
    DEPOSIT_WITHDRAWAL_REJECT: (txId: string) => `${API_BASE_URL}/admin/transactions/${txId}/reject`,
    DEPOSIT_WITHDRAWAL_STATS: `${API_BASE_URL}/admin/transactions/stats`,
    DEPOSIT_WITHDRAWAL_DAILY_VOLUME: `${API_BASE_URL}/admin/transactions/daily-volume`,
    DEPOSIT_WITHDRAWAL_METHOD_BREAKDOWN: `${API_BASE_URL}/admin/transactions/method-breakdown`,

    // Order Flow / Market Microstructure endpoints
    ORDER_FLOW_TRADES: `${API_BASE_URL}/admin/order-flow/trades`,
    ORDER_FLOW_METRICS: `${API_BASE_URL}/admin/order-flow/metrics`,
    ORDER_FLOW_IMBALANCE: `${API_BASE_URL}/admin/order-flow/imbalance`,
    ORDER_FLOW_LARGE_ORDERS: `${API_BASE_URL}/admin/order-flow/large-orders`,
    ORDER_FLOW_HEATMAP: `${API_BASE_URL}/admin/order-flow/heatmap`,
    ORDER_FLOW_TOP_SYMBOLS: `${API_BASE_URL}/admin/order-flow/top-symbols`,

    // MAM / PAMM Manager endpoints
    MAM_MANAGERS_LIST: `${API_BASE_URL}/admin/mam/managers`,
    MAM_MANAGER_BY_ID: (managerId: string) => `${API_BASE_URL}/admin/mam/managers/${managerId}`,
    MAM_MANAGER_CREATE: `${API_BASE_URL}/admin/mam/managers`,
    MAM_MANAGER_UPDATE: (managerId: string) => `${API_BASE_URL}/admin/mam/managers/${managerId}`,
    MAM_MANAGER_DELETE: (managerId: string) => `${API_BASE_URL}/admin/mam/managers/${managerId}`,
    MAM_MANAGER_INVESTORS: (managerId: string) => `${API_BASE_URL}/admin/mam/managers/${managerId}/investors`,
    MAM_MANAGER_PERFORMANCE: (managerId: string) => `${API_BASE_URL}/admin/mam/managers/${managerId}/performance`,
    MAM_MANAGER_STATS: `${API_BASE_URL}/admin/mam/stats`,
    MAM_INVESTOR_ADD: (managerId: string) => `${API_BASE_URL}/admin/mam/managers/${managerId}/investors`,
    MAM_INVESTOR_REMOVE: (managerId: string, investorId: string) => `${API_BASE_URL}/admin/mam/managers/${managerId}/investors/${investorId}`,

    // Performance Monitoring endpoints (Task #178)
    PERFORMANCE_METRICS: `${API_BASE_URL}/admin/performance/metrics`,
    PERFORMANCE_HISTORY: `${API_BASE_URL}/admin/performance/history`,
    PERFORMANCE_ENDPOINTS: `${API_BASE_URL}/admin/performance/endpoints`,
    PERFORMANCE_SERVICES: `${API_BASE_URL}/admin/performance/services`,
    PERFORMANCE_SLOW_QUERIES: `${API_BASE_URL}/admin/performance/slow-queries`,
    PERFORMANCE_ERROR_RATE: `${API_BASE_URL}/admin/performance/error-rate`,
    PERFORMANCE_CONNECTIONS: `${API_BASE_URL}/admin/performance/connections`,

    // Customer Lifecycle / Retention endpoints
    LIFECYCLE_OVERVIEW: `${API_BASE_URL}/admin/lifecycle/overview`,
    LIFECYCLE_FUNNEL: `${API_BASE_URL}/admin/lifecycle/funnel`,
    LIFECYCLE_RETENTION: `${API_BASE_URL}/admin/lifecycle/retention`,
    LIFECYCLE_AT_RISK: `${API_BASE_URL}/admin/lifecycle/at-risk`,
    LIFECYCLE_SEGMENTS: `${API_BASE_URL}/admin/lifecycle/segments`,
    LIFECYCLE_TRENDS: `${API_BASE_URL}/admin/lifecycle/trends`,
    LIFECYCLE_CLIENT_TIMELINE: (clientId: string) => `${API_BASE_URL}/admin/lifecycle/client/${clientId}`,

    // Scheduled Reports Management endpoints (Task #184)
    SCHEDULED_REPORTS_LIST: `${API_BASE_URL}/admin/reports/scheduled`,
    SCHEDULED_REPORT_CREATE: `${API_BASE_URL}/admin/reports/scheduled`,
    SCHEDULED_REPORT_BY_ID: (reportId: string) => `${API_BASE_URL}/admin/reports/scheduled/${reportId}`,
    SCHEDULED_REPORT_UPDATE: (reportId: string) => `${API_BASE_URL}/admin/reports/scheduled/${reportId}`,
    SCHEDULED_REPORT_DELETE: (reportId: string) => `${API_BASE_URL}/admin/reports/scheduled/${reportId}`,
    SCHEDULED_REPORT_RUN_NOW: (reportId: string) => `${API_BASE_URL}/admin/reports/scheduled/${reportId}/run`,
    SCHEDULED_REPORT_EXECUTIONS: (reportId: string) => `${API_BASE_URL}/admin/reports/scheduled/${reportId}/executions`,
    SCHEDULED_REPORT_TOGGLE_ACTIVE: (reportId: string) => `${API_BASE_URL}/admin/reports/scheduled/${reportId}/toggle`,
    SCHEDULED_REPORTS_TEMPLATES: `${API_BASE_URL}/admin/reports/templates`,
    SCHEDULED_REPORTS_STATS: `${API_BASE_URL}/admin/reports/scheduled/stats`,

    // Demo Account / Paper Trading endpoints (Task #186)
    DEMO_ACCOUNTS_LIST: `${API_BASE_URL}/admin/demo-accounts`,
    DEMO_ACCOUNTS_CREATE: `${API_BASE_URL}/admin/demo-accounts`,
    DEMO_ACCOUNT_BY_ID: (accountId: string) => `${API_BASE_URL}/admin/demo-accounts/${accountId}`,
    DEMO_ACCOUNT_UPDATE: (accountId: string) => `${API_BASE_URL}/admin/demo-accounts/${accountId}`,
    DEMO_ACCOUNT_DELETE: (accountId: string) => `${API_BASE_URL}/admin/demo-accounts/${accountId}`,
    DEMO_ACCOUNT_POSITIONS: (accountId: string) => `${API_BASE_URL}/admin/demo-accounts/${accountId}/positions`,
    DEMO_ACCOUNTS_STATS: `${API_BASE_URL}/admin/demo-accounts/stats`,
    DEMO_ACCOUNTS_CONVERSIONS: `${API_BASE_URL}/admin/demo-accounts/conversions`,

    // Client Documents / KYC Review endpoints (Task #189)
    CLIENT_DOCUMENTS_LIST: `${API_BASE_URL}/admin/documents`,
    CLIENT_DOCUMENTS_BY_CLIENT: (clientId: string) => `${API_BASE_URL}/admin/documents/client/${clientId}`,
    CLIENT_DOCUMENT_BY_ID: (documentId: string) => `${API_BASE_URL}/admin/documents/${documentId}`,
    CLIENT_DOCUMENT_REVIEW: (documentId: string) => `${API_BASE_URL}/admin/documents/${documentId}/review`,
    CLIENT_DOCUMENTS_PENDING: `${API_BASE_URL}/admin/documents/pending`,
    CLIENT_DOCUMENTS_EXPIRING: `${API_BASE_URL}/admin/documents/expiring`,
    CLIENT_DOCUMENTS_STATS: `${API_BASE_URL}/admin/documents/stats`,
    CLIENT_DOCUMENT_REQUEST_RESUBMISSION: (documentId: string) => `${API_BASE_URL}/admin/documents/${documentId}/request-resubmission`,

    // Commission Tiers / Rebate Management endpoints (Task #193)
    COMMISSION_TIERS_LIST: `${API_BASE_URL}/admin/commission/tiers`,
    COMMISSION_TIER_BY_ID: (tierId: string) => `${API_BASE_URL}/admin/commission/tiers/${tierId}`,
    COMMISSION_TIER_UPDATE: (tierId: string) => `${API_BASE_URL}/admin/commission/tiers/${tierId}`,
    CLIENT_TIER_ASSIGNMENTS: `${API_BASE_URL}/admin/commission/clients`,
    CLIENT_TIER_ASSIGN: (clientId: string) => `${API_BASE_URL}/admin/commission/clients/${clientId}/assign`,
    COMMISSION_TRANSACTIONS: `${API_BASE_URL}/admin/commission/transactions`,
    COMMISSION_STATS: `${API_BASE_URL}/admin/commission/stats`,
    COMMISSION_CALCULATE: `${API_BASE_URL}/admin/commission/calculate`,
    REBATE_DISTRIBUTION: `${API_BASE_URL}/admin/commission/rebates/distribution`,

    // Fraud Detection / IP Monitoring endpoints (Task #191)
    FRAUD_LOGINS: `${API_BASE_URL}/admin/fraud/logins`,
    FRAUD_CLIENT_LOGINS: (clientId: string) => `${API_BASE_URL}/admin/fraud/logins/client/${clientId}`,
    FRAUD_ALERTS: `${API_BASE_URL}/admin/fraud/alerts`,
    FRAUD_ALERT_RESOLVE: (alertId: string) => `${API_BASE_URL}/admin/fraud/alerts/${alertId}/resolve`,
    FRAUD_RULES: `${API_BASE_URL}/admin/fraud/rules`,
    FRAUD_RULE_UPDATE: (ruleId: string) => `${API_BASE_URL}/admin/fraud/rules/${ruleId}`,
    FRAUD_GEO: `${API_BASE_URL}/admin/fraud/geo`,
    FRAUD_STATS: `${API_BASE_URL}/admin/fraud/stats`,

    // Backup / Recovery Management endpoints (Task #196)
    BACKUP_LIST: `${API_BASE_URL}/admin/backups`,
    BACKUP_CREATE: `${API_BASE_URL}/admin/backups/create`,
    BACKUP_GET: (backupId: string) => `${API_BASE_URL}/admin/backups/${backupId}`,
    BACKUP_DELETE: (backupId: string) => `${API_BASE_URL}/admin/backups/${backupId}`,
    BACKUP_SCHEDULES: `${API_BASE_URL}/admin/backups/schedules`,
    BACKUP_SCHEDULE_UPDATE: (scheduleId: string) => `${API_BASE_URL}/admin/backups/schedules/${scheduleId}`,
    BACKUP_COMPONENTS: `${API_BASE_URL}/admin/backups/components`,
    BACKUP_RESTORE: (backupId: string) => `${API_BASE_URL}/admin/backups/${backupId}/restore`,
    BACKUP_STATS: `${API_BASE_URL}/admin/backups/stats`,

    // Referral Program Management endpoints (Task #200)
    REFERRAL_LINKS: `${API_BASE_URL}/admin/referrals/links`,
    REFERRAL_LINK_DETAIL: (linkId: string) => `${API_BASE_URL}/admin/referrals/links/${linkId}`,
    REFERRAL_LIST: `${API_BASE_URL}/admin/referrals/list`,
    REFERRAL_APPROVE: (referralId: string) => `${API_BASE_URL}/admin/referrals/${referralId}/approve`,
    REFERRAL_TIERS: `${API_BASE_URL}/admin/referrals/tiers`,
    REFERRAL_TIER_UPDATE: (tierId: string) => `${API_BASE_URL}/admin/referrals/tiers/${tierId}`,
    REFERRAL_STATS: `${API_BASE_URL}/admin/referrals/stats`,
    REFERRAL_TOP_REFERRERS: `${API_BASE_URL}/admin/referrals/top-referrers`,

    // Account Statement / Trade Export endpoints (Task #203)
    STATEMENT_GENERATE: (clientId: number) => `${API_BASE_URL}/admin/statements/generate/${clientId}`,
    STATEMENT_TEMPLATES: `${API_BASE_URL}/admin/statements/templates`,
    STATEMENT_HISTORY: `${API_BASE_URL}/admin/statements/history`,
    STATEMENT_EXPORT: `${API_BASE_URL}/admin/statements/export`,
    STATEMENT_SCHEDULED: `${API_BASE_URL}/admin/statements/scheduled`,
    STATEMENT_SCHEDULE_CREATE: `${API_BASE_URL}/admin/statements/scheduled`,
    STATEMENT_SCHEDULE_UPDATE: (scheduleId: number) => `${API_BASE_URL}/admin/statements/scheduled/${scheduleId}`,
    STATEMENT_STATS: `${API_BASE_URL}/admin/statements/stats`,

    // Trading Signals / Signal Provider Management endpoints (Task #198)
    SIGNALS_PROVIDERS: `${API_BASE_URL}/admin/signals/providers`,
    SIGNALS_PROVIDER_DETAIL: (providerId: string) => `${API_BASE_URL}/admin/signals/providers/${providerId}`,
    SIGNALS_PROVIDER_UPDATE: (providerId: string) => `${API_BASE_URL}/admin/signals/providers/${providerId}`,
    SIGNALS_ACTIVE: `${API_BASE_URL}/admin/signals/active`,
    SIGNALS_HISTORY: `${API_BASE_URL}/admin/signals/history`,
    SIGNALS_SUBSCRIPTIONS: `${API_BASE_URL}/admin/signals/subscriptions`,
    SIGNALS_STATS: `${API_BASE_URL}/admin/signals/stats`,
    SIGNALS_LEADERBOARD: `${API_BASE_URL}/admin/signals/leaderboard`,

    // Multi-Currency Wallet Management endpoints (Task #201)
    WALLET_LIST: `${API_BASE_URL}/admin/wallets`,
    WALLET_CLIENT: (clientId: string) => `${API_BASE_URL}/admin/wallets/client/${clientId}`,
    WALLET_ADJUST: (walletId: string) => `${API_BASE_URL}/admin/wallets/${walletId}/adjust`,
    WALLET_CURRENCIES: `${API_BASE_URL}/admin/wallets/currencies`,
    WALLET_TRANSACTIONS: `${API_BASE_URL}/admin/wallets/transactions`,
    WALLET_TRANSFER: `${API_BASE_URL}/admin/wallets/transfer`,
    WALLET_STATS: `${API_BASE_URL}/admin/wallets/stats`,
    WALLET_RECONCILIATION: `${API_BASE_URL}/admin/wallets/reconciliation`,

    // Instrument Groups / Symbol Category Management endpoints (Task #206)
    INSTRUMENT_GROUPS: `${API_BASE_URL}/admin/instruments/groups`,
    INSTRUMENT_GROUP_DETAIL: (groupId: string) => `${API_BASE_URL}/admin/instruments/groups/${groupId}`,
    INSTRUMENT_GROUP_UPDATE: (groupId: string) => `${API_BASE_URL}/admin/instruments/groups/${groupId}`,
    INSTRUMENT_GROUP_DELETE: (groupId: string) => `${API_BASE_URL}/admin/instruments/groups/${groupId}`,
    INSTRUMENT_GROUP_SYMBOLS: (groupId: string) => `${API_BASE_URL}/admin/instruments/groups/${groupId}/symbols`,
    INSTRUMENT_GROUP_ADD_SYMBOL: (groupId: string) => `${API_BASE_URL}/admin/instruments/groups/${groupId}/symbols`,
    INSTRUMENT_GROUP_REMOVE_SYMBOL: (groupId: string, symbol: string) => `${API_BASE_URL}/admin/instruments/groups/${groupId}/symbols/${symbol}`,
    INSTRUMENT_STATS: `${API_BASE_URL}/admin/instruments/stats`,

    // Client Risk Scoring / Credit Assessment endpoints (Task #208)
    RISK_SCORING_CLIENTS: `${API_BASE_URL}/admin/risk-scoring/clients`,
    RISK_SCORING_CLIENT_DETAIL: (clientId: number) => `${API_BASE_URL}/admin/risk-scoring/clients/${clientId}`,
    RISK_SCORING_RECALCULATE: (clientId: number) => `${API_BASE_URL}/admin/risk-scoring/clients/${clientId}/recalculate`,
    RISK_SCORING_FACTORS: `${API_BASE_URL}/admin/risk-scoring/factors`,
    RISK_SCORING_FACTOR_UPDATE: (factorId: number) => `${API_BASE_URL}/admin/risk-scoring/factors/${factorId}`,
    RISK_SCORING_DISTRIBUTION: `${API_BASE_URL}/admin/risk-scoring/distribution`,
    RISK_SCORING_ALERTS: `${API_BASE_URL}/admin/risk-scoring/alerts`,
    RISK_SCORING_STATS: `${API_BASE_URL}/admin/risk-scoring/stats`,

    // Trading Sessions / Market Hours Configuration endpoints (Task #211)
    TRADING_SESSIONS_LIST: `${API_BASE_URL}/admin/sessions/market-hours`,
    TRADING_SESSIONS_SYMBOL: (symbol: string) => `${API_BASE_URL}/admin/sessions/market-hours/${symbol}`,
    TRADING_SESSIONS_SYMBOL_UPDATE: (symbol: string) => `${API_BASE_URL}/admin/sessions/market-hours/${symbol}`,
    TRADING_SESSIONS_TEMPLATES: `${API_BASE_URL}/admin/sessions/templates`,
    TRADING_SESSIONS_TEMPLATE_CREATE: `${API_BASE_URL}/admin/sessions/templates`,
    TRADING_SESSIONS_HOLIDAYS: `${API_BASE_URL}/admin/sessions/holidays`,
    TRADING_SESSIONS_HOLIDAY_CREATE: `${API_BASE_URL}/admin/sessions/holidays`,
    TRADING_SESSIONS_STATUS: `${API_BASE_URL}/admin/sessions/status`,

    // Compliance Reports / Regulatory Reporting endpoints (Task #214)
    COMPLIANCE_REPORTS_LIST: `${API_BASE_URL}/admin/compliance/reports`,
    COMPLIANCE_REPORT_DETAIL: (reportId: number) => `${API_BASE_URL}/admin/compliance/reports/${reportId}`,
    COMPLIANCE_REPORT_TEMPLATES: `${API_BASE_URL}/admin/compliance/reports/templates`,
    COMPLIANCE_CHECKS: `${API_BASE_URL}/admin/compliance/checks`,
    COMPLIANCE_CHECKS_RUN: `${API_BASE_URL}/admin/compliance/checks/run`,
    COMPLIANCE_BREACHES: `${API_BASE_URL}/admin/compliance/breaches`,
    COMPLIANCE_STATS: `${API_BASE_URL}/admin/compliance/stats`,
};

/**
 * Get WebSocket URL with optional JWT token
 */
export function getWebSocketUrl(baseUrl: string, token?: string): string {
    if (!token && typeof window !== 'undefined') {
        // Try to get token from localStorage
        token = localStorage.getItem('admin_token') ||
                localStorage.getItem('rtx_token') ||
                localStorage.getItem('jwt_token') ||
                undefined;
    }
    return token ? `${baseUrl}?token=${encodeURIComponent(token)}` : baseUrl;
}

// Timeframe options for charts
export const TIMEFRAMES = [
    { value: 'M1', label: 'M1' },
    { value: 'M5', label: 'M5' },
    { value: 'M15', label: 'M15' },
    { value: 'M30', label: 'M30' },
    { value: 'H1', label: 'H1' },
    { value: 'H4', label: 'H4' },
    { value: 'D1', label: 'D1' },
    { value: 'W1', label: 'W1' },
    { value: 'MN', label: 'MN' },
] as const;

export type Timeframe = typeof TIMEFRAMES[number]['value'];

// Export API_ENDPOINTS as alias for API_CONFIG for backward compatibility
export const API_ENDPOINTS = API_CONFIG;

// Export individual URLs for backward compatibility
export const ADMIN_API_URL = API_CONFIG.ADMIN_API_URL;
export const MARKET_DATA_URL = API_CONFIG.MARKET_DATA_URL;
export const MARKET_WS_URL = API_CONFIG.MARKET_WS_URL;
export const ADMIN_WS_URL = API_CONFIG.ADMIN_WS_URL;
