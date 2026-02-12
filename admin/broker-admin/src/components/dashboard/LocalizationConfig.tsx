import React, { useState, useRef } from 'react';
import { Globe, Download, Upload, Check, X, Eye, Calendar, Hash } from 'lucide-react';

type LanguageCode = 'en' | 'ar' | 'zh' | 'es' | 'fr' | 'ja' | 'ru' | 'pt' | 'hi' | 'ko';

interface Language {
  code: LanguageCode;
  name: string;
  nativeName: string;
  enabled: boolean;
  completionPercent: number;
  isRTL: boolean;
}

interface TranslationKey {
  id: string;
  key: string;
  category: string;
  englishValue: string;
  translations: Record<LanguageCode, string>;
}

interface LocaleFormat {
  dateFormat: string;
  timeFormat: string;
  numberFormat: 'comma' | 'period' | 'space';
  currencySymbol: string;
  currencyPosition: 'before' | 'after';
}

const LANGUAGES: Language[] = [
  { code: 'en', name: 'English', nativeName: 'English', enabled: true, completionPercent: 100, isRTL: false },
  { code: 'ar', name: 'Arabic', nativeName: 'العربية', enabled: true, completionPercent: 85, isRTL: true },
  { code: 'zh', name: 'Chinese', nativeName: '中文', enabled: true, completionPercent: 92, isRTL: false },
  { code: 'es', name: 'Spanish', nativeName: 'Español', enabled: true, completionPercent: 88, isRTL: false },
  { code: 'fr', name: 'French', nativeName: 'Français', enabled: true, completionPercent: 90, isRTL: false },
  { code: 'ja', name: 'Japanese', nativeName: '日本語', enabled: true, completionPercent: 78, isRTL: false },
  { code: 'ru', name: 'Russian', nativeName: 'Русский', enabled: true, completionPercent: 82, isRTL: false },
  { code: 'pt', name: 'Portuguese', nativeName: 'Português', enabled: true, completionPercent: 86, isRTL: false },
  { code: 'hi', name: 'Hindi', nativeName: 'हिन्दी', enabled: false, completionPercent: 45, isRTL: false },
  { code: 'ko', name: 'Korean', nativeName: '한국어', enabled: false, completionPercent: 68, isRTL: false },
];

const SAMPLE_TRANSLATIONS: TranslationKey[] = [
  {
    id: '1',
    key: 'dashboard.title',
    category: 'Dashboard',
    englishValue: 'Trading Dashboard',
    translations: {
      en: 'Trading Dashboard',
      ar: 'لوحة تحكم التداول',
      zh: '交易仪表板',
      es: 'Panel de Trading',
      fr: 'Tableau de bord de trading',
      ja: '取引ダッシュボード',
      ru: 'Торговая панель',
      pt: 'Painel de Negociação',
      hi: 'ट्रेडिंग डैशबोर्ड',
      ko: '거래 대시보드',
    },
  },
  {
    id: '2',
    key: 'account.balance',
    category: 'Account',
    englishValue: 'Account Balance',
    translations: {
      en: 'Account Balance',
      ar: 'رصيد الحساب',
      zh: '账户余额',
      es: 'Saldo de la Cuenta',
      fr: 'Solde du Compte',
      ja: '口座残高',
      ru: 'Баланс счета',
      pt: 'Saldo da Conta',
      hi: 'खाता शेष',
      ko: '계좌 잔액',
    },
  },
  {
    id: '3',
    key: 'order.buy',
    category: 'Orders',
    englishValue: 'Buy',
    translations: {
      en: 'Buy',
      ar: 'شراء',
      zh: '买入',
      es: 'Comprar',
      fr: 'Acheter',
      ja: '買い',
      ru: 'Купить',
      pt: 'Comprar',
      hi: 'खरीदें',
      ko: '매수',
    },
  },
  {
    id: '4',
    key: 'order.sell',
    category: 'Orders',
    englishValue: 'Sell',
    translations: {
      en: 'Sell',
      ar: 'بيع',
      zh: '卖出',
      es: 'Vender',
      fr: 'Vendre',
      ja: '売り',
      ru: 'Продать',
      pt: 'Vender',
      hi: 'बेचें',
      ko: '매도',
    },
  },
  {
    id: '5',
    key: 'common.save',
    category: 'Common',
    englishValue: 'Save',
    translations: {
      en: 'Save',
      ar: 'حفظ',
      zh: '保存',
      es: 'Guardar',
      fr: 'Enregistrer',
      ja: '保存',
      ru: 'Сохранить',
      pt: 'Salvar',
      hi: 'सहेजें',
      ko: '저장',
    },
  },
];

const DEFAULT_LOCALE_FORMATS: Record<LanguageCode, LocaleFormat> = {
  en: { dateFormat: 'MM/DD/YYYY', timeFormat: '12h', numberFormat: 'comma', currencySymbol: '$', currencyPosition: 'before' },
  ar: { dateFormat: 'DD/MM/YYYY', timeFormat: '12h', numberFormat: 'comma', currencySymbol: 'ر.س', currencyPosition: 'after' },
  zh: { dateFormat: 'YYYY-MM-DD', timeFormat: '24h', numberFormat: 'comma', currencySymbol: '¥', currencyPosition: 'before' },
  es: { dateFormat: 'DD/MM/YYYY', timeFormat: '24h', numberFormat: 'period', currencySymbol: '€', currencyPosition: 'after' },
  fr: { dateFormat: 'DD/MM/YYYY', timeFormat: '24h', numberFormat: 'space', currencySymbol: '€', currencyPosition: 'after' },
  ja: { dateFormat: 'YYYY年MM月DD日', timeFormat: '24h', numberFormat: 'comma', currencySymbol: '¥', currencyPosition: 'before' },
  ru: { dateFormat: 'DD.MM.YYYY', timeFormat: '24h', numberFormat: 'space', currencySymbol: '₽', currencyPosition: 'after' },
  pt: { dateFormat: 'DD/MM/YYYY', timeFormat: '24h', numberFormat: 'period', currencySymbol: 'R$', currencyPosition: 'before' },
  hi: { dateFormat: 'DD/MM/YYYY', timeFormat: '12h', numberFormat: 'comma', currencySymbol: '₹', currencyPosition: 'before' },
  ko: { dateFormat: 'YYYY년 MM월 DD일', timeFormat: '24h', numberFormat: 'comma', currencySymbol: '₩', currencyPosition: 'before' },
};

export default function LocalizationConfig() {
  const [languages, setLanguages] = useState<Language[]>(LANGUAGES);
  const [translations, setTranslations] = useState<TranslationKey[]>(SAMPLE_TRANSLATIONS);
  const [defaultLanguage, setDefaultLanguage] = useState<LanguageCode>('en');
  const [selectedLanguage, setSelectedLanguage] = useState<LanguageCode>('en');
  const [localeFormats, setLocaleFormats] = useState<Record<LanguageCode, LocaleFormat>>(DEFAULT_LOCALE_FORMATS);
  const [showPreview, setShowPreview] = useState(false);
  const fileInputRef = useRef<HTMLInputElement>(null);

  const handleToggleLanguage = (code: LanguageCode) => {
    setLanguages((prev) =>
      prev.map((lang) => (lang.code === code ? { ...lang, enabled: !lang.enabled } : lang))
    );
  };

  const handleExportTranslations = () => {
    const exportData = {
      languages: languages.filter((l) => l.enabled),
      translations: translations,
      defaultLanguage,
      localeFormats,
    };
    const blob = new Blob([JSON.stringify(exportData, null, 2)], { type: 'application/json' });
    const url = URL.createObjectURL(blob);
    const a = document.createElement('a');
    a.href = url;
    a.download = `translations_${new Date().toISOString().split('T')[0]}.json`;
    a.click();
    URL.revokeObjectURL(url);
  };

  const handleImportTranslations = (event: React.ChangeEvent<HTMLInputElement>) => {
    const file = event.target.files?.[0];
    if (!file) return;

    const reader = new FileReader();
    reader.onload = (e) => {
      try {
        const data = JSON.parse(e.target?.result as string);
        if (data.languages) setLanguages(data.languages);
        if (data.translations) setTranslations(data.translations);
        if (data.defaultLanguage) setDefaultLanguage(data.defaultLanguage);
        if (data.localeFormats) setLocaleFormats(data.localeFormats);
        alert('Translations imported successfully!');
      } catch (error) {
        alert('Failed to import translations. Invalid JSON format.');
      }
    };
    reader.readAsText(file);
  };

  const enabledLanguages = languages.filter((l) => l.enabled);
  const avgCompletion = enabledLanguages.reduce((sum, l) => sum + l.completionPercent, 0) / enabledLanguages.length;

  return (
    <div className="flex flex-col h-full bg-[#121316] text-[#E4E4E7] overflow-y-auto">
      {/* Header */}
      <div className="flex items-center justify-between px-4 py-3 bg-[#1E2026] border-b border-[#383A42] flex-shrink-0">
        <h1 className="text-base font-bold text-[#F5C542]">Multi-Language & Localization Configuration</h1>
        <div className="flex gap-2">
          <button
            onClick={() => fileInputRef.current?.click()}
            className="flex items-center gap-2 px-3 py-1.5 bg-[#25272E] hover:bg-[#2D2F35] text-[#E4E4E7] text-xs font-bold rounded transition-colors"
          >
            <Upload size={14} />
            Import JSON
          </button>
          <button
            onClick={handleExportTranslations}
            className="flex items-center gap-2 px-3 py-1.5 bg-[#3B82F6] hover:bg-[#2563EB] text-white text-xs font-bold rounded transition-colors"
          >
            <Download size={14} />
            Export JSON
          </button>
          <input
            ref={fileInputRef}
            type="file"
            accept=".json"
            onChange={handleImportTranslations}
            className="hidden"
          />
        </div>
      </div>

      <div className="flex-1 p-4 space-y-4">
        {/* Summary Cards */}
        <div className="grid grid-cols-3 gap-4">
          <div className="bg-[#1E2026] border border-[#383A42] rounded p-4">
            <div className="flex items-center gap-2 mb-2">
              <Globe size={16} className="text-[#3B82F6]" />
              <div className="text-xs text-[#888]">Enabled Languages</div>
            </div>
            <div className="text-2xl font-bold text-[#E4E4E7]">
              {enabledLanguages.length} / {languages.length}
            </div>
          </div>

          <div className="bg-[#1E2026] border border-[#383A42] rounded p-4">
            <div className="flex items-center gap-2 mb-2">
              <Check size={16} className="text-[#10b981]" />
              <div className="text-xs text-[#888]">Avg Completion</div>
            </div>
            <div className="text-2xl font-bold text-[#10b981]">{avgCompletion.toFixed(0)}%</div>
          </div>

          <div className="bg-[#1E2026] border border-[#383A42] rounded p-4">
            <div className="flex items-center gap-2 mb-2">
              <Hash size={16} className="text-[#F5C542]" />
              <div className="text-xs text-[#888]">Translation Keys</div>
            </div>
            <div className="text-2xl font-bold text-[#E4E4E7]">{translations.length}</div>
          </div>
        </div>

        {/* Language Toggles & Default Selector */}
        <div className="bg-[#1E2026] border border-[#383A42] rounded">
          <div className="px-4 py-2 bg-[#25272E] border-b border-[#383A42] flex items-center justify-between">
            <h3 className="text-sm font-bold text-[#F5C542]">Supported Languages</h3>
            <div className="flex items-center gap-2">
              <span className="text-xs text-[#888]">Default Language:</span>
              <select
                value={defaultLanguage}
                onChange={(e) => setDefaultLanguage(e.target.value as LanguageCode)}
                className="bg-[#1E2026] border border-[#383A42] rounded px-2 py-1 text-xs text-[#E4E4E7]"
              >
                {enabledLanguages.map((lang) => (
                  <option key={lang.code} value={lang.code}>
                    {lang.name} ({lang.nativeName})
                  </option>
                ))}
              </select>
            </div>
          </div>

          <div className="p-4 grid grid-cols-2 gap-3">
            {languages.map((lang) => (
              <div
                key={lang.code}
                className={`bg-[#25272E] border rounded p-3 transition-colors ${
                  lang.enabled ? 'border-[#3B82F6]' : 'border-[#383A42]'
                }`}
              >
                <div className="flex items-center justify-between mb-2">
                  <div className="flex items-center gap-2">
                    <Globe size={14} className="text-[#3B82F6]" />
                    <div>
                      <div className="text-sm font-bold text-[#E4E4E7]">{lang.name}</div>
                      <div className="text-xs text-[#888]">{lang.nativeName}</div>
                    </div>
                  </div>
                  <label className="relative inline-block w-10 h-5">
                    <input
                      type="checkbox"
                      checked={lang.enabled}
                      onChange={() => handleToggleLanguage(lang.code)}
                      className="sr-only peer"
                    />
                    <span className="absolute inset-0 bg-[#383A42] rounded-full peer-checked:bg-[#3B82F6] transition-colors"></span>
                    <span className="absolute left-1 top-1 w-3 h-3 bg-white rounded-full peer-checked:translate-x-5 transition-transform"></span>
                  </label>
                </div>

                <div className="space-y-1">
                  <div className="flex justify-between text-xs">
                    <span className="text-[#888]">Completion</span>
                    <span className="text-[#E4E4E7] font-bold">{lang.completionPercent}%</span>
                  </div>
                  <div className="w-full bg-[#383A42] rounded-full h-1.5">
                    <div
                      className="bg-[#10b981] h-1.5 rounded-full transition-all"
                      style={{ width: `${lang.completionPercent}%` }}
                    ></div>
                  </div>
                  {lang.isRTL && (
                    <div className="pt-1 border-t border-[#383A42] mt-2">
                      <span className="text-[#F5C542] text-xs">✓ RTL Support Enabled</span>
                    </div>
                  )}
                </div>
              </div>
            ))}
          </div>
        </div>

        {/* Translation Management Table */}
        <div className="bg-[#1E2026] border border-[#383A42] rounded overflow-hidden">
          <div className="px-4 py-2 bg-[#25272E] border-b border-[#383A42] flex items-center justify-between">
            <h3 className="text-sm font-bold text-[#F5C542]">Translation Keys</h3>
            <button
              onClick={() => setShowPreview(!showPreview)}
              className="flex items-center gap-2 px-2 py-1 bg-[#1E2026] hover:bg-[#25272E] text-[#3B82F6] text-xs font-bold rounded transition-colors"
            >
              <Eye size={12} />
              {showPreview ? 'Hide Preview' : 'Show Preview'}
            </button>
          </div>

          {/* Language Selector for Table */}
          <div className="px-4 py-2 bg-[#1E2026] border-b border-[#383A42] flex items-center gap-2">
            <span className="text-xs text-[#888]">Viewing translations for:</span>
            <select
              value={selectedLanguage}
              onChange={(e) => setSelectedLanguage(e.target.value as LanguageCode)}
              className="bg-[#25272E] border border-[#383A42] rounded px-2 py-1 text-xs text-[#E4E4E7]"
            >
              {languages.map((lang) => (
                <option key={lang.code} value={lang.code}>
                  {lang.name} ({lang.nativeName})
                </option>
              ))}
            </select>
          </div>

          <div className="overflow-x-auto max-h-96">
            <table className="w-full text-xs">
              <thead className="bg-[#25272E] sticky top-0">
                <tr>
                  <th className="text-left px-3 py-2 text-[#888] font-bold border-r border-[#383A42]">Key</th>
                  <th className="text-left px-3 py-2 text-[#888] font-bold border-r border-[#383A42]">Category</th>
                  <th className="text-left px-3 py-2 text-[#888] font-bold border-r border-[#383A42]">English</th>
                  <th className="text-left px-3 py-2 text-[#888] font-bold">
                    {languages.find((l) => l.code === selectedLanguage)?.name}
                  </th>
                </tr>
              </thead>
              <tbody className="divide-y divide-[#383A42]">
                {translations.map((trans) => (
                  <tr key={trans.id} className="hover:bg-[#25272E]">
                    <td className="px-3 py-2 text-[#3B82F6] font-mono text-[10px] border-r border-[#383A42]">
                      {trans.key}
                    </td>
                    <td className="px-3 py-2 text-[#888] border-r border-[#383A42]">{trans.category}</td>
                    <td className="px-3 py-2 text-[#E4E4E7] border-r border-[#383A42]">{trans.englishValue}</td>
                    <td className="px-3 py-2 text-[#E4E4E7]">{trans.translations[selectedLanguage]}</td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        </div>

        {/* Locale Format Configuration */}
        <div className="bg-[#1E2026] border border-[#383A42] rounded">
          <div className="px-4 py-2 bg-[#25272E] border-b border-[#383A42]">
            <h3 className="text-sm font-bold text-[#F5C542]">Locale Format Configuration</h3>
          </div>

          <div className="p-4 space-y-3">
            <div className="flex items-center gap-2 mb-2">
              <span className="text-xs text-[#888]">Configure formats for:</span>
              <select
                value={selectedLanguage}
                onChange={(e) => setSelectedLanguage(e.target.value as LanguageCode)}
                className="bg-[#25272E] border border-[#383A42] rounded px-2 py-1 text-xs text-[#E4E4E7]"
              >
                {languages.map((lang) => (
                  <option key={lang.code} value={lang.code}>
                    {lang.name} ({lang.nativeName})
                  </option>
                ))}
              </select>
            </div>

            <div className="grid grid-cols-2 gap-3">
              <div>
                <label className="text-xs text-[#888] mb-1 block">Date Format</label>
                <input
                  type="text"
                  value={localeFormats[selectedLanguage].dateFormat}
                  onChange={(e) =>
                    setLocaleFormats((prev) => ({
                      ...prev,
                      [selectedLanguage]: { ...prev[selectedLanguage], dateFormat: e.target.value },
                    }))
                  }
                  className="w-full bg-[#25272E] border border-[#383A42] rounded px-2 py-1.5 text-xs text-[#E4E4E7]"
                  placeholder="MM/DD/YYYY"
                />
              </div>

              <div>
                <label className="text-xs text-[#888] mb-1 block">Time Format</label>
                <select
                  value={localeFormats[selectedLanguage].timeFormat}
                  onChange={(e) =>
                    setLocaleFormats((prev) => ({
                      ...prev,
                      [selectedLanguage]: { ...prev[selectedLanguage], timeFormat: e.target.value as '12h' | '24h' },
                    }))
                  }
                  className="w-full bg-[#25272E] border border-[#383A42] rounded px-2 py-1.5 text-xs text-[#E4E4E7]"
                >
                  <option value="12h">12-hour (AM/PM)</option>
                  <option value="24h">24-hour</option>
                </select>
              </div>

              <div>
                <label className="text-xs text-[#888] mb-1 block">Number Format</label>
                <select
                  value={localeFormats[selectedLanguage].numberFormat}
                  onChange={(e) =>
                    setLocaleFormats((prev) => ({
                      ...prev,
                      [selectedLanguage]: {
                        ...prev[selectedLanguage],
                        numberFormat: e.target.value as 'comma' | 'period' | 'space',
                      },
                    }))
                  }
                  className="w-full bg-[#25272E] border border-[#383A42] rounded px-2 py-1.5 text-xs text-[#E4E4E7]"
                >
                  <option value="comma">1,234.56 (Comma)</option>
                  <option value="period">1.234,56 (Period)</option>
                  <option value="space">1 234.56 (Space)</option>
                </select>
              </div>

              <div>
                <label className="text-xs text-[#888] mb-1 block">Currency Symbol</label>
                <input
                  type="text"
                  value={localeFormats[selectedLanguage].currencySymbol}
                  onChange={(e) =>
                    setLocaleFormats((prev) => ({
                      ...prev,
                      [selectedLanguage]: { ...prev[selectedLanguage], currencySymbol: e.target.value },
                    }))
                  }
                  className="w-full bg-[#25272E] border border-[#383A42] rounded px-2 py-1.5 text-xs text-[#E4E4E7]"
                  placeholder="$"
                />
              </div>

              <div>
                <label className="text-xs text-[#888] mb-1 block">Currency Position</label>
                <select
                  value={localeFormats[selectedLanguage].currencyPosition}
                  onChange={(e) =>
                    setLocaleFormats((prev) => ({
                      ...prev,
                      [selectedLanguage]: {
                        ...prev[selectedLanguage],
                        currencyPosition: e.target.value as 'before' | 'after',
                      },
                    }))
                  }
                  className="w-full bg-[#25272E] border border-[#383A42] rounded px-2 py-1.5 text-xs text-[#E4E4E7]"
                >
                  <option value="before">Before ($1,234)</option>
                  <option value="after">After (1,234$)</option>
                </select>
              </div>
            </div>
          </div>
        </div>

        {/* Preview Panel */}
        {showPreview && (
          <div className="bg-[#1E2026] border border-[#383A42] rounded">
            <div className="px-4 py-2 bg-[#25272E] border-b border-[#383A42]">
              <h3 className="text-sm font-bold text-[#F5C542]">UI Preview ({selectedLanguage.toUpperCase()})</h3>
            </div>

            <div
              className={`p-4 ${languages.find((l) => l.code === selectedLanguage)?.isRTL ? 'rtl' : 'ltr'}`}
              dir={languages.find((l) => l.code === selectedLanguage)?.isRTL ? 'rtl' : 'ltr'}
            >
              <div className="bg-[#25272E] border border-[#383A42] rounded p-3 space-y-2">
                <div className="text-sm font-bold text-[#F5C542]">
                  {translations.find((t) => t.key === 'dashboard.title')?.translations[selectedLanguage]}
                </div>
                <div className="flex gap-4 text-xs">
                  <div>
                    <span className="text-[#888]">
                      {translations.find((t) => t.key === 'account.balance')?.translations[selectedLanguage]}:
                    </span>
                    <span className="text-[#E4E4E7] font-bold ml-1">
                      {localeFormats[selectedLanguage].currencyPosition === 'before'
                        ? `${localeFormats[selectedLanguage].currencySymbol}10,250.00`
                        : `10,250.00${localeFormats[selectedLanguage].currencySymbol}`}
                    </span>
                  </div>
                </div>
                <div className="flex gap-2">
                  <button className="px-3 py-1 bg-[#10b981] text-white text-xs font-bold rounded">
                    {translations.find((t) => t.key === 'order.buy')?.translations[selectedLanguage]}
                  </button>
                  <button className="px-3 py-1 bg-[#ef4444] text-white text-xs font-bold rounded">
                    {translations.find((t) => t.key === 'order.sell')?.translations[selectedLanguage]}
                  </button>
                  <button className="px-3 py-1 bg-[#3B82F6] text-white text-xs font-bold rounded">
                    {translations.find((t) => t.key === 'common.save')?.translations[selectedLanguage]}
                  </button>
                </div>
              </div>
            </div>
          </div>
        )}

        {/* Save Button */}
        <div className="flex justify-end">
          <button className="px-4 py-2 bg-[#3B82F6] hover:bg-[#2563EB] text-white text-xs font-bold rounded transition-colors">
            Save Localization Settings
          </button>
        </div>
      </div>
    </div>
  );
}
