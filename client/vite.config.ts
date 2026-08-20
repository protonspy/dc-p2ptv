import { defineConfig } from 'vite';

export default defineConfig({
  build: {
    lib: {
      entry: 'src/index.ts',
      name: 'dcP2ptv',
      fileName: 'dc-p2ptv',
    },
    target: 'es2022',
  },
  test: {
    globals: true,
    environment: 'jsdom',
    include: ['src/**/*.test.ts'],
    coverage: {
      provider: 'v8',
      reporter: ['text-summary', 'lcov'],
      include: ['src/**/*.ts'],
      exclude: ['src/**/*.test.ts'],
      // The gate the pull request has to clear. Written here rather than in the
      // workflow so that a local run fails for the same reason CI does.
      thresholds: {
        lines: 86,
        functions: 86,
        branches: 86,
        statements: 86,
      },
    },
  },
});
