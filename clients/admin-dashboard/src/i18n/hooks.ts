import { useTranslation } from 'react-i18next';
import { useCallback, useMemo } from 'react';
import type { SupportedLanguage } from './config';
import { SUPPORTED_LANGUAGES } from './config';
import { createFormatters } from './formatters';

/**
 * Enhanced i18n hook with formatters
 */
export const useI18n = () => {
  const { t, i18n } = useTranslation();

  const currentLanguage = i18n.language as SupportedLanguage;
  const languageInfo = SUPPORTED_LANGUAGES[currentLanguage];

  const formatters = useMemo(
    () => createFormatters(currentLanguage),
    [currentLanguage]
  );

  const changeLanguage = useCallback(
    (lng: SupportedLanguage) => {
      i18n.changeLanguage(lng);

      // Update HTML attributes
      document.documentElement.lang = lng;
      document.documentElement.dir = SUPPORTED_LANGUAGES[lng].rtl ? 'rtl' : 'ltr';

      // Store preference
      localStorage.setItem('i18nextLng', lng);
    },
    [i18n]
  );

  const isRTL = useMemo(
    () => languageInfo?.rtl || false,
    [languageInfo]
  );

  return {
    t,
    i18n,
    currentLanguage,
    languageInfo,
    formatters,
    changeLanguage,
    isRTL,
    languages: SUPPORTED_LANGUAGES,
  };
};

/**
 * Hook for currency formatting
 */
export const useCurrency = () => {
  const { formatters } = useI18n();

  return useCallback(
    (value: number, currency: string = 'USD') =>
      formatters.number.currency(value, currency),
    [formatters]
  );
};

/**
 * Hook for date formatting
 */
export const useDate = () => {
  const { formatters } = useI18n();

  return useMemo(
    () => ({
      date: (d: Date | number | string) => formatters.date.date(d),
      time: (d: Date | number | string) => formatters.date.time(d),
      dateTime: (d: Date | number | string) => formatters.date.dateTime(d),
      relative: (d: Date | number | string) => formatters.date.relative(d),
      longDate: (d: Date | number | string) => formatters.date.longDate(d),
      weekday: (d: Date | number | string) => formatters.date.weekday(d),
    }),
    [formatters]
  );
};

/**
 * Hook for number formatting
 */
export const useNumber = () => {
  const { formatters } = useI18n();

  return useMemo(
    () => ({
      format: (n: number, opts?: Intl.NumberFormatOptions) =>
        formatters.number.format(n, opts),
      decimal: (n: number, decimals?: number) =>
        formatters.number.decimal(n, decimals),
      percentage: (n: number, decimals?: number) =>
        formatters.number.percentage(n, decimals),
      compact: (n: number) => formatters.number.compact(n),
    }),
    [formatters]
  );
};

/**
 * Hook for pluralization
 */
export const usePlural = () => {
  const { formatters } = useI18n();

  return useCallback(
    (
      count: number,
      forms: {
        zero?: string;
        one?: string;
        two?: string;
        few?: string;
        many?: string;
        other: string;
      }
    ) => formatters.plural.format(count, forms),
    [formatters]
  );
};
