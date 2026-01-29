// System Metrics
export type SystemMetrics = {
  timestamp: number;
  cpu: {
    usage: number;
    cores: number;
  };
  memory: {
    used: number;
    total: number;
    percentage: number;
  };
  connections: {
    active: number;
    total: number;
  };
  throughput: {
    ordersPerSecond: number;
    messagesPerSecond: number;
  };
};

// Order Flow
export type OrderStatus = 'pending' | 'filled' | 'partial' | 'rejected' | 'cancelled';
export type OrderSide = 'buy' | 'sell';
export type OrderType = 'market' | 'limit' | 'stop' | 'stop_limit';

export type Order = {
  id: string;
  userId: string;
  symbol: string;
  side: OrderSide;
  type: OrderType;
  quantity: number;
  price?: number;
  stopPrice?: number;
  status: OrderStatus;
  filledQuantity: number;
  averagePrice: number;
  timestamp: number;
  lpRoute?: string;
};

// LP Health
export type LPStatus = 'connected' | 'disconnected' | 'degraded';

export type LiquidityProvider = {
  id: string;
  name: string;
  status: LPStatus;
  latency: number;
  uptime: number;
  lastHeartbeat: number;
  activeSymbols: string[];
  metrics: {
    ordersRouted: number;
    fillRate: number;
    rejectionRate: number;
    averageLatency: number;
  };
};

// User Activity
export type UserSession = {
  userId: string;
  username: string;
  sessionId: string;
  loginTime: number;
  lastActivity: number;
  ipAddress: string;
  activeOrders: number;
  totalOrders: number;
  pnl: number;
};

// Error Tracking
export type ErrorSeverity = 'info' | 'warning' | 'error' | 'critical';

export type ErrorLog = {
  id: string;
  timestamp: number;
  severity: ErrorSeverity;
  component: string;
  message: string;
  stack?: string;
  userId?: string;
  orderId?: string;
  metadata?: Record<string, unknown>;
};

// Routing Configuration
export type RoutingMode = 'a-book' | 'b-book' | 'c-book';

export type RoutingRule = {
  id: string;
  name: string;
  mode: RoutingMode;
  priority: number;
  conditions: {
    userGroups?: string[];
    symbols?: string[];
    minVolume?: number;
    maxVolume?: number;
    timeOfDay?: { start: string; end: string };
  };
  lpTargets?: string[];
  enabled: boolean;
};

// Symbol Configuration
export type TradingSymbol = {
  id: string;
  symbol: string;
  name: string;
  enabled: boolean;
  baseAsset: string;
  quoteAsset: string;
  minQuantity: number;
  maxQuantity: number;
  tickSize: number;
  lotSize: number;
  margin: number;
  tradingHours?: {
    open: string;
    close: string;
    timezone: string;
  };
  lpSources: string[];
};

// User Management
export type UserRole = 'admin' | 'trader' | 'viewer' | 'risk_manager';

export type User = {
  id: string;
  username: string;
  email: string;
  role: UserRole;
  enabled: boolean;
  createdAt: number;
  lastLogin?: number;
  permissions: string[];
  riskProfile: {
    maxLeverage: number;
    maxPositionSize: number;
    maxDailyLoss: number;
    allowedSymbols: string[];
  };
};

// Risk Control
export type RiskLimit = {
  id: string;
  name: string;
  type: 'user' | 'symbol' | 'global';
  targetId?: string;
  limits: {
    maxOpenPositions?: number;
    maxPositionSize?: number;
    maxLeverage?: number;
    maxDailyLoss?: number;
    maxDrawdown?: number;
  };
  actions: {
    onBreach: 'warn' | 'restrict' | 'close_positions' | 'disable_trading';
    notifyAdmin: boolean;
  };
  enabled: boolean;
};

// Analytics
export type PerformanceMetric = {
  timestamp: number;
  metric: string;
  value: number;
};

export type TradingAnalytics = {
  period: string;
  totalVolume: number;
  totalOrders: number;
  avgOrderSize: number;
  topSymbols: Array<{ symbol: string; volume: number }>;
  pnl: {
    gross: number;
    net: number;
    fees: number;
  };
  fillRates: {
    overall: number;
    byLP: Record<string, number>;
  };
};

export type UserAnalytics = {
  userId: string;
  username: string;
  stats: {
    totalOrders: number;
    totalVolume: number;
    avgOrderSize: number;
    winRate: number;
    pnl: number;
    sharpeRatio?: number;
  };
  topSymbols: Array<{ symbol: string; count: number }>;
  activityTrend: Array<{ date: string; orders: number }>;
};

// Audit Trail
export type AuditAction = 'create' | 'update' | 'delete' | 'login' | 'logout' | 'execute' | 'configure';

export type AuditLog = {
  id: string;
  timestamp: number;
  userId: string;
  username: string;
  action: AuditAction;
  resource: string;
  resourceId?: string;
  changes?: Record<string, { before: unknown; after: unknown }>;
  ipAddress: string;
  userAgent: string;
};

// WebSocket Messages
export type WSMessageType =
  | 'system_metrics'
  | 'order_update'
  | 'lp_status'
  | 'user_activity'
  | 'error_log'
  | 'alert';

export type WSMessage<T = unknown> = {
  type: WSMessageType;
  timestamp: number;
  data: T;
};

// Alerts
export type AlertType = 'info' | 'warning' | 'error' | 'critical';

export type Alert = {
  id: string;
  timestamp: number;
  type: AlertType;
  title: string;
  message: string;
  acknowledged: boolean;
  acknowledgedBy?: string;
  acknowledgedAt?: number;
};

// API Response
export type APIResponse<T = unknown> = {
  success: boolean;
  data?: T;
  error?: string;
  timestamp: number;
};

// Table Filters
export type FilterOperator = 'eq' | 'ne' | 'gt' | 'gte' | 'lt' | 'lte' | 'contains' | 'in';

export type Filter = {
  field: string;
  operator: FilterOperator;
  value: unknown;
};

export type SortDirection = 'asc' | 'desc';

export type Sort = {
  field: string;
  direction: SortDirection;
};

export type Pagination = {
  page: number;
  pageSize: number;
  total: number;
};
