import { defineConfig } from 'vite'
import react from '@vitejs/plugin-react'
import path from 'path'

export default defineConfig({
  plugins: [react()],
  root: '.',
  base: '/static/assets/',
  build: {
    outDir: '../static/assets',
    emptyOutDir: true,
    rollupOptions: {
      input: {
        app: path.resolve(__dirname, 'index.html')
      },
      output: {
        entryFileNames: 'app.js',
        chunkFileNames: 'vendor.js',
        assetFileNames: '[name][extname]'
      }
    }
  },
  resolve: {
    alias: {
      '@': path.resolve(__dirname, 'app')
    }
  },
  server: {
    proxy: {
      '/api': 'http://localhost:9999'
    }
  }
})
