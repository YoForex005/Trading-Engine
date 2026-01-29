import { Currency } from '@/types';

export const formatCurrency = (
  value: number,
  currency: Currency = 'USD',
  decimals: number = 2,
): string => {
  const currencySymbols: Record<Currency, string> = {
    USD: '$',
    EUR: '€',
    GBP: '£',
    BTC: '₿',
    ETH: 'Ξ',
  };

  const symbol = currencySymbols[currency] || currency;
  const formattedValue = value.toFixed(decimals).replace(/\B(?=(\d{3})+(?!\d))/g, ',');

  return `${symbol}${formattedValue}`;
};

export const formatPercent = (value: number, decimals: number = 2): string => {
  const sign = value >= 0 ? '+' : '';
  return `${sign}${value.toFixed(decimals)}%`;
};

export const formatNumber = (value: number, decimals: number = 2): string => {
  return value.toFixed(decimals).replace(/\B(?=(\d{3})+(?!\d))/g, ',');
};

export const formatVolume = (value: number): string => {
  if (value >= 1000000) {
    return `${(value / 1000000).toFixed(2)}M`;
  }
  if (value >= 1000) {
    return `${(value / 1000).toFixed(2)}K`;
  }
  return value.toFixed(2);
};

export const formatDate = (date: string | Date): string => {
  const d = typeof date === 'string' ? new Date(date) : date;
  return d.toLocaleDateString('en-US', {
    year: 'numeric',
    month: 'short',
    day: 'numeric',
  });
};

export const formatDateTime = (date: string | Date): string => {
  const d = typeof date === 'string' ? new Date(date) : date;
  return d.toLocaleString('en-US', {
    year: 'numeric',
    month: 'short',
    day: 'numeric',
    hour: '2-digit',
    minute: '2-digit',
  });
};

export const formatTime = (date: string | Date): string => {
  const d = typeof date === 'string' ? new Date(date) : date;
  return d.toLocaleTimeString('en-US', {
    hour: '2-digit',
    minute: '2-digit',
  });
};

export const formatRelativeTime = (date: string | Date): string => {
  const d = typeof date === 'string' ? new Date(date) : date;
  const now = new Date();
  const diffMs = now.getTime() - d.getTime();
  const diffSecs = Math.floor(diffMs / 1000);
  const diffMins = Math.floor(diffSecs / 60);
  const diffHours = Math.floor(diffMins / 60);
  const diffDays = Math.floor(diffHours / 24);

  if (diffSecs < 60) {
    return 'Just now';
  }
  if (diffMins < 60) {
    return `${diffMins}m ago`;
  }
  if (diffHours < 24) {
    return `${diffHours}h ago`;
  }
  if (diffDays < 7) {
    return `${diffDays}d ago`;
  }
  return formatDate(d);
};

export const formatPrice = (price: number, symbol: string): string => {
  // Determine decimals based on price magnitude
  let decimals = 2;
  if (price < 0.01) {
    decimals = 8;
  } else if (price < 1) {
    decimals = 4;
  } else if (price >= 1000) {
    decimals = 1;
  }

  return price.toFixed(decimals);
};

export const abbreviateNumber = (num: number): string => {
  if (num >= 1e9) {
    return (num / 1e9).toFixed(1) + 'B';
  }
  if (num >= 1e6) {
    return (num / 1e6).toFixed(1) + 'M';
  }
  if (num >= 1e3) {
    return (num / 1e3).toFixed(1) + 'K';
  }
  return num.toFixed(0);
};
