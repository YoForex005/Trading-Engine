/**
 * i18n Module Entry Point
 *
 * Exports all internationalization functionality
 */

export { default as i18n, SUPPORTED_LANGUAGES, type SupportedLanguage } from './config';
export { useI18n, useCurrency, useDate, useNumber, usePlural } from './hooks';
export {
  NumberFormatter,
  DateFormatter,
  PluralFormatter,
  ListFormatter,
  createFormatters
} from './formatters';
export { LanguageSelector } from '../components/LanguageSelector';
export {
  pseudoLocalize,
  createPseudoTranslations,
  detectHardcodedStrings,
  enableRTLTesting,
  testCharacterEncoding
} from './pseudo';
