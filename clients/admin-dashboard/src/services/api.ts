import type {
  APIResponse,
  SystemMetrics,
  Order,
  LiquidityProvider,
  UserSession,
  ErrorLog,
  RoutingRule,
  TradingSymbol,
  User,
  RiskLimit,
  TradingAnalytics,
  UserAnalytics,
  AuditLog,
  Filter,
  Sort,
  Pagination,
} from '@/types';

class APIService {
  private baseURL: string;
  private token: string | null = null;

  constructor(baseURL = '/api') {
    this.baseURL = baseURL;
    this.token = localStorage.getItem('auth_token');
  }

  setToken(token: string): void {
    this.token = token;
    localStorage.setItem('auth_token', token);
  }

  clearToken(): void {
    this.token = null;
    localStorage.removeItem('auth_token');
  }

  private async request<T>(
    endpoint: string,
    options: RequestInit = {}
  ): Promise<T> {
    const headers: HeadersInit = {
      'Content-Type': 'application/json',
      ...options.headers,
    };

    if (this.token) {
      headers['Authorization'] = `Bearer ${this.token}`;
    }

    const response = await fetch(`${this.baseURL}${endpoint}`, {
      ...options,
      headers,
    });

    if (!response.ok) {
      const error = await response.json().catch(() => ({ error: 'Request failed' }));
      throw new Error(error.error || `HTTP ${response.status}`);
    }

    const data: APIResponse<T> = await response.json();

    if (!data.success) {
      throw new Error(data.error || 'Request failed');
    }

    return data.data as T;
  }

  // Authentication
  async login(username: string, password: string): Promise<{ token: string; user: User }> {
    return this.request('/auth/login', {
      method: 'POST',
      body: JSON.stringify({ username, password }),
    });
  }

  async logout(): Promise<void> {
    await this.request('/auth/logout', { method: 'POST' });
    this.clearToken();
  }

  // System Metrics
  async getSystemMetrics(): Promise<SystemMetrics> {
    return this.request('/metrics/system');
  }

  // Orders
  async getOrders(filters?: Filter[], sort?: Sort, pagination?: Pagination): Promise<{
    orders: Order[];
    pagination: Pagination;
  }> {
    const params = new URLSearchParams();
    if (filters) params.set('filters', JSON.stringify(filters));
    if (sort) params.set('sort', JSON.stringify(sort));
    if (pagination) {
      params.set('page', pagination.page.toString());
      params.set('pageSize', pagination.pageSize.toString());
    }
    return this.request(`/orders?${params}`);
  }

  async getOrder(orderId: string): Promise<Order> {
    return this.request(`/orders/${orderId}`);
  }

  async cancelOrder(orderId: string): Promise<void> {
    return this.request(`/orders/${orderId}/cancel`, { method: 'POST' });
  }

  // Liquidity Providers
  async getLPs(): Promise<LiquidityProvider[]> {
    return this.request('/lps');
  }

  async getLP(lpId: string): Promise<LiquidityProvider> {
    return this.request(`/lps/${lpId}`);
  }

  async updateLP(lpId: string, data: Partial<LiquidityProvider>): Promise<LiquidityProvider> {
    return this.request(`/lps/${lpId}`, {
      method: 'PUT',
      body: JSON.stringify(data),
    });
  }

  async testLPConnection(lpId: string): Promise<{ success: boolean; latency: number }> {
    return this.request(`/lps/${lpId}/test`, { method: 'POST' });
  }

  // User Sessions
  async getUserSessions(): Promise<UserSession[]> {
    return this.request('/sessions');
  }

  async terminateSession(sessionId: string): Promise<void> {
    return this.request(`/sessions/${sessionId}`, { method: 'DELETE' });
  }

  // Error Logs
  async getErrorLogs(filters?: Filter[], pagination?: Pagination): Promise<{
    logs: ErrorLog[];
    pagination: Pagination;
  }> {
    const params = new URLSearchParams();
    if (filters) params.set('filters', JSON.stringify(filters));
    if (pagination) {
      params.set('page', pagination.page.toString());
      params.set('pageSize', pagination.pageSize.toString());
    }
    return this.request(`/errors?${params}`);
  }

  // Routing Rules
  async getRoutingRules(): Promise<RoutingRule[]> {
    return this.request('/routing/rules');
  }

  async createRoutingRule(rule: Omit<RoutingRule, 'id'>): Promise<RoutingRule> {
    return this.request('/routing/rules', {
      method: 'POST',
      body: JSON.stringify(rule),
    });
  }

  async updateRoutingRule(ruleId: string, data: Partial<RoutingRule>): Promise<RoutingRule> {
    return this.request(`/routing/rules/${ruleId}`, {
      method: 'PUT',
      body: JSON.stringify(data),
    });
  }

  async deleteRoutingRule(ruleId: string): Promise<void> {
    return this.request(`/routing/rules/${ruleId}`, { method: 'DELETE' });
  }

  // Trading Symbols
  async getSymbols(): Promise<TradingSymbol[]> {
    return this.request('/symbols');
  }

  async updateSymbol(symbolId: string, data: Partial<TradingSymbol>): Promise<TradingSymbol> {
    return this.request(`/symbols/${symbolId}`, {
      method: 'PUT',
      body: JSON.stringify(data),
    });
  }

  // Users
  async getUsers(filters?: Filter[]): Promise<User[]> {
    const params = filters ? `?filters=${JSON.stringify(filters)}` : '';
    return this.request(`/users${params}`);
  }

  async getUser(userId: string): Promise<User> {
    return this.request(`/users/${userId}`);
  }

  async createUser(user: Omit<User, 'id' | 'createdAt'>): Promise<User> {
    return this.request('/users', {
      method: 'POST',
      body: JSON.stringify(user),
    });
  }

  async updateUser(userId: string, data: Partial<User>): Promise<User> {
    return this.request(`/users/${userId}`, {
      method: 'PUT',
      body: JSON.stringify(data),
    });
  }

  async deleteUser(userId: string): Promise<void> {
    return this.request(`/users/${userId}`, { method: 'DELETE' });
  }

  // Risk Limits
  async getRiskLimits(): Promise<RiskLimit[]> {
    return this.request('/risk/limits');
  }

  async createRiskLimit(limit: Omit<RiskLimit, 'id'>): Promise<RiskLimit> {
    return this.request('/risk/limits', {
      method: 'POST',
      body: JSON.stringify(limit),
    });
  }

  async updateRiskLimit(limitId: string, data: Partial<RiskLimit>): Promise<RiskLimit> {
    return this.request(`/risk/limits/${limitId}`, {
      method: 'PUT',
      body: JSON.stringify(data),
    });
  }

  async deleteRiskLimit(limitId: string): Promise<void> {
    return this.request(`/risk/limits/${limitId}`, { method: 'DELETE' });
  }

  // Analytics
  async getTradingAnalytics(period: string): Promise<TradingAnalytics> {
    return this.request(`/analytics/trading?period=${period}`);
  }

  async getUserAnalytics(userId?: string): Promise<UserAnalytics[]> {
    const endpoint = userId ? `/analytics/users/${userId}` : '/analytics/users';
    return this.request(endpoint);
  }

  async getPerformanceMetrics(metric: string, period: string): Promise<Array<{ timestamp: number; value: number }>> {
    return this.request(`/analytics/performance?metric=${metric}&period=${period}`);
  }

  // Audit Trail
  async getAuditLogs(filters?: Filter[], pagination?: Pagination): Promise<{
    logs: AuditLog[];
    pagination: Pagination;
  }> {
    const params = new URLSearchParams();
    if (filters) params.set('filters', JSON.stringify(filters));
    if (pagination) {
      params.set('page', pagination.page.toString());
      params.set('pageSize', pagination.pageSize.toString());
    }
    return this.request(`/audit?${params}`);
  }

  // Export
  async exportData(type: string, format: 'csv' | 'pdf', filters?: Filter[]): Promise<Blob> {
    const params = new URLSearchParams({ type, format });
    if (filters) params.set('filters', JSON.stringify(filters));

    const response = await fetch(`${this.baseURL}/export?${params}`, {
      headers: this.token ? { 'Authorization': `Bearer ${this.token}` } : {},
    });

    if (!response.ok) {
      throw new Error('Export failed');
    }

    return response.blob();
  }
}

export const api = new APIService();
