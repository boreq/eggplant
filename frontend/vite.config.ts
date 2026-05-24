import { defineConfig } from 'vite';
import vue from '@vitejs/plugin-vue2';
import checker from 'vite-plugin-checker';
import path from 'path';

export default defineConfig({
  plugins: [
    vue(),
    checker({
      vueTsc: true,
      eslint: {
        lintCommand: 'eslint .',
        useFlatConfig: true,
      },
    }),
  ],
  envPrefix: ['VUE_APP_', 'VITE_'],
  resolve: {
    alias: {
      '@': path.resolve(__dirname, 'src'),
    },
  },
  css: {
    preprocessorOptions: {
      scss: {
        api: 'modern',
        additionalData: `@use "@/scss/variables" as *;`,
      },
    },
  },
  build: {
    sourcemap: false,
  },
});
