/**
 * Comprehensive API Service Layer
 * Centralized API communication with type safety and error handling
 * Includes authentication with JWT token management and 401 handling
 */

import type { Position, Order, Account, Trade, Tick } from '../store/useAppStore';
import { useAppStore } from '../store/useAppStore';

// ============================================
// API Response Types
// ============================================

export interface ApiResponse<T = any> {
  success: boolean;
  data?: T;
  error?: string;
  message?: string;
}

export interface BrokerConfig {
  brokerName: string;
  priceFeedLP: string;
  executionMode: 'ABOOK' | 'BBOOK';
  defaultLeverage: number;
  defaultBalance: number;
  marginMode: 'HEDGING' | 'NETTING';
  maxTicksPerSymbol: number;
  disabledSymbols?: Record<string, boolean>;
}

export interface Symbol {
  symbol: string;
  displayName: string;
  category: string;
  enabled: boolean;
  contractSize: number;
  pipValue: number;
  minVolume: number;
  maxVolume: number;
  stepVolume: number;
  spread?: number;
}

export interface LPStatus {
  id: string;
  name: string;
  type: string;
  enabled: boolean;
  connected: boolean;
  symbolCount: number;
  latency?: number;
  lastQuote?: number;
}

export interface MarginPreview {
  symbol: string;
  volume: number;
  leverage: number;
  requiredMargin: number;
  currentMargin: number;
  freeMargin: number;
  marginAfter: number;
  freeMarginAfter: number;
  marginLevel: number;
  marginLevelAfter: number;
  canTrade: boolean;
}

export interface LotCalculation {
  symbol: string;
  riskPercent: number;
  slPips: number;
  balance: number;
  recommendedLot: number;
  riskAmount: number;
  pipValue: number;
}

export interface LedgerEntry {
  id: number;
  accountId: number;
  type: string;
  amount: number;
  balance: number;
  description: string;
  timestamp: string;
  referenceId?: string;
}

export interface AdminAccount {
  id: number;
  accountNumber: string;
  username: string;
  displayName: string;
  balance: number;
  equity: number;
  margin: number;
  freeMargin: number;
  leverage: number;
  executionMode: string;
  status: string;
  createdAt: string;
}

// ============================================
// API Configuration
// ============================================

const API_BASE_URL = import.meta.env.VITE_API_URL || 'http://localhost:8080';
const DEFAULT_TIMEOUT = 10000; // 10 seconds

// ============================================
// HTTP Client with Error Handling
// ============================================

class ApiError extends Error {
  statusCode?: number;
  response?: unknown;

  constructor(
    message: string,
    statusCode?: number,
    response?: unknown
  ) {
    super(message);
    this.name = 'ApiError';
    this.statusCode = statusCode;
    this.response = response;
  }
}

async function fetchWithTimeout(
  url: string,
  options: RequestInit = {},
  timeout = DEFAULT_TIMEOUT
): Promise<Response> {
  const controller = new AbortController();
  const timeoutId = setTimeout(() => controller.abort(), timeout);

  try {
    // Get auth token from Zustand store
    const authToken = useAppStore.getState().authToken;
    const headers: Record<string, string> = {
      'Content-Type': 'application/json',
      ...options.headers as Record<string, string>,
    };

    // Add Authorization header if token exists
    if (authToken) {
      headers['Authorization'] = `Bearer ${authToken}`;
    }

    const response = await fetch(url, {
      ...options,
      signal: controller.signal,
      headers,
    });

    clearTimeout(timeoutId);

    // Handle 401 Unauthorized - clear auth state and redirect to login
    if (response.status === 401) {
      useAppStore.getState().clearAuth();
      // Redirect to login page by triggering window navigation
      if (typeof window !== 'undefined') {
        window.location.href = '/';
      }
      throw new ApiError('Unauthorized - please log in again', 401, response);
    }

    return response;
  } catch (error: any) {
    clearTimeout(timeoutId);
    if (error.name === 'AbortError') {
      throw new ApiError('Request timeout', 408);
    }
    throw error;
  }
}

async function handleResponse<T>(response: Response): Promise<T> {
  const contentType = response.headers.get('content-type');
  const isJson = contentType?.includes('application/json');

  if (!response.ok) {
    let errorMessage = `HTTP ${response.status}: ${response.statusText}`;

    if (isJson) {
      const errorData = await response.json();
      errorMessage = errorData.error || errorData.message || errorMessage;
    } else {
      errorMessage = await response.text() || errorMessage;
    }

    throw new ApiError(errorMessage, response.status, response);
  }

  if (isJson) {
    return await response.json();
  }

  return {} as T;
}

// ============================================
// Authentication API
// ============================================

export const authApi = {
  async login(username: string, password: string): Promise<{ token: string; user: any }> {
    // Don't use fetchWithTimeout for login since we don't have a token yet
    const controller = new AbortController();
    const timeoutId = setTimeout(() => controller.abort(), DEFAULT_TIMEOUT);

    try {
      const response = await fetch(`${API_BASE_URL}/login`, {
        method: 'POST',
        signal: controller.signal,
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify({ username, password }),
      });

      clearTimeout(timeoutId);

      const result = await handleResponse<{ token: string; user: any }>(response);

      // Store token in Zustand store after successful login
      if (result.token) {
        useAppStore.getState().setAuthToken(result.token);
      }

      return result;
    } catch (error: any) {
      clearTimeout(timeoutId);
      if (error.name === 'AbortError') {
        throw new ApiError('Login request timeout', 408);
      }
      throw error;
    }
  },

  logout(): void {
    useAppStore.getState().clearAuth();
  },
};

// ============================================
// Account API
// ============================================

export const accountApi = {
  async getAccountSummary(accountId: string): Promise<Account> {
    const response = await fetchWithTimeout(
      `${API_BASE_URL}/api/account/summary?accountId=${accountId}`
    );

    return handleResponse(response);
  },

  async createAccount(data: {
    username: string;
    displayName: string;
    password: string;
    balance: number;
  }): Promise<AdminAccount> {
    const response = await fetchWithTimeout(`${API_BASE_URL}/api/account/create`, {
      method: 'POST',
      body: JSON.stringify(data),
    });

    return handleResponse(response);
  },
};

// ============================================
// Positions API
// ============================================

export const positionsApi = {
  async getPositions(accountId: string): Promise<Position[]> {
    const response = await fetchWithTimeout(
      `${API_BASE_URL}/api/positions?accountId=${accountId}`
    );

    const data = await handleResponse<Position[]>(response);
    return data || [];
  },

  async closePosition(positionId: number, volume?: number): Promise<void> {
    const response = await fetchWithTimeout(`${API_BASE_URL}/api/positions/close`, {
      method: 'POST',
      body: JSON.stringify({ positionId, volume }),
    });

    return handleResponse(response);
  },

  async modifyPosition(positionId: number, sl: number, tp: number): Promise<void> {
    const response = await fetchWithTimeout(`${API_BASE_URL}/api/positions/modify`, {
      method: 'POST',
      body: JSON.stringify({ positionId, sl, tp }),
    });

    return handleResponse(response);
  },

  async closeBulk(
    accountId: number,
    type: 'ALL' | 'WINNERS' | 'LOSERS',
    symbol?: string
  ): Promise<{ closed: number; message: string }> {
    const response = await fetchWithTimeout(`${API_BASE_URL}/api/positions/close-bulk`, {
      method: 'POST',
      body: JSON.stringify({ accountId, type, symbol }),
    });

    return handleResponse(response);
  },
};

// ============================================
// Orders API
// ============================================

export const ordersApi = {
  async placeMarketOrder(data: {
    accountId: number;
    symbol: string;
    side: 'BUY' | 'SELL';
    volume: number;
    sl?: number;
    tp?: number;
  }): Promise<Position> {
    const response = await fetchWithTimeout(`${API_BASE_URL}/api/orders/market`, {
      method: 'POST',
      body: JSON.stringify(data),
    });

    return handleResponse(response);
  },

  async placeLimitOrder(data: {
    symbol: string;
    side: 'BUY' | 'SELL';
    volume: number;
    price: number;
    sl?: number;
    tp?: number;
  }): Promise<Order> {
    const response = await fetchWithTimeout(`${API_BASE_URL}/order/limit`, {
      method: 'POST',
      body: JSON.stringify(data),
    });

    return handleResponse(response);
  },

  async placeStopOrder(data: {
    symbol: string;
    side: 'BUY' | 'SELL';
    volume: number;
    triggerPrice: number;
    sl?: number;
    tp?: number;
  }): Promise<Order> {
    const response = await fetchWithTimeout(`${API_BASE_URL}/order/stop`, {
      method: 'POST',
      body: JSON.stringify(data),
    });

    return handleResponse(response);
  },

  async placeStopLimitOrder(data: {
    symbol: string;
    side: 'BUY' | 'SELL';
    volume: number;
    triggerPrice: number;
    limitPrice: number;
    sl?: number;
    tp?: number;
  }): Promise<Order> {
    const response = await fetchWithTimeout(`${API_BASE_URL}/order/stop-limit`, {
      method: 'POST',
      body: JSON.stringify(data),
    });

    return handleResponse(response);
  },

  async getPendingOrders(): Promise<Order[]> {
    const response = await fetchWithTimeout(`${API_BASE_URL}/orders/pending`);

    const data = await handleResponse<Order[]>(response);
    return data || [];
  },

  async cancelOrder(orderId: string): Promise<void> {
    const response = await fetchWithTimeout(`${API_BASE_URL}/order/cancel`, {
      method: 'POST',
      body: JSON.stringify({ orderId }),
    });

    return handleResponse(response);
  },
};

// ============================================
// Market Data API
// ============================================

export const marketApi = {
  async getSymbols(): Promise<Symbol[]> {
    const response = await fetchWithTimeout(`${API_BASE_URL}/api/symbols`);

    const data = await handleResponse<Symbol[]>(response);
    return data || [];
  },

  async getTicks(symbol: string, limit = 500): Promise<Tick[]> {
    const response = await fetchWithTimeout(
      `${API_BASE_URL}/ticks?symbol=${symbol}&limit=${limit}`
    );

    const data = await handleResponse<Tick[]>(response);
    return data || [];
  },

  async getOHLC(
    symbol: string,
    timeframe: string,
    limit = 500
  ): Promise<any[]> {
    const response = await fetchWithTimeout(
      `${API_BASE_URL}/ohlc?symbol=${symbol}&timeframe=${timeframe}&limit=${limit}`
    );

    const data = await handleResponse<any[]>(response);
    return data || [];
  },
};

// ============================================
// Risk Management API
// ============================================

export const riskApi = {
  async calculateLot(
    symbol: string,
    riskPercent: number,
    slPips: number
  ): Promise<LotCalculation> {
    const response = await fetchWithTimeout(
      `${API_BASE_URL}/risk/calculate-lot?symbol=${symbol}&riskPercent=${riskPercent}&slPips=${slPips}`
    );

    return handleResponse(response);
  },

  async previewMargin(
    symbol: string,
    volume: number,
    side: string
  ): Promise<MarginPreview> {
    const response = await fetchWithTimeout(
      `${API_BASE_URL}/risk/margin-preview?symbol=${symbol}&volume=${volume}&side=${side}`
    );

    return handleResponse(response);
  },
};

// ============================================
// Trading History API
// ============================================

export const historyApi = {
  async getTrades(accountId: string): Promise<Trade[]> {
    const response = await fetchWithTimeout(
      `${API_BASE_URL}/api/trades?accountId=${accountId}`
    );

    const data = await handleResponse<Trade[]>(response);
    return data || [];
  },

  async getLedger(accountId: string): Promise<LedgerEntry[]> {
    const response = await fetchWithTimeout(
      `${API_BASE_URL}/api/ledger?accountId=${accountId}`
    );

    const data = await handleResponse<LedgerEntry[]>(response);
    return data || [];
  },
};

// ============================================
// Configuration API
// ============================================

export const configApi = {
  async getConfig(): Promise<BrokerConfig> {
    const response = await fetchWithTimeout(`${API_BASE_URL}/api/config`);

    return handleResponse(response);
  },

  async updateConfig(config: Partial<BrokerConfig>): Promise<{ success: boolean; config: BrokerConfig }> {
    const response = await fetchWithTimeout(`${API_BASE_URL}/api/config`, {
      method: 'POST',
      body: JSON.stringify(config),
    });

    return handleResponse(response);
  },
};

// ============================================
// Admin API
// ============================================

export const adminApi = {
  async getAccounts(): Promise<AdminAccount[]> {
    const response = await fetchWithTimeout(`${API_BASE_URL}/admin/accounts`);

    const data = await handleResponse<AdminAccount[]>(response);
    return data || [];
  },

  async deposit(accountId: number, amount: number, method: string, description: string): Promise<void> {
    const response = await fetchWithTimeout(`${API_BASE_URL}/admin/deposit`, {
      method: 'POST',
      body: JSON.stringify({ accountId, amount, method, description }),
    });

    return handleResponse(response);
  },

  async withdraw(accountId: number, amount: number, method: string, description: string): Promise<void> {
    const response = await fetchWithTimeout(`${API_BASE_URL}/admin/withdraw`, {
      method: 'POST',
      body: JSON.stringify({ accountId, amount, method, description }),
    });

    return handleResponse(response);
  },

  async adjust(accountId: number, amount: number, description: string): Promise<void> {
    const response = await fetchWithTimeout(`${API_BASE_URL}/admin/adjust`, {
      method: 'POST',
      body: JSON.stringify({ accountId, amount, description }),
    });

    return handleResponse(response);
  },

  async bonus(accountId: number, amount: number, description: string): Promise<void> {
    const response = await fetchWithTimeout(`${API_BASE_URL}/admin/bonus`, {
      method: 'POST',
      body: JSON.stringify({ accountId, amount, description }),
    });

    return handleResponse(response);
  },

  async getLedgerAll(): Promise<LedgerEntry[]> {
    const response = await fetchWithTimeout(`${API_BASE_URL}/admin/ledger`);

    const data = await handleResponse<LedgerEntry[]>(response);
    return data || [];
  },

  async updateAccount(accountId: number, data: Partial<AdminAccount>): Promise<void> {
    const response = await fetchWithTimeout(`${API_BASE_URL}/admin/account/update`, {
      method: 'POST',
      body: JSON.stringify({ accountId, ...data }),
    });

    return handleResponse(response);
  },

  async resetPassword(accountId: number, newPassword: string): Promise<void> {
    const response = await fetchWithTimeout(`${API_BASE_URL}/admin/reset-password`, {
      method: 'POST',
      body: JSON.stringify({ accountId, newPassword }),
    });

    return handleResponse(response);
  },

  async getSymbols(): Promise<Symbol[]> {
    const response = await fetchWithTimeout(`${API_BASE_URL}/admin/symbols`);

    const data = await handleResponse<Symbol[]>(response);
    return data || [];
  },

  async toggleSymbol(symbol: string, enabled: boolean): Promise<void> {
    const response = await fetchWithTimeout(`${API_BASE_URL}/admin/symbols/toggle`, {
      method: 'POST',
      body: JSON.stringify({ symbol, enabled }),
    });

    return handleResponse(response);
  },

  async getExecutionMode(): Promise<{ mode: string; description: any; priceFeed: string }> {
    const response = await fetchWithTimeout(`${API_BASE_URL}/admin/execution-mode`);

    return handleResponse(response);
  },

  async setExecutionMode(mode: 'ABOOK' | 'BBOOK'): Promise<any> {
    const response = await fetchWithTimeout(`${API_BASE_URL}/admin/execution-mode`, {
      method: 'POST',
      body: JSON.stringify({ mode }),
    });

    return handleResponse(response);
  },

  async getLPStatus(): Promise<LPStatus[]> {
    const response = await fetchWithTimeout(`${API_BASE_URL}/admin/lp-status`);

    const data = await handleResponse<LPStatus[]>(response);
    return data || [];
  },

  async restartServer(): Promise<{ success: boolean; message: string }> {
    const response = await fetchWithTimeout(`${API_BASE_URL}/admin/restart`, {
      method: 'POST',
    });

    return handleResponse(response);
  },
};

// ============================================
// Alerts API
// ============================================

export type AlertRule = {
  id: string;
  name: string;
  enabled: boolean;
  condition: {
    type: 'price_above' | 'price_below' | 'pnl_above' | 'pnl_below' | 'margin_level_below';
    symbol?: string;
    threshold: number;
  };
  severity: 'LOW' | 'MEDIUM' | 'HIGH' | 'CRITICAL';
  message: string;
};

export const alertsApi = {
  async getRules(): Promise<AlertRule[]> {
    const response = await fetchWithTimeout(`${API_BASE_URL}/api/alerts/rules`);
    const data = await handleResponse<AlertRule[]>(response);
    return data || [];
  },

  async createRule(rule: Omit<AlertRule, 'id'>): Promise<AlertRule> {
    const response = await fetchWithTimeout(`${API_BASE_URL}/api/alerts/rules`, {
      method: 'POST',
      body: JSON.stringify(rule),
    });
    return handleResponse(response);
  },

  async updateRule(id: string, updates: Partial<AlertRule>): Promise<AlertRule> {
    const response = await fetchWithTimeout(`${API_BASE_URL}/api/alerts/rules/${id}`, {
      method: 'PUT',
      body: JSON.stringify(updates),
    });
    return handleResponse(response);
  },

  async deleteRule(id: string): Promise<void> {
    const response = await fetchWithTimeout(`${API_BASE_URL}/api/alerts/rules/${id}`, {
      method: 'DELETE',
    });
    return handleResponse(response);
  },

  async testRule(id: string): Promise<void> {
    const response = await fetchWithTimeout(`${API_BASE_URL}/api/alerts/rules/${id}/test`, {
      method: 'POST',
    });
    return handleResponse(response);
  },
};

// ============================================
// Analytics & Export API
// ============================================

export const analyticsApi = {
  async exportPDF(accountId: string, startDate: string, endDate: string): Promise<Blob> {
    const queryParams = new URLSearchParams({
      start: startDate,
      end: endDate,
      accountId,
    });

    const response = await fetchWithTimeout(
      `${API_BASE_URL}/api/analytics/export/pdf?${queryParams}`
    );

    if (!response.ok) {
      throw new ApiError('PDF export failed', response.status);
    }

    return await response.blob();
  },

  async exportCSV(accountId: string, startDate: string, endDate: string, dataType: string): Promise<string> {
    const queryParams = new URLSearchParams({
      start: startDate,
      end: endDate,
      accountId,
      type: dataType,
    });

    const response = await fetchWithTimeout(
      `${API_BASE_URL}/api/analytics/export/csv?${queryParams}`
    );

    if (!response.ok) {
      throw new ApiError('CSV export failed', response.status);
    }

    return await response.text();
  },

  async getPerformanceReport(accountId: string, startDate: string, endDate: string): Promise<any> {
    const queryParams = new URLSearchParams({
      start: startDate,
      end: endDate,
      accountId,
    });

    const response = await fetchWithTimeout(
      `${API_BASE_URL}/api/performance?${queryParams}`
    );

    return handleResponse(response);
  },
};

// ============================================
// Export All APIs
// ============================================

export const api = {
  auth: authApi,
  account: accountApi,
  positions: positionsApi,
  orders: ordersApi,
  market: marketApi,
  risk: riskApi,
  history: historyApi,
  config: configApi,
  admin: adminApi,
  analytics: analyticsApi,
  alerts: alertsApi,
};

export default api;
