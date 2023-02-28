import { defineConfig } from 'vite';
import solidPlugin from 'vite-plugin-solid';
import {resolve} from 'path'

export default defineConfig({
  plugins: [solidPlugin()],
  server: {
    port: 3000,
  },
  build: {
    target: 'esnext',
    rollupOptions: {
      input: {
        main: resolve(__dirname, 'console.html'),
        nested: resolve(__dirname, 'index.html')
      }
    }
  },
});
