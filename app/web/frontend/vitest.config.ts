import path from 'node:path'
import { defineConfig } from 'vitest/config'

// Unit tests target the pure TypeScript utilities under src/lib/utils.
// Node environment is sufficient (URLSearchParams, Blob, JSON are globals);
// no Svelte plugin or jsdom needed until component tests are added.
export default defineConfig({
  resolve: {
    alias: {
      $lib: path.resolve(__dirname, './src/lib'),
    },
  },
  test: {
    environment: 'node',
    include: ['src/**/*.{test,spec}.ts'],
    coverage: {
      provider: 'v8',
      include: ['src/lib/utils/**/*.ts'],
      // cert-icons.ts imports Svelte components and can't run in the Node env;
      // it holds no logic worth covering.
      exclude: ['src/lib/utils/cert-icons.ts', 'src/**/*.{test,spec}.ts'],
      reporter: ['text', 'html'],
    },
  },
})
