import { render, RenderOptions } from '@testing-library/react';
import { ReactElement } from 'react';

/**
 * Custom render function that wraps components with common providers
 * Add providers here as needed (Router, Theme, etc.)
 */
export function renderWithProviders(
  ui: ReactElement,
  options?: Omit<RenderOptions, 'wrapper'>
) {
  return render(ui, { ...options });
}

/**
 * Creates mock market data for testing charts
 */
export function createMockOHLC(count: number = 10) {
  const data = [];
  const baseTime = new Date('2026-01-16T00:00:00Z').getTime();

  for (let i = 0; i < count; i++) {
    const time = baseTime + i * 60000; // 1 minute intervals
    const open = 1.085 + Math.random() * 0.01;
    const close = open + (Math.random() - 0.5) * 0.005;
    const high = Math.max(open, close) + Math.random() * 0.002;
    const low = Math.min(open, close) - Math.random() * 0.002;

    data.push({
      time: new Date(time).toISOString(),
      open,
      high,
      low,
      close,
      volume: Math.floor(Math.random() * 1000),
    });
  }

  return data;
}

/**
 * Creates mock tick data for testing real-time updates
 */
export function createMockTick(symbol: string, bid: number, ask: number) {
  return {
    type: 'tick',
    symbol,
    bid,
    ask,
    timestamp: new Date().toISOString(),
  };
}

/**
 * Creates mock account data for testing account displays
 */
export function createMockAccount(
  overrides?: Partial<{
    balance: number;
    equity: number;
    margin: number;
    freeMargin: number;
    marginLevel: number;
  }>
) {
  return {
    accountNumber: 'TEST001',
    balance: 10000,
    equity: 10000,
    margin: 0,
    freeMargin: 10000,
    marginLevel: 0,
    currency: 'USD',
    leverage: 100,
    ...overrides,
  };
}

// Re-export everything from Testing Library for convenience
export * from '@testing-library/react';
export { default as userEvent } from '@testing-library/user-event';
