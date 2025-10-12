import react from '@vitejs/plugin-react';
import { defineConfig, loadEnv } from 'vite';

export default defineConfig(({ mode }) => {
  const env = loadEnv(mode, process.cwd(), '');
  const proxyTarget = env.VITE_API_PROXY ?? 'http://localhost:8090';

  return {
    plugins: [react()],
    build: {
      outDir: 'dist',
      sourcemap: true,
    },
    server: {
      port: 5173,
      proxy: {
        '/api': {
          target: proxyTarget,
          changeOrigin: true,
        },
        '/projects': {
          target: proxyTarget,
          changeOrigin: true,
          bypass: (req) => {
            const accept = req.headers.accept ?? '';
            if (accept.includes('text/html')) {
              return '/index.html';
            }
            return undefined;
          },
        },
        '/uploads': {
          target: proxyTarget,
          changeOrigin: true,
        },
      },
    },
  };
});
