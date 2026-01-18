/**
 * Pseudo-localization for testing i18n implementation
 *
 * Transforms text to help identify:
 * - Hardcoded strings (won't be transformed)
 * - UI layout issues with longer text
 * - Missing translations
 * - Text truncation problems
 */

const PSEUDO_CHARS: Record<string, string> = {
  a: 'á',
  b: 'ƀ',
  c: 'ç',
  d: 'ð',
  e: 'é',
  f: 'ƒ',
  g: 'ĝ',
  h: 'ĥ',
  i: 'î',
  j: 'ĵ',
  k: 'ķ',
  l: 'ļ',
  m: 'ɱ',
  n: 'ñ',
  o: 'ö',
  p: 'þ',
  q: 'ǫ',
  r: 'ŕ',
  s: 'š',
  t: 'ţ',
  u: 'û',
  v: 'ṽ',
  w: 'ŵ',
  x: 'ẋ',
  y: 'ý',
  z: 'ž',
  A: 'Á',
  B: 'Ɓ',
  C: 'Ç',
  D: 'Ð',
  E: 'É',
  F: 'Ƒ',
  G: 'Ĝ',
  H: 'Ĥ',
  I: 'Î',
  J: 'Ĵ',
  K: 'Ķ',
  L: 'Ļ',
  M: 'Ṁ',
  N: 'Ñ',
  O: 'Ö',
  P: 'Þ',
  Q: 'Ǫ',
  R: 'Ŕ',
  S: 'Š',
  T: 'Ţ',
  U: 'Û',
  V: 'Ṽ',
  W: 'Ŵ',
  X: 'Ẋ',
  Y: 'Ý',
  Z: 'Ž',
};

/**
 * Transform text to pseudo-localized version
 */
export const pseudoLocalize = (text: string, options?: {
  expand?: boolean;
  brackets?: boolean;
  accents?: boolean;
}): string => {
  const {
    expand = true,      // Expand text by ~30% to test layout
    brackets = true,    // Add brackets to identify pseudo text
    accents = true,     // Add accents to characters
  } = options || {};

  let result = text;

  // Add accents
  if (accents) {
    result = result
      .split('')
      .map((char) => PSEUDO_CHARS[char] || char)
      .join('');
  }

  // Expand text to simulate longer translations
  if (expand) {
    // Add extra characters proportional to length
    const expansion = Math.ceil(result.length * 0.3);
    result = result + ' ' + '·'.repeat(expansion);
  }

  // Add brackets to identify pseudo-localized text
  if (brackets) {
    result = `[${result}]`;
  }

  return result;
};

/**
 * Create pseudo-localized translations from English
 */
export const createPseudoTranslations = (
  translations: Record<string, any>
): Record<string, any> => {
  const pseudo: Record<string, any> = {};

  const transform = (obj: any): any => {
    if (typeof obj === 'string') {
      // Skip interpolation variables
      const hasVariables = /\{\{.*?\}\}/.test(obj);
      if (hasVariables) {
        // Only pseudo-localize the text parts, preserve variables
        return obj.replace(/([^{]+)(\{\{.*?\}\})?/g, (match, text, variable) => {
          const pseudoText = pseudoLocalize(text.trim());
          return variable ? `${pseudoText}${variable}` : pseudoText;
        });
      }
      return pseudoLocalize(obj);
    }

    if (Array.isArray(obj)) {
      return obj.map(transform);
    }

    if (typeof obj === 'object' && obj !== null) {
      const result: Record<string, any> = {};
      for (const [key, value] of Object.entries(obj)) {
        result[key] = transform(value);
      }
      return result;
    }

    return obj;
  };

  for (const [key, value] of Object.entries(translations)) {
    pseudo[key] = transform(value);
  }

  return pseudo;
};

/**
 * Validate that all text is using i18n (no hardcoded strings)
 */
export const detectHardcodedStrings = (
  node: Element,
  options?: {
    ignoredClasses?: string[];
    ignoredIds?: string[];
    ignoredTags?: string[];
  }
): string[] => {
  const {
    ignoredClasses = [],
    ignoredIds = [],
    ignoredTags = ['script', 'style', 'svg', 'path'],
  } = options || {};

  const hardcodedStrings: string[] = [];

  const checkNode = (element: Element) => {
    // Skip ignored elements
    if (
      ignoredTags.includes(element.tagName.toLowerCase()) ||
      ignoredClasses.some((cls) => element.classList.contains(cls)) ||
      ignoredIds.some((id) => element.id === id)
    ) {
      return;
    }

    // Check text nodes
    Array.from(element.childNodes).forEach((child) => {
      if (child.nodeType === Node.TEXT_NODE) {
        const text = child.textContent?.trim();
        if (text && text.length > 0) {
          // Check if text is pseudo-localized (should start with [)
          if (!text.startsWith('[')) {
            hardcodedStrings.push(text);
          }
        }
      }
    });

    // Recurse through children
    Array.from(element.children).forEach(checkNode);
  };

  checkNode(node);
  return hardcodedStrings;
};

/**
 * Test RTL layout by mirroring the UI
 */
export const enableRTLTesting = () => {
  document.documentElement.dir = 'rtl';
  document.documentElement.lang = 'ar-SA';
};

/**
 * Test character encoding
 */
export const testCharacterEncoding = (): {
  supportsUnicode: boolean;
  supportedScripts: string[];
} => {
  const testStrings = {
    arabic: 'العربية',
    chinese: '简体中文',
    japanese: '日本語',
    cyrillic: 'Русский',
    emoji: '🚀🎉',
  };

  const supportedScripts: string[] = [];
  let supportsUnicode = true;

  // Create temporary element to test rendering
  const testElement = document.createElement('div');
  testElement.style.position = 'absolute';
  testElement.style.visibility = 'hidden';
  document.body.appendChild(testElement);

  for (const [script, text] of Object.entries(testStrings)) {
    testElement.textContent = text;
    const rendered = testElement.textContent === text;

    if (rendered) {
      supportedScripts.push(script);
    } else {
      supportsUnicode = false;
    }
  }

  document.body.removeChild(testElement);

  return {
    supportsUnicode,
    supportedScripts,
  };
};
