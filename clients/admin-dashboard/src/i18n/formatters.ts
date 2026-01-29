import type { SupportedLanguage } from './config';

/**
 * Number formatting utilities with locale support
 */
export class NumberFormatter {
  private locale: string;

  constructor(locale: string) {
    this.locale = locale;
  }

  /**
   * Format a number with locale-specific separators
   */
  format(value: number, options?: Intl.NumberFormatOptions): string {
    return new Intl.NumberFormat(this.locale, options).format(value);
  }

  /**
   * Format as currency
   */
  currency(value: number, currency: string = 'USD'): string {
    return new Intl.NumberFormat(this.locale, {
      style: 'currency',
      currency,
      minimumFractionDigits: 2,
      maximumFractionDigits: 2,
    }).format(value);
  }

  /**
   * Format as percentage
   */
  percentage(value: number, decimals: number = 2): string {
    return new Intl.NumberFormat(this.locale, {
      style: 'percent',
      minimumFractionDigits: decimals,
      maximumFractionDigits: decimals,
    }).format(value / 100);
  }

  /**
   * Format with compact notation (1.2K, 1.5M, etc.)
   */
  compact(value: number): string {
    return new Intl.NumberFormat(this.locale, {
      notation: 'compact',
      compactDisplay: 'short',
    }).format(value);
  }

  /**
   * Format decimal number
   */
  decimal(value: number, decimals: number = 2): string {
    return new Intl.NumberFormat(this.locale, {
      minimumFractionDigits: decimals,
      maximumFractionDigits: decimals,
    }).format(value);
  }
}

/**
 * Date/time formatting utilities with locale support
 */
export class DateFormatter {
  private locale: string;

  constructor(locale: string) {
    this.locale = locale;
  }

  /**
   * Format date (short format)
   */
  date(date: Date | number | string): string {
    return new Intl.DateTimeFormat(this.locale, {
      year: 'numeric',
      month: '2-digit',
      day: '2-digit',
    }).format(new Date(date));
  }

  /**
   * Format time (short format)
   */
  time(date: Date | number | string): string {
    return new Intl.DateTimeFormat(this.locale, {
      hour: '2-digit',
      minute: '2-digit',
      second: '2-digit',
      hour12: this.locale.startsWith('en-US'),
    }).format(new Date(date));
  }

  /**
   * Format date and time
   */
  dateTime(date: Date | number | string): string {
    return new Intl.DateTimeFormat(this.locale, {
      year: 'numeric',
      month: '2-digit',
      day: '2-digit',
      hour: '2-digit',
      minute: '2-digit',
      second: '2-digit',
      hour12: this.locale.startsWith('en-US'),
    }).format(new Date(date));
  }

  /**
   * Format relative time (e.g., "2 hours ago")
   */
  relative(date: Date | number | string): string {
    const now = new Date();
    const target = new Date(date);
    const diffInSeconds = Math.floor((now.getTime() - target.getTime()) / 1000);

    const rtf = new Intl.RelativeTimeFormat(this.locale, { numeric: 'auto' });

    if (Math.abs(diffInSeconds) < 60) {
      return rtf.format(-diffInSeconds, 'second');
    } else if (Math.abs(diffInSeconds) < 3600) {
      return rtf.format(-Math.floor(diffInSeconds / 60), 'minute');
    } else if (Math.abs(diffInSeconds) < 86400) {
      return rtf.format(-Math.floor(diffInSeconds / 3600), 'hour');
    } else if (Math.abs(diffInSeconds) < 2592000) {
      return rtf.format(-Math.floor(diffInSeconds / 86400), 'day');
    } else if (Math.abs(diffInSeconds) < 31536000) {
      return rtf.format(-Math.floor(diffInSeconds / 2592000), 'month');
    } else {
      return rtf.format(-Math.floor(diffInSeconds / 31536000), 'year');
    }
  }

  /**
   * Format long date (e.g., "January 1, 2024")
   */
  longDate(date: Date | number | string): string {
    return new Intl.DateTimeFormat(this.locale, {
      year: 'numeric',
      month: 'long',
      day: 'numeric',
    }).format(new Date(date));
  }

  /**
   * Format weekday
   */
  weekday(date: Date | number | string): string {
    return new Intl.DateTimeFormat(this.locale, {
      weekday: 'long',
    }).format(new Date(date));
  }
}

/**
 * Pluralization helper
 */
export class PluralFormatter {
  private locale: string;

  constructor(locale: string) {
    this.locale = locale;
  }

  /**
   * Get plural form (zero, one, two, few, many, other)
   */
  select(count: number): Intl.LDMLPluralRule {
    const pr = new Intl.PluralRules(this.locale);
    return pr.select(count);
  }

  /**
   * Format with plural rules
   */
  format(
    count: number,
    forms: {
      zero?: string;
      one?: string;
      two?: string;
      few?: string;
      many?: string;
      other: string;
    }
  ): string {
    const rule = this.select(count);
    return forms[rule] || forms.other;
  }
}

/**
 * List formatting utilities
 */
export class ListFormatter {
  private locale: string;

  constructor(locale: string) {
    this.locale = locale;
  }

  /**
   * Format list with conjunctions (and/or)
   */
  format(items: string[], type: 'conjunction' | 'disjunction' = 'conjunction'): string {
    return new Intl.ListFormat(this.locale, {
      style: 'long',
      type,
    }).format(items);
  }
}

/**
 * Create formatters for a specific locale
 */
export const createFormatters = (locale: SupportedLanguage) => ({
  number: new NumberFormatter(locale),
  date: new DateFormatter(locale),
  plural: new PluralFormatter(locale),
  list: new ListFormatter(locale),
});
