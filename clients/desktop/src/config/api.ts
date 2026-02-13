/**
 * Centralized API Configuration
 * All API endpoints and WebSocket URLs are defined here
 * Environment variables (VITE_*) can override defaults
 */

// Base URLs
const API_BASE_URL = import.meta.env.VITE_API_URL || 'http://localhost:7999';
const ADMIN_API_URL = import.meta.env.VITE_ADMIN_URL || 'http://localhost:7999';

// WebSocket URLs
const WS_URL = import.meta.env.VITE_WS_URL || 'ws://localhost:7999/ws';
const MARKET_WS_URL = import.meta.env.VITE_MARKET_WS_URL || 'ws://localhost:7999/market-data';
const ADMIN_WS_URL = import.meta.env.VITE_ADMIN_WS_URL || 'ws://localhost:7999/admin-ws';

// API Endpoints
export const API_ENDPOINTS = {
  // Auth
  login: `${API_BASE_URL}/login`,
  logout: `${API_BASE_URL}/logout`,

  // Market Data
  symbols: `${API_BASE_URL}/api/symbols`,
  ticks: `${API_BASE_URL}/ticks`,
  quotes: `${API_BASE_URL}/quotes`,
  candles: `${API_BASE_URL}/candles`,
  history: `${API_BASE_URL}/history`,

  // Trading
  order: `${API_BASE_URL}/order`,
  orders: `${API_BASE_URL}/orders`,
  positions: `${API_BASE_URL}/positions`,
  closePosition: `${API_BASE_URL}/close-position`,
  modifyOrder: `${API_BASE_URL}/modify-order`,

  // Account
  account: {
    summary: `${API_BASE_URL}/account/summary`,
    settings: `${API_BASE_URL}/account/settings`,
    password: `${API_BASE_URL}/account/password`,
  },
  accounts: `${API_BASE_URL}/accounts`,
  balance: `${API_BASE_URL}/balance`,

  // Admin
  admin: {
    symbols: `${ADMIN_API_URL}/admin/symbols`,
    accounts: `${ADMIN_API_URL}/admin/accounts`,
    users: `${ADMIN_API_URL}/admin/users`,
    config: `${API_BASE_URL}/api/config`,
    lp: `${ADMIN_API_URL}/admin/lp`,
    routing: `${ADMIN_API_URL}/admin/routing`,
    routingRules: `${ADMIN_API_URL}/admin/routing-rules`,
    routingMetrics: `${ADMIN_API_URL}/admin/routing/metrics`,
    routingRuleEffectiveness: `${ADMIN_API_URL}/admin/routing/rules/effectiveness`,
    routingPreview: `${ADMIN_API_URL}/admin/routing/preview`,
    lpStatus: `${ADMIN_API_URL}/admin/lp/status`,
    lpMetrics: `${ADMIN_API_URL}/admin/lp/metrics`,
    lpCompare: `${ADMIN_API_URL}/admin/lp/compare`,
  },

  // Alerts (Price Alerts)
  priceAlerts: {
    list: `${API_BASE_URL}/price-alerts`,
    create: `${API_BASE_URL}/price-alerts`,
    update: `${API_BASE_URL}/price-alerts`,
    delete: `${API_BASE_URL}/price-alerts`,
  },

  // Legacy alerts (keep for backward compatibility)
  alerts: `${API_BASE_URL}/alerts`,
  alertRules: `${API_BASE_URL}/alert-rules`,

  // Workspace
  workspace: {
    layouts: `${API_BASE_URL}/workspace/layouts`,
    drawings: `${API_BASE_URL}/workspace/drawings`,
    preferences: `${API_BASE_URL}/workspace/preferences`,
    templates: `${API_BASE_URL}/workspace/templates`,
  },

  // Data Management
  dataFolder: `${API_BASE_URL}/data-folder`,

  // Chart Analysis
  chartAnalysis: `${API_BASE_URL}/chart-analysis`,

  // Economic Calendar (placeholder for future real API integration)
  // Can be replaced with ForexFactory, Investing.com, or custom API
  economicCalendar: `${API_BASE_URL}/economic-calendar`,

  // Market News / Economic Calendar (admin backend)
  marketNews: {
    articles: `${ADMIN_API_URL}/admin/market-news/articles`,
    calendar: `${ADMIN_API_URL}/admin/market-news/calendar`,
    sources: `${ADMIN_API_URL}/admin/market-news/sources`,
    stats: `${ADMIN_API_URL}/admin/market-news/stats`,
  },

  // Trading Signals (admin backend)
  signals: {
    providers: `${ADMIN_API_URL}/admin/signals/providers`,
    active: `${ADMIN_API_URL}/admin/signals/active`,
    history: `${ADMIN_API_URL}/admin/signals/history`,
    subscriptions: `${ADMIN_API_URL}/admin/signals/subscriptions`,
    stats: `${ADMIN_API_URL}/admin/signals/stats`,
    leaderboard: `${ADMIN_API_URL}/admin/signals/leaderboard`,
  },

  // Social / Copy Trading (admin backend)
  socialTrading: {
    providers: `${ADMIN_API_URL}/admin/social-trading/providers`,
    copies: `${ADMIN_API_URL}/admin/social-trading/copies`,
    stats: `${ADMIN_API_URL}/admin/social-trading/stats`,
  },

  // Trades
  trades: {
    history: `${API_BASE_URL}/api/trades/history`,
  },

  // Notifications
  notifications: {
    list: `${API_BASE_URL}/api/notifications`,
    unreadCount: `${API_BASE_URL}/api/notifications/unread-count`,
    markRead: `${API_BASE_URL}/api/notifications/mark-read`,
  },

  // Payments
  payments: {
    deposit: `${API_BASE_URL}/payments/deposit`,
    withdrawal: `${API_BASE_URL}/payments/withdrawal`,
    history: `${API_BASE_URL}/payments/history`,
  },
};

// WebSocket Endpoints
export const WS_ENDPOINTS = {
  market: MARKET_WS_URL,
  admin: ADMIN_WS_URL,
  general: WS_URL,
};

// Helper functions
export function buildApiUrl(endpoint: string): string {
  if (endpoint.startsWith('http://') || endpoint.startsWith('https://')) {
    return endpoint;
  }
  return `${API_BASE_URL}${endpoint.startsWith('/') ? '' : '/'}${endpoint}`;
}

export function buildAdminUrl(endpoint: string): string {
  if (endpoint.startsWith('http://') || endpoint.startsWith('https://')) {
    return endpoint;
  }
  return `${ADMIN_API_URL}${endpoint.startsWith('/') ? '' : '/'}${endpoint}`;
}

export function buildWsUrl(endpoint: string): string {
  if (endpoint.startsWith('ws://') || endpoint.startsWith('wss://')) {
    return endpoint;
  }
  const wsBase = WS_URL.replace(/\/ws$/, '');
  return `${wsBase}${endpoint.startsWith('/') ? '' : '/'}${endpoint}`;
}

// Export base URLs for backward compatibility
export { API_BASE_URL, ADMIN_API_URL, WS_URL, MARKET_WS_URL, ADMIN_WS_URL };
