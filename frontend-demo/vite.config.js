import { fileURLToPath, URL } from "node:url";
import { defineConfig } from "vite";
import vue from "@vitejs/plugin-vue";

// Vite 配置：Vue 单文件组件 + 路径别名 + 开发代理。
// 开发时前端跑在 5173，通过 proxy 把 /api 转发到后端 8080，避免跨域。
export default defineConfig({
  plugins: [vue()],
  resolve: {
    alias: {
      "@": fileURLToPath(new URL("./src", import.meta.url)),
    },
  },
  server: {
    port: 5173,
    proxy: {
      "/api": {
        target: "http://localhost:8080",
        changeOrigin: true,
      },
      // 帖子图片经后端 /uploads 静态提供，同样代理过去，避免图片 404。
      "/uploads": {
        target: "http://localhost:8080",
        changeOrigin: true,
      },
    },
  },
});
