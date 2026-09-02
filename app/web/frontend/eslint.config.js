import js from '@eslint/js';
import svelte from 'eslint-plugin-svelte';
import globals from 'globals';
import tseslint from 'typescript-eslint';

/** @type {import('eslint').Linter.Config[]} */
export default [
  js.configs.recommended,
  ...tseslint.configs.recommended,
  ...svelte.configs['flat/recommended'],
  {
    languageOptions: {
      globals: {
        ...globals.browser,
        ...globals.node
      }
    }
  },
  {
    files: ['**/*.svelte', '**/*.svelte.ts', '**/*.svelte.js'],
    languageOptions: {
      parserOptions: {
        parser: tseslint.parser
      }
    }
  },
  {
    rules: {
      '@typescript-eslint/no-explicit-any': 'warn',
      '@typescript-eslint/no-unused-vars': ['warn', { argsIgnorePattern: '^_' }],
      'svelte/prefer-svelte-reactivity': 'warn'
    }
  },
  {
    // Bare reads in $effect (e.g. `foo; bar;`) are the Svelte reactive-dependency
    // idiom; the rule misreads them as dead expressions. Keep it on for plain .ts.
    files: ['**/*.svelte'],
    rules: {
      '@typescript-eslint/no-unused-expressions': 'off'
    }
  },
  {
    ignores: ['dist/', 'coverage/', 'node_modules/', '**/*.d.ts']
  }
];
