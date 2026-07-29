import { defineConfig } from 'vite'
import react from '@vitejs/plugin-react'

export default defineConfig({
  plugins: [react()],
  base: '/app/music-dl/',
  build: {
    outDir: 'dist',
    emptyOutDir: true,
  },
  server: {
    port: 5173,
    proxy: {
      '/app/music-dl': {
        target: 'http://localhost:8080',
        changeOrigin: true,
      },
    },
  },
})
