/**
 * Trading Terminal Type Definitions
 * Comprehensive type system for professional trading interface
 */

// Market Data Types
export type MarketDepthEntry = {
  price: number;
  volume: number;
  cumulative: number;
  orders?: number;
};

export type MarketDepth = {
  symbol: string;
  bids: MarketDepthEntry[];
  asks: MarketDepthEntry[];
  spread: number;
  spreadPips: number;
  timestamp: number;
};

export type TimeSalesEntry = {
  id: string;
  symbol: string;
  price: number;
  volume: number;
  side: 'BUY' | 'SELL';
  timestamp: number;
  aggressor?: 'BUYER' | 'SELLER';
};

export type MarketWatchItem = {
  symbol: string;
  description?: string;
  bid: number;
  ask: number;
  last: number;
  change: number;
  changePercent: number;
  volume: number;
  high24h: number;
  low24h: number;
  timestamp: number;
  direction?: 'up' | 'down' | 'neutral';
};

// Alert Types
export type AlertType = 'PRICE' | 'POSITION' | 'ACCOUNT' | 'SYSTEM' | 'NEWS';
export type AlertPriority = 'LOW' | 'MEDIUM' | 'HIGH' | 'CRITICAL';
export type AlertStatus = 'ACTIVE' | 'TRIGGERED' | 'EXPIRED' | 'CANCELLED';

export type PriceAlert = {
  id: string;
  symbol: string;
  type: 'PRICE';
  condition: 'ABOVE' | 'BELOW' | 'CROSSES_UP' | 'CROSSES_DOWN';
  targetPrice: number;
  currentPrice?: number;
  status: AlertStatus;
  priority: AlertPriority;
  createdAt: string;
  triggeredAt?: string;
  message?: string;
};

export type Alert = PriceAlert | {
  id: string;
  type: Exclude<AlertType, 'PRICE'>;
  title: string;
  message: string;
  status: AlertStatus;
  priority: AlertPriority;
  createdAt: string;
  triggeredAt?: string;
};

export type Notification = {
  id: string;
  type: 'SUCCESS' | 'ERROR' | 'WARNING' | 'INFO';
  title: string;
  message: string;
  timestamp: number;
  duration?: number;
  action?: {
    label: string;
    onClick: () => void;
  };
};

// Chart Types
export type DrawingTool =
  | 'TREND_LINE'
  | 'HORIZONTAL_LINE'
  | 'VERTICAL_LINE'
  | 'FIBONACCI'
  | 'SUPPORT_RESISTANCE'
  | 'CHANNEL'
  | 'RECTANGLE'
  | 'TRIANGLE'
  | 'TEXT_NOTE';

export type Drawing = {
  id: string;
  type: DrawingTool;
  points: { time: number; price: number }[];
  style: {
    color: string;
    lineWidth: number;
    lineStyle: 'solid' | 'dashed' | 'dotted';
  };
  locked?: boolean;
  visible?: boolean;
};

export type TechnicalIndicator = {
  id: string;
  name: string;
  type: 'OVERLAY' | 'OSCILLATOR' | 'VOLUME';
  parameters: Record<string, number | string>;
  visible: boolean;
  style: {
    colors: string[];
    lineWidth: number;
  };
};

export type ChartSettings = {
  showGrid: boolean;
  showVolume: boolean;
  showCrosshair: boolean;
  priceScale: 'AUTO' | 'LOGARITHMIC' | 'PERCENTAGE';
  backgroundColor: string;
  gridColor: string;
  textColor: string;
};

// Order Types (Extended)
export type OrderType = 'MARKET' | 'LIMIT' | 'STOP' | 'STOP_LIMIT' | 'TRAILING_STOP';
export type OrderSide = 'BUY' | 'SELL';
export type OrderStatus = 'PENDING' | 'PARTIALLY_FILLED' | 'FILLED' | 'CANCELLED' | 'REJECTED' | 'EXPIRED';
export type TimeInForce = 'GTC' | 'IOC' | 'FOK' | 'DAY';

export type BaseOrder = {
  id: string;
  symbol: string;
  accountId: string;
  side: OrderSide;
  volume: number;
  filledVolume: number;
  status: OrderStatus;
  timeInForce: TimeInForce;
  sl?: number;
  tp?: number;
  createdAt: string;
  updatedAt: string;
  filledAt?: string;
};

export type MarketOrder = BaseOrder & {
  type: 'MARKET';
  executionPrice?: number;
};

export type LimitOrder = BaseOrder & {
  type: 'LIMIT';
  limitPrice: number;
};

export type StopOrder = BaseOrder & {
  type: 'STOP';
  stopPrice: number;
  executionPrice?: number;
};

export type StopLimitOrder = BaseOrder & {
  type: 'STOP_LIMIT';
  stopPrice: number;
  limitPrice: number;
};

export type TrailingStopOrder = BaseOrder & {
  type: 'TRAILING_STOP';
  trailingDistance: number;
  trailingDistanceType: 'PIPS' | 'PERCENT' | 'PRICE';
  currentStopPrice: number;
};

export type Order = MarketOrder | LimitOrder | StopOrder | StopLimitOrder | TrailingStopOrder;

// Position Types (Extended)
export type PositionSide = 'LONG' | 'SHORT';

export type Position = {
  id: string;
  symbol: string;
  accountId: string;
  side: PositionSide;
  volume: number;
  openPrice: number;
  currentPrice: number;
  unrealizedPnL: number;
  realizedPnL: number;
  commission: number;
  swap: number;
  netPnL: number;
  sl?: number;
  tp?: number;
  openTime: string;
  updateTime: string;
  comment?: string;
};

// Account Types (Extended)
export type AccountMetrics = {
  balance: number;
  equity: number;
  margin: number;
  freeMargin: number;
  marginLevel: number;
  unrealizedPnL: number;
  realizedPnL: number;
  todayPnL: number;
  weekPnL: number;
  monthPnL: number;
  totalTrades: number;
  winningTrades: number;
  losingTrades: number;
  winRate: number;
  profitFactor: number;
  sharpeRatio?: number;
  maxDrawdown: number;
  currency: string;
  leverage: number;
};

export type EquityCurvePoint = {
  timestamp: number;
  equity: number;
  balance: number;
  drawdown: number;
};

export type TradeStatistics = {
  totalTrades: number;
  winningTrades: number;
  losingTrades: number;
  winRate: number;
  avgWin: number;
  avgLoss: number;
  profitFactor: number;
  largestWin: number;
  largestLoss: number;
  avgHoldingTime: number;
  consecutiveWins: number;
  consecutiveLosses: number;
};

// Trade History
export type Trade = {
  id: string;
  symbol: string;
  accountId: string;
  side: OrderSide;
  volume: number;
  openPrice: number;
  closePrice: number;
  openTime: string;
  closeTime: string;
  commission: number;
  swap: number;
  profit: number;
  netProfit: number;
  pips: number;
  duration: number;
  comment?: string;
};

// WebSocket Message Types
export type WSMessageType =
  | 'TICK'
  | 'DEPTH'
  | 'TIME_SALES'
  | 'POSITION_UPDATE'
  | 'ORDER_UPDATE'
  | 'ACCOUNT_UPDATE'
  | 'ALERT_TRIGGERED'
  | 'SYSTEM_MESSAGE';

export type WSMessage = {
  type: WSMessageType;
  timestamp: number;
  data: unknown;
};

export type TickMessage = WSMessage & {
  type: 'TICK';
  data: {
    symbol: string;
    bid: number;
    ask: number;
    last: number;
    volume: number;
    timestamp: number;
  };
};

export type DepthMessage = WSMessage & {
  type: 'DEPTH';
  data: MarketDepth;
};

export type TimeSalesMessage = WSMessage & {
  type: 'TIME_SALES';
  data: TimeSalesEntry;
};

// UI State Types
export type PanelLayout = {
  marketWatch: { visible: boolean; width: number };
  orderBook: { visible: boolean; width: number };
  timeSales: { visible: boolean; width: number };
  orderEntry: { visible: boolean; height: number };
  chart: { visible: boolean; maximized: boolean };
  positions: { visible: boolean; height: number };
  alerts: { visible: boolean; height: number };
};

export type ThemeColors = {
  background: string;
  surface: string;
  border: string;
  textPrimary: string;
  textSecondary: string;
  textMuted: string;
  success: string;
  danger: string;
  warning: string;
  info: string;
  accent: string;
  buy: string;
  sell: string;
  bidColor: string;
  askColor: string;
};

// Context Menu Types
export type ContextMenuItem = {
  label: string;
  icon?: React.ReactNode;
  onClick: () => void;
  disabled?: boolean;
  divider?: boolean;
  shortcut?: string;
};

export type ContextMenuPosition = {
  x: number;
  y: number;
};

// Utility Types
export type SortDirection = 'asc' | 'desc';
export type SortColumn = string;

export type SortConfig = {
  column: SortColumn;
  direction: SortDirection;
};

// Filter Types
export type SymbolFilter = {
  searchTerm: string;
  categories?: string[];
  favorites?: boolean;
};

// Keyboard Shortcut Types
export type KeyboardShortcut = {
  key: string;
  ctrl?: boolean;
  shift?: boolean;
  alt?: boolean;
  action: string;
  description: string;
};
