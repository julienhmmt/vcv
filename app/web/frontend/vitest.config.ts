import path from 'node:path'
import { defineConfig } from 'vitest/config'
import { svelte } from '@sveltejs/vite-plugin-svelte'
import { svelteTesting } from '@testing-library/svelte/vite'

// Pure TypeScript utilities run in the default Node environment. Tests that
// need the DOM (component renders, the export download glue) opt in per file
// with `// @vitest-environment jsdom`.
export default defineConfig({
  plugins: [svelte(), svelteTesting()],
  resolve: {
    alias: {
      $lib: path.resolve(__dirname, './src/lib'),
    },
  },
  test: {
    globals: true,
    environment: 'node',
    setupFiles: ['./vitest.setup.ts'],
    include: ['src/**/*.{test,spec}.ts'],
    coverage: {
      provider: 'v8',
      include: ['src/lib/utils/**/*.ts'],
      // cert-icons.ts is a thin Svelte-component lookup with no logic to cover.
      exclude: ['src/lib/utils/cert-icons.ts', 'src/**/*.{test,spec}.ts'],
      reporter: ['text', 'html'],
    },
  },
})
