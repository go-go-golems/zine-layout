import react from '@vitejs/plugin-react';
import { defineConfig } from 'vite';

export default defineConfig({
  plugins: [react()],
  build: {
    outDir: 'dist',
    sourcemap: true,
  },
  server: {
    port: 5173,
    proxy: {
      // Proxy API requests to the Go server when running `pnpm dev`
      '/api': {
        target: 'http://localhost:8088',
        changeOrigin: true,
      },
    },
  },
});
