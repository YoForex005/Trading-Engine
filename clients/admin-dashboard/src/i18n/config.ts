import i18n from 'i18next';
import { initReactI18next } from 'react-i18next';
import LanguageDetector from 'i18next-browser-languagedetector';
import Backend from 'i18next-http-backend';

// Supported languages
export const SUPPORTED_LANGUAGES = {
  'en-US': {
    code: 'en-US',
    name: 'English',
    nativeName: 'English',
    flag: '🇺🇸',
    rtl: false,
  },
  'es-ES': {
    code: 'es-ES',
    name: 'Spanish',
    nativeName: 'Español',
    flag: '🇪🇸',
    rtl: false,
  },
  'fr-FR': {
    code: 'fr-FR',
    name: 'French',
    nativeName: 'Français',
    flag: '🇫🇷',
    rtl: false,
  },
  'de-DE': {
    code: 'de-DE',
    name: 'German',
    nativeName: 'Deutsch',
    flag: '🇩🇪',
    rtl: false,
  },
  'ja-JP': {
    code: 'ja-JP',
    name: 'Japanese',
    nativeName: '日本語',
    flag: '🇯🇵',
    rtl: false,
  },
  'zh-CN': {
    code: 'zh-CN',
    name: 'Chinese (Simplified)',
    nativeName: '简体中文',
    flag: '🇨🇳',
    rtl: false,
  },
  'ar-SA': {
    code: 'ar-SA',
    name: 'Arabic',
    nativeName: 'العربية',
    flag: '🇸🇦',
    rtl: true,
  },
  'ru-RU': {
    code: 'ru-RU',
    name: 'Russian',
    nativeName: 'Русский',
    flag: '🇷🇺',
    rtl: false,
  },
  'pt-BR': {
    code: 'pt-BR',
    name: 'Portuguese (Brazil)',
    nativeName: 'Português (Brasil)',
    flag: '🇧🇷',
    rtl: false,
  },
  'it-IT': {
    code: 'it-IT',
    name: 'Italian',
    nativeName: 'Italiano',
    flag: '🇮🇹',
    rtl: false,
  },
} as const;

export type SupportedLanguage = keyof typeof SUPPORTED_LANGUAGES;

i18n
  .use(Backend)
  .use(LanguageDetector)
  .use(initReactI18next)
  .init({
    fallbackLng: 'en-US',
    supportedLngs: Object.keys(SUPPORTED_LANGUAGES),
    defaultNS: 'common',
    ns: ['common', 'trading', 'errors', 'notifications', 'legal'],

    // Detection options
    detection: {
      order: ['localStorage', 'navigator', 'htmlTag', 'path', 'subdomain'],
      caches: ['localStorage'],
      lookupLocalStorage: 'i18nextLng',
    },

    // Backend options for dynamic imports
    backend: {
      loadPath: '/locales/{{lng}}/{{ns}}.json',
      addPath: '/locales/add/{{lng}}/{{ns}}',
    },

    interpolation: {
      escapeValue: false, // React already escapes
      formatSeparator: ',',
      format: (value, format, lng) => {
        if (format === 'uppercase') return value.toUpperCase();
        if (format === 'lowercase') return value.toLowerCase();
        if (format === 'capitalize') {
          return value.charAt(0).toUpperCase() + value.slice(1);
        }
        return value;
      },
    },

    react: {
      useSuspense: true,
      bindI18n: 'languageChanged loaded',
      bindI18nStore: 'added removed',
      transEmptyNodeValue: '',
      transSupportBasicHtmlNodes: true,
      transKeepBasicHtmlNodesFor: ['br', 'strong', 'i', 'p'],
    },

    // Debug mode (disable in production)
    debug: process.env.NODE_ENV === 'development',

    // Warn when translations are missing
    saveMissing: process.env.NODE_ENV === 'development',
    missingKeyHandler: (lngs, ns, key, fallbackValue) => {
      if (process.env.NODE_ENV === 'development') {
        console.warn(`Missing translation: ${key} in namespace ${ns}`);
      }
    },
  });

export default i18n;
