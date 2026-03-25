import { defineConfig } from 'vite'
import vue from '@vitejs/plugin-vue'
import path from 'path'

export default defineConfig({
  plugins: [vue()],
  resolve: {
    alias: {
      '@': path.resolve(__dirname, './src')
    }
  },
  server: {
    port: 3000,
    proxy: {
      '/api/user': {
        target: 'http://localhost:8888',
        changeOrigin: true,
        rewrite: (path) => path.replace(/^\/api\/user/, '/user')
      },
      '/api/content': {
        target: 'http://localhost:8890',
        changeOrigin: true,
        rewrite: (path) => path.replace(/^\/api\/content/, '/content')
      },
      '/api/relation': {
        target: 'http://localhost:8889',
        changeOrigin: true,
        rewrite: (path) => path.replace(/^\/api\/relation/, '/relation')
      },
      '/api/comment': {
        target: 'http://localhost:8892',
        changeOrigin: true,
        rewrite: (path) => path.replace(/^\/api\/comment/, '/comment')
      }
    }
  }
})
