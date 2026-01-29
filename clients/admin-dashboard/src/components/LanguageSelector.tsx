import { useState, useRef, useEffect } from 'react';
import { useI18n } from '../i18n/hooks';
import { ChevronDown, Globe } from 'lucide-react';
import type { SupportedLanguage } from '../i18n/config';

export const LanguageSelector = () => {
  const { currentLanguage, languageInfo, changeLanguage, languages } = useI18n();
  const [isOpen, setIsOpen] = useState(false);
  const dropdownRef = useRef<HTMLDivElement>(null);

  useEffect(() => {
    const handleClickOutside = (event: MouseEvent) => {
      if (dropdownRef.current && !dropdownRef.current.contains(event.target as Node)) {
        setIsOpen(false);
      }
    };

    document.addEventListener('mousedown', handleClickOutside);
    return () => document.removeEventListener('mousedown', handleClickOutside);
  }, []);

  const handleLanguageChange = (lng: SupportedLanguage) => {
    changeLanguage(lng);
    setIsOpen(false);
  };

  return (
    <div className="relative" ref={dropdownRef}>
      <button
        onClick={() => setIsOpen(!isOpen)}
        className="flex items-center gap-2 px-3 py-2 rounded-lg bg-white dark:bg-gray-800
                   border border-gray-200 dark:border-gray-700 hover:bg-gray-50
                   dark:hover:bg-gray-700 transition-colors"
        aria-label="Select language"
        aria-expanded={isOpen}
      >
        <Globe className="w-4 h-4" />
        <span className="text-sm font-medium">
          {languageInfo?.flag} {languageInfo?.nativeName}
        </span>
        <ChevronDown
          className={`w-4 h-4 transition-transform ${isOpen ? 'rotate-180' : ''}`}
        />
      </button>

      {isOpen && (
        <div
          className="absolute right-0 mt-2 w-64 bg-white dark:bg-gray-800 rounded-lg
                     shadow-lg border border-gray-200 dark:border-gray-700 z-50
                     max-h-96 overflow-y-auto"
        >
          <div className="p-2">
            {Object.entries(languages).map(([code, lang]) => (
              <button
                key={code}
                onClick={() => handleLanguageChange(code as SupportedLanguage)}
                className={`w-full flex items-center gap-3 px-3 py-2 rounded-md text-left
                           transition-colors ${
                             currentLanguage === code
                               ? 'bg-blue-50 dark:bg-blue-900/20 text-blue-600 dark:text-blue-400'
                               : 'hover:bg-gray-50 dark:hover:bg-gray-700'
                           }`}
              >
                <span className="text-xl">{lang.flag}</span>
                <div className="flex-1">
                  <div className="font-medium">{lang.nativeName}</div>
                  <div className="text-xs text-gray-500 dark:text-gray-400">
                    {lang.name}
                  </div>
                </div>
                {currentLanguage === code && (
                  <div className="w-2 h-2 rounded-full bg-blue-600" />
                )}
              </button>
            ))}
          </div>
        </div>
      )}
    </div>
  );
};
