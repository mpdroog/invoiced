import js from '@eslint/js';
import tseslint from '@typescript-eslint/eslint-plugin';
import tsParser from '@typescript-eslint/parser';
import react from 'eslint-plugin-react';
import reactHooks from 'eslint-plugin-react-hooks';
import globals from 'globals';

export default [
  js.configs.recommended,
  {
    files: ['**/*.{ts,tsx}'],
    languageOptions: {
      parser: tsParser,
      parserOptions: {
        ecmaVersion: 'latest',
        sourceType: 'module',
        ecmaFeatures: {
          jsx: true
        },
        project: './tsconfig.json'
      },
      globals: {
        ...globals.browser,
        ...globals.es2020,
        handleErr: 'readonly',
        prettyErr: 'readonly',
        React: 'readonly'
      }
    },
    plugins: {
      '@typescript-eslint': tseslint,
      'react': react,
      'react-hooks': reactHooks
    },
    settings: {
      react: {
        version: 'detect'
      }
    },
    rules: {
      // TypeScript strict rules
      '@typescript-eslint/no-explicit-any': 'warn',
      '@typescript-eslint/explicit-function-return-type': 'warn',
      '@typescript-eslint/explicit-module-boundary-types': 'warn',
      '@typescript-eslint/no-inferrable-types': 'off',
      '@typescript-eslint/no-unused-vars': ['error', {
        argsIgnorePattern: '^_',
        varsIgnorePattern: '^_'
      }],
      '@typescript-eslint/strict-boolean-expressions': 'off',
      '@typescript-eslint/no-non-null-assertion': 'warn',
      '@typescript-eslint/prefer-nullish-coalescing': 'off',
      '@typescript-eslint/prefer-optional-chain': 'warn',

      // Promise rules - catch forgotten await and unhandled rejections
      '@typescript-eslint/no-floating-promises': 'error',
      '@typescript-eslint/no-misused-promises': 'error',

      // React rules
      'react/prop-types': 'off',
      'react/react-in-jsx-scope': 'off',
      'react/button-has-type': 'error', // Prevent accidental form submissions
      'react/jsx-key': 'error', // Catch missing keys in iterators
      'react/jsx-no-target-blank': 'error', // Security: require rel="noopener" with target="_blank"
      'react/no-direct-mutation-state': 'error', // Prevent this.state.x = y or this.state.arr.push()
      'react-hooks/rules-of-hooks': 'error',
      'react-hooks/exhaustive-deps': 'warn',

      // TypeScript advanced checks
      '@typescript-eslint/switch-exhaustiveness-check': 'error', // Ensure all enum/union cases handled
      '@typescript-eslint/await-thenable': 'error', // Only await on Promises
      '@typescript-eslint/no-unnecessary-condition': 'warn', // Catch always true/false conditions

      // General rules
      'no-unused-vars': 'off', // Use TypeScript's instead
      'no-undef': 'off', // TypeScript handles this
      'prefer-const': 'error',
      'no-var': 'error',
      'eqeqeq': ['error', 'always', { null: 'ignore' }]
    }
  },
  {
    ignores: ['node_modules/**', 'dist/**', '*.js']
  }
];
