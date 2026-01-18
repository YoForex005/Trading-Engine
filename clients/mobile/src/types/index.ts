// Core types
export type Currency = 'USD' | 'EUR' | 'GBP' | 'BTC' | 'ETH';
export type OrderSide = 'BUY' | 'SELL';
export type OrderType = 'MARKET' | 'LIMIT' | 'STOP' | 'STOP_LIMIT';
export type OrderStatus = 'PENDING' | 'FILLED' | 'PARTIAL' | 'CANCELLED' | 'REJECTED';
export type TimeInForce = 'GTC' | 'IOC' | 'FOK' | 'DAY';

export interface User {
  id: string;
  email: string;
  firstName: string;
  lastName: string;
  phoneNumber?: string;
  kycStatus: 'PENDING' | 'APPROVED' | 'REJECTED';
  twoFactorEnabled: boolean;
  biometricEnabled: boolean;
  createdAt: string;
}

export interface Account {
  id: string;
  userId: string;
  accountType: 'DEMO' | 'LIVE';
  currency: Currency;
  balance: number;
  equity: number;
  margin: number;
  freeMargin: number;
  marginLevel: number;
  leverage: number;
  profit: number;
}

export interface Position {
  id: string;
  accountId: string;
  symbol: string;
  side: OrderSide;
  volume: number;
  openPrice: number;
  currentPrice: number;
  stopLoss?: number;
  takeProfit?: number;
  profit: number;
  profitPercent: number;
  openedAt: string;
  commission: number;
  swap: number;
}

export interface Order {
  id: string;
  accountId: string;
  symbol: string;
  side: OrderSide;
  type: OrderType;
  volume: number;
  price?: number;
  stopPrice?: number;
  stopLoss?: number;
  takeProfit?: number;
  timeInForce: TimeInForce;
  status: OrderStatus;
  filledVolume: number;
  remainingVolume: number;
  createdAt: string;
  updatedAt: string;
  commission: number;
}

export interface Ticker {
  symbol: string;
  bid: number;
  ask: number;
  last: number;
  high24h: number;
  low24h: number;
  volume24h: number;
  change24h: number;
  changePercent24h: number;
  timestamp: string;
}

export interface Candle {
  time: number;
  open: number;
  high: number;
  low: number;
  close: number;
  volume: number;
}

export interface Trade {
  id: string;
  orderId: string;
  symbol: string;
  side: OrderSide;
  volume: number;
  price: number;
  commission: number;
  profit?: number;
  executedAt: string;
}

export interface Alert {
  id: string;
  symbol: string;
  condition: 'ABOVE' | 'BELOW';
  price: number;
  active: boolean;
  triggered: boolean;
  createdAt: string;
}

export interface Notification {
  id: string;
  type: 'ORDER_FILLED' | 'MARGIN_CALL' | 'ALERT_TRIGGERED' | 'NEWS' | 'SYSTEM';
  title: string;
  message: string;
  read: boolean;
  createdAt: string;
  data?: any;
}

export interface NewsItem {
  id: string;
  title: string;
  summary: string;
  content: string;
  source: string;
  category: string;
  sentiment: 'POSITIVE' | 'NEGATIVE' | 'NEUTRAL';
  publishedAt: string;
  imageUrl?: string;
  url: string;
}

export interface Deposit {
  id: string;
  accountId: string;
  amount: number;
  currency: Currency;
  method: 'BANK_TRANSFER' | 'CREDIT_CARD' | 'CRYPTO';
  status: 'PENDING' | 'COMPLETED' | 'FAILED';
  createdAt: string;
  completedAt?: string;
}

export interface Withdrawal {
  id: string;
  accountId: string;
  amount: number;
  currency: Currency;
  method: 'BANK_TRANSFER' | 'CRYPTO';
  status: 'PENDING' | 'PROCESSING' | 'COMPLETED' | 'REJECTED';
  createdAt: string;
  completedAt?: string;
  rejectionReason?: string;
}

export interface ChartData {
  symbol: string;
  timeframe: '1m' | '5m' | '15m' | '1h' | '4h' | '1d';
  candles: Candle[];
}

// Navigation types
export type RootStackParamList = {
  Auth: undefined;
  Main: undefined;
  Login: undefined;
  Register: undefined;
  Dashboard: undefined;
  Trading: { symbol?: string };
  Chart: { symbol: string };
  Positions: undefined;
  Orders: undefined;
  History: undefined;
  Account: undefined;
  Deposits: undefined;
  Withdrawals: undefined;
  Settings: undefined;
  Notifications: undefined;
  Alerts: undefined;
};

// API Response types
export interface ApiResponse<T> {
  success: boolean;
  data?: T;
  error?: string;
  message?: string;
}

export interface LoginRequest {
  email: string;
  password: string;
}

export interface LoginResponse {
  user: User;
  token: string;
  refreshToken: string;
}

export interface RegisterRequest {
  email: string;
  password: string;
  firstName: string;
  lastName: string;
  phoneNumber?: string;
}

export interface PlaceOrderRequest {
  symbol: string;
  side: OrderSide;
  type: OrderType;
  volume: number;
  price?: number;
  stopPrice?: number;
  stopLoss?: number;
  takeProfit?: number;
  timeInForce?: TimeInForce;
}
