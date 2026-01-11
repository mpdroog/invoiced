import { defineConfig } from 'vite'
import path from 'path'

export default defineConfig({
  root: '.',
  base: '/static/assets/',
  esbuild: {
    jsxFactory: 'React.createElement',
    jsxFragment: 'React.Fragment'
  },
  build: {
    outDir: '../static/assets',
    emptyOutDir: false,
    minify: false,  // Disable minification for easier debugging
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
