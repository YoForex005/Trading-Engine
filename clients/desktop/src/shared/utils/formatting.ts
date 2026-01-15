/**
 * Formatting Utilities - Shared formatting functions
 * Eliminates duplicated formatting logic
 */

/**
 * Formats a number as currency
 */
export function formatCurrency(value: number, currency = 'USD', decimals = 2): string {
  return new Intl.NumberFormat('en-US', {
    style: 'currency',
    currency,
    minimumFractionDigits: decimals,
    maximumFractionDigits: decimals,
  }).format(value);
}

/**
 * Formats a number with specified decimal places
 */
export function formatNumber(value: number, decimals = 2): string {
  return new Intl.NumberFormat('en-US', {
    minimumFractionDigits: decimals,
    maximumFractionDigits: decimals,
  }).format(value);
}

/**
 * Formats a number as percentage
 */
export function formatPercent(value: number, decimals = 2): string {
  return new Intl.NumberFormat('en-US', {
    style: 'percent',
    minimumFractionDigits: decimals,
    maximumFractionDigits: decimals,
  }).format(value / 100);
}

/**
 * Formats a price with appropriate precision based on asset type
 */
export function formatPrice(price: number, symbol: string): string {
  // Crypto typically needs more precision
  if (symbol.includes('BTC') || symbol.includes('ETH')) {
    return formatNumber(price, 2);
  }
  // Forex pairs (JPY pairs typically 3 decimals, others 5)
  if (symbol.includes('JPY')) {
    return formatNumber(price, 3);
  }
  // Default forex precision
  return formatNumber(price, 5);
}

/**
 * Formats a timestamp as a readable date/time
 */
export function formatDateTime(timestamp: number | string | Date): string {
  const date = new Date(timestamp);
  return new Intl.DateTimeFormat('en-US', {
    year: 'numeric',
    month: 'short',
    day: 'numeric',
    hour: '2-digit',
    minute: '2-digit',
    second: '2-digit',
  }).format(date);
}

/**
 * Formats a timestamp as a readable date
 */
export function formatDate(timestamp: number | string | Date): string {
  const date = new Date(timestamp);
  return new Intl.DateTimeFormat('en-US', {
    year: 'numeric',
    month: 'short',
    day: 'numeric',
  }).format(date);
}

/**
 * Formats a timestamp as a readable time
 */
export function formatTime(timestamp: number | string | Date): string {
  const date = new Date(timestamp);
  return new Intl.DateTimeFormat('en-US', {
    hour: '2-digit',
    minute: '2-digit',
    second: '2-digit',
  }).format(date);
}

/**
 * Formats a large number with K/M/B suffixes
 */
export function formatCompactNumber(value: number): string {
  return new Intl.NumberFormat('en-US', {
    notation: 'compact',
    compactDisplay: 'short',
  }).format(value);
}

/**
 * Formats volume/lot size
 */
export function formatVolume(volume: number): string {
  return formatNumber(volume, 2);
}

/**
 * Adds color class based on positive/negative value
 */
export function getColorClass(value: number): string {
  if (value > 0) return 'text-green-600';
  if (value < 0) return 'text-red-600';
  return 'text-gray-600';
}

/**
 * Formats profit/loss with sign and color
 */
export function formatPnL(value: number, currency = 'USD'): {
  text: string;
  colorClass: string;
} {
  const sign = value >= 0 ? '+' : '';
  return {
    text: sign + formatCurrency(value, currency),
    colorClass: getColorClass(value),
  };
}
