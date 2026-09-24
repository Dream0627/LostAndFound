#!/usr/bin/env bash
# ============================================================================
# 【在目标服务器(无外网)执行】
# 前置：已把 build-and-save.sh 产出的 dist-offline/laf-images.tar 随项目一起拷入。
# 作用：导入镜像 -> 启动容器 -> 健康检查（全程不需要外网）。
#      优先用 docker compose（v2 插件 或 v1 二进制 docker-compose）；
#      若两者都没有，自动降级为 docker run 手动编排，保证无 compose 环境也能启动。
# 用法：在项目根目录执行  bash offline/load-and-run.sh
# ============================================================================
set -euo pipefail
cd "$(dirname "$0")/.."

HTTP_PORT="${HTTP_PORT:-8080}"
TAR="dist-offline/laf-images.tar"
CFG="config/config.docker.yaml"

log(){ printf "\033[1;32m[offline]\033[0m %s\n" "$*"; }
err(){ printf "\033[1;31m[offline]\033[0m %s\n" "$*" >&2; }

command -v docker >/dev/null 2>&1 || { err "服务器未安装 Docker（离线环境需先离线安装 Docker）。"; exit 1; }
[ -f "$TAR" ] || { err "缺少镜像包 $TAR"; exit 1; }
[ -f "$CFG" ] || { err "缺少 $CFG"; exit 1; }

# 写入对外地址(内网) + 随机 JWT
if grep -q "public_base_url" "$CFG"; then
  sed -i.bak -E "s#(public_base_url:[[:space:]]*).*#\1\"http://$(hostname -I | awk '{print $1}'):${HTTP_PORT}\"#" "$CFG"
  rm -f "$CFG".bak
fi
if grep -q "CHANGE_ME_WITH_RANDOM_SECRET" "$CFG"; then
  SECRET="$(head -c 32 /dev/urandom | od -An -tx1 | tr -d ' \n')"
  sed -i.bak -E "s#(secret:[[:space:]]*).*#\1\"${SECRET}\"#" "$CFG"; rm -f "$CFG".bak
fi

log "导入镜像..."
docker load -i "$TAR"

# ---- 选择编排方式：docker compose(v2) / docker-compose(v1) / 手动 docker run ----
if docker compose version >/dev/null 2>&1; then
  DC="docker compose"
elif command -v docker-compose >/dev/null 2>&1; then
  DC="docker-compose"
else
  DC=""
fi

if [ -n "$DC" ]; then
  log "使用 [$DC] 启动容器（不构建）..."
  $DC up -d --no-build
else
  err "未检测到 docker compose / docker-compose，改用 docker run 手动编排（离线可用）。"
  NET="laf-net"
  docker network inspect "$NET" >/dev/null 2>&1 || docker network create "$NET" >/dev/null
  ROOT="$(pwd)"

  # 若同名容器已存在，先清理以便重跑
  docker rm -f laf-frontend laf-backend laf-mysql >/dev/null 2>&1 || true

  log "启动 MySQL..."
  docker run -d --name laf-mysql --restart unless-stopped \
    --network "$NET" --network-alias mysql \
    -e MYSQL_ROOT_PASSWORD=root123456 -e MYSQL_DATABASE=laf_db \
    -e MYSQL_USER=laf -e MYSQL_PASSWORD=laf123456 -e TZ=Asia/Shanghai \
    -v laf-mysql-data:/var/lib/mysql \
    -v "${ROOT}/migrations/tables.sql:/docker-entrypoint-initdb.d/01_tables.sql:ro" \
    mysql:8.0 --character-set-server=utf8mb4 --collation-server=utf8mb4_unicode_ci

  log "等待 MySQL 就绪（首次初始化较慢）..."
  for i in $(seq 1 60); do
    docker exec laf-mysql mysqladmin ping -h127.0.0.1 -uroot -proot123456 --silent >/dev/null 2>&1 && break
    sleep 3
  done

  log "启动 backend..."
  docker run -d --name laf-backend --restart unless-stopped \
    --network "$NET" --network-alias backend \
    -e TZ=Asia/Shanghai \
    -v "${ROOT}/config/config.docker.yaml:/app/config/config.yaml:ro" \
    -v laf-uploads:/app/uploads \
    laf-backend:latest

  log "启动 frontend..."
  docker run -d --name laf-frontend --restart unless-stopped \
    --network "$NET" \
    -p "${HTTP_PORT}:80" \
    laf-frontend:latest
fi

log "等待前端就绪..."
for i in $(seq 1 60); do
  code="$(curl -s -o /dev/null -w '%{http_code}' --max-time 3 "http://127.0.0.1:${HTTP_PORT}/" || true)"
  { [ "$code" = "200" ] || [ "$code" = "304" ]; } && { log "已就绪 HTTP ${code}"; break; }
  sleep 2
done
docker ps --filter "name=laf-" --format 'table {{.Names}}\t{{.Status}}\t{{.Ports}}'
IP="$(hostname -I | awk '{print $1}')"
log "完成。内网访问： http://${IP}:${HTTP_PORT}"
