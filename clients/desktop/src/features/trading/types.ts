import type { Time } from 'lightweight-charts';

export type ChartType = 'candlestick' | 'heikinAshi' | 'bar' | 'line' | 'area';
export type Timeframe = '1m' | '5m' | '15m' | '1h' | '4h' | '1d';

export type OHLCData = {
  time: Time;
  open: number;
  high: number;
  low: number;
  close: number;
};

export type TickData = {
  symbol: string;
  bid: number;
  ask: number;
  timestamp: number;
};

export type Position = {
  id: number;
  symbol: string;
  type: string;
  volume: number;
  openPrice: number;
  currentPrice: number;
  profit: number;
  stopLoss?: number;
  takeProfit?: number;
};

export type Order = {
  id: number;
  symbol: string;
  type: string;
  triggerPrice: number;
  volume: number;
  stopLoss?: number;
  takeProfit?: number;
  ocoLinkId?: number;
  expiryTime?: number;
};

export type Drawing = {
  id: string;
  type: string;
  accountId: number;
  symbol: string;
  points: Array<{ time: number; price: number }>;
  color?: string;
  lineWidth?: number;
};
