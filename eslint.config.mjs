import js from '@eslint/js';
import { defineConfig, globalIgnores } from 'eslint/config';
import prettier from 'eslint-config-prettier/flat';
import globals from 'globals';
import tseslint from 'typescript-eslint';
import noArbitraryTailwindClasses from './scripts/eslint/no-arbitrary-tailwind-classes.mjs';

export default defineConfig([
  globalIgnores(['bin/**', 'release/**', 'cli/assets/dist/**']),
  {
    files: ['**/*.{js,mjs,cjs,ts,tsx}'],
    extends: [js.configs.recommended],
  },
  {
    files: ['frontend/src/**/*.{ts,tsx}'],
    extends: [tseslint.configs.recommended],
    languageOptions: { globals: globals.browser },
    plugins: {
      local: {
        rules: { 'no-arbitrary-tailwind-classes': noArbitraryTailwindClasses },
      },
    },
    rules: { 'local/no-arbitrary-tailwind-classes': 'error' },
  },
  {
    files: ['**/*.{js,mjs,cjs}'],
    languageOptions: { globals: globals.node },
  },
  {
    // Playwright callbacks execute in the browser.
    files: ['scripts/browser-test.mjs', 'scripts/dev-test.mjs'],
    languageOptions: { globals: globals.browser },
  },
  prettier,
]);
