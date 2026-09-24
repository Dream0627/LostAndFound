# syntax=docker/dockerfile:1

# ============================================================================
# LAF 前端镜像（前后端分离部署用，多阶段构建）
# 阶段一用 Node 执行 vite build 产出静态文件；
# 阶段二用 Nginx 托管 dist，并把 /api、/uploads 反向代理到后端容器。
# ============================================================================

# ---- 构建阶段 ----
FROM node:20-alpine AS builder

WORKDIR /app

# 构建期注入后端 API 前缀：Nginx 同源反代，保持相对路径 /api/v1 即可，无需跨域。
ARG VITE_API_BASE_URL=/api/v1
ENV VITE_API_BASE_URL=$VITE_API_BASE_URL

# 先装依赖，利用层缓存。
# 关键：同时拷贝 package-lock.json，使安装可复现、更快（避免全量解析）。
# 并为 npm 增加重试与超时，降低网络抖动导致的失败。
COPY package.json package-lock.json ./
RUN npm config set registry https://registry.npmmirror.com \
    && npm config set fetch-retries 5 \
    && npm config set fetch-retry-mintimeout 20000 \
    && npm config set fetch-retry-maxtimeout 120000 \
    && npm config set fetch-timeout 120000 \
    && npm install --no-audit --no-fund

# 再拷贝源码构建。
COPY . ./
RUN npm run build

# ---- 运行阶段 ----
FROM nginx:1.27-alpine

# SPA 静态资源 + 反向代理配置。
COPY --from=builder /app/dist /usr/share/nginx/html
COPY nginx.conf /etc/nginx/conf.d/default.conf

EXPOSE 80

CMD ["nginx", "-g", "daemon off;"]
