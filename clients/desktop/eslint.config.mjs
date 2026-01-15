import js from '@eslint/js'
import globals from 'globals'
import reactHooks from 'eslint-plugin-react-hooks'
import reactRefresh from 'eslint-plugin-react-refresh'
import tseslint from 'typescript-eslint'

export default tseslint.config(
  { ignores: ['dist'] },
  js.configs.recommended,
  ...tseslint.configs.recommended,
  ...tseslint.configs.stylistic,
  {
    files: ['**/*.{ts,tsx}'],
    languageOptions: {
      ecmaVersion: 2020,
      globals: globals.browser,
    },
    plugins: {
      'react-hooks': reactHooks,
      'react-refresh': reactRefresh,
    },
    rules: {
      // React hooks rules - all downgraded to warnings for gradual improvement
      ...Object.fromEntries(
        Object.entries(reactHooks.configs.recommended.rules).map(([key, value]) => [
          key,
          typeof value === 'string' && value === 'error' ? 'warn' : (Array.isArray(value) && value[0] === 'error' ? ['warn', ...value.slice(1)] : value)
        ])
      ),
      'react-refresh/only-export-components': [
        'warn',
        { allowConstantExport: true },
      ],
      // Allow unused vars starting with _
      '@typescript-eslint/no-unused-vars': ['error', {
        argsIgnorePattern: '^_',
        varsIgnorePattern: '^_'
      }],
      // Warn on console.log (should use proper logging)
      'no-console': ['warn', { allow: ['warn', 'error'] }],
      // Enforce type imports
      '@typescript-eslint/consistent-type-imports': 'error',
      // Prefer type over interface per CLAUDE.md
      '@typescript-eslint/consistent-type-definitions': ['warn', 'type'],
      // Allow any types for now - will fix gradually
      '@typescript-eslint/no-explicit-any': 'warn',
      // Allow empty functions in tests
      '@typescript-eslint/no-empty-function': 'warn',
      // Allow type inference simplifications
      '@typescript-eslint/no-inferrable-types': 'warn',
      // Allow generic constructor patterns
      '@typescript-eslint/consistent-generic-constructors': 'warn',
      // Allow empty blocks and unused vars in existing code
      'no-empty': 'warn',
      // Allow index signatures
      '@typescript-eslint/consistent-indexed-object-style': 'warn',
      // Allow Array<T> syntax
      '@typescript-eslint/array-type': 'warn',
      // Allow variable use before declaration in callbacks (false positive)
      'no-use-before-define': 'off',
      '@typescript-eslint/no-use-before-define': 'off',
    },
  },
)
