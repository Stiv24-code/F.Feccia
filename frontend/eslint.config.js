import js from '@eslint/js';
import react from 'eslint-plugin-react';
import reactHooks from 'eslint-plugin-react-hooks';
import jsxA11y from 'eslint-plugin-jsx-a11y';
import importPlugin from 'eslint-plugin-import';
import tsParser from '@typescript-eslint/parser';
import tsPlugin from '@typescript-eslint/eslint-plugin';
import globals from 'globals';

const reactSettings = { react: { version: '19.0' } };

const sharedRules = {
  ...react.configs.recommended.rules,
  ...react.configs['jsx-runtime'].rules, // React 19: nuovo JSX transform, "React" non deve essere in scope
  ...reactHooks.configs.recommended.rules,
  ...jsxA11y.configs.recommended.rules,
  'react/prop-types': 'off', // niente PropTypes in questo codebase
  // catch(e) { toast.error(...) } senza usare `e` è un pattern ricorrente in
  // tutte le pagine CRUD del progetto, non un bug da segnalare per ognuna.
  'no-unused-vars': ['error', { caughtErrors: 'none' }],
};

export default [
  {
    // build/ è l'output di vite; src/api è generato da swagger-typescript-api
    // (rigenerato con `yarn generate:api`, non si edita a mano); src/components/ui
    // sono i primitivi shadcn — CLAUDE.md vieta di toccarli, quindi non ha
    // senso lintarli per regole che non potremmo comunque applicare.
    ignores: ['build/**', 'node_modules/**', 'src/api/**', 'src/components/ui/**'],
  },
  js.configs.recommended,
  {
    files: ['**/*.{js,jsx}'],
    languageOptions: {
      ecmaVersion: 2022,
      sourceType: 'module',
      globals: { ...globals.browser, ...globals.node },
      parserOptions: { ecmaFeatures: { jsx: true } },
    },
    plugins: { react, 'react-hooks': reactHooks, 'jsx-a11y': jsxA11y, import: importPlugin },
    settings: reactSettings,
    rules: sharedRules,
  },
  {
    files: ['**/*.{ts,tsx}'],
    languageOptions: {
      ecmaVersion: 2022,
      sourceType: 'module',
      parser: tsParser,
      globals: { ...globals.browser, ...globals.node },
      parserOptions: { ecmaFeatures: { jsx: true } },
    },
    plugins: {
      react, 'react-hooks': reactHooks, 'jsx-a11y': jsxA11y, import: importPlugin,
      '@typescript-eslint': tsPlugin,
    },
    settings: reactSettings,
    rules: {
      ...sharedRules,
      ...tsPlugin.configs.recommended.rules,
      'no-unused-vars': 'off',
      '@typescript-eslint/no-unused-vars': ['warn', { caughtErrors: 'none' }],
    },
  },
];
