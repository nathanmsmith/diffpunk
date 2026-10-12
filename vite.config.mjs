import { defineConfig } from 'vite';
import react from '@vitejs/plugin-react';
import tailwindcss from '@tailwindcss/vite';
import { fileURLToPath } from 'node:url';

// Go probes this endpoint before forwarding frontend requests to Vite.
const devBase = '/__diff_punk_dev/';

export default defineConfig(({ command }) => ({
  root: fileURLToPath(new URL('./frontend', import.meta.url)),
  base: command === 'serve' ? devBase : './',
  plugins: [
    react(),
    tailwindcss(),
    {
      name: 'diff-punk-dev-server',
      configureServer(server) {
        server.middlewares.use((req, res, next) => {
          if (req.url !== `${devBase}status`) return next();
          res.setHeader('Content-Type', 'text/plain');
          res.setHeader('Cache-Control', 'no-store');
          res.end('diff-punk-vite-v1');
        });
      },
    },
  ],
  server: {
    host: '127.0.0.1',
    port: 5173,
    strictPort: true,
    cors: false,
  },
  build: {
    outDir: '../cli/assets/dist',
    emptyOutDir: true,
    target: 'es2022',
  },
}));
