#!/usr/bin/env bash
# ============================================================================
# LAF 前后端分离 · 单端口(8080) 一键部署脚本
# 用法：把项目上传到服务器后，在项目根目录执行：  bash deploy.sh
# 作用：内存/swap 自检 -> Docker 自检(可自动安装) -> 注入 public_base_url ->
#       生成 JWT secret -> 构建启动 -> 健康检查
# 非交互：ASSUME_YES=1 bash deploy.sh
# 注意：需在 Linux 运行，使用 LF 换行（从 Windows 拷贝时勿被转成 CRLF）。
# ============================================================================
set -euo pipefail
cd "$(dirname "$0")"

HTTP_PORT="${HTTP_PORT:-8080}"
COMPOSE_FILE="docker-compose.yml"
CFG="config/config.docker.yaml"
ASSUME_YES="${ASSUME_YES:-0}"

log()  { printf "\033[1;32m[deploy]\033[0m %s\n" "$*"; }
warn() { printf "\033[1;33m[deploy]\033[0m %s\n" "$*"; }
err()  { printf "\033[1;31m[deploy]\033[0m %s\n" "$*" >&2; }
ask()  {
  [ "$ASSUME_YES" = "1" ] && return 0
  read -r -p "$1 [y/N] " a; case "$a" in y|Y|yes|YES) return 0;; *) return 1;; esac
}

# ---- 0. 内存 / swap 自检（2G 机器构建 + MySQL 易 OOM）---------------------
MEM_MB=$(awk '/MemTotal/{printf "%d",$2/1024}' /proc/meminfo 2>/dev/null || echo 0)
SWAP_MB=$(awk '/SwapTotal/{printf "%d",$2/1024}' /proc/meminfo 2>/dev/null || echo 0)
if [ "$MEM_MB" -gt 0 ] && [ "$MEM_MB" -lt 2560 ] && [ "$SWAP_MB" -lt 512 ]; then
  warn "检测到内存 ${MEM_MB}MB 且几乎无 swap，构建镜像/启动 MySQL 可能 OOM。"
  if ask "是否创建 2GB swap 交换文件？"; then
    if [ ! -f /swapfile ]; then
      fallocate -l 2G /swapfile 2>/dev/null || dd if=/dev/zero of=/swapfile bs=1M count=2048
      chmod 600 /swapfile; mkswap /swapfile >/dev/null; swapon /swapfile
      grep -q '/swapfile' /etc/fstab || echo '/swapfile none swap sw 0 0' >> /etc/fstab
      log "已启用 2GB swap"
    else
      warn "/swapfile 已存在，跳过"
    fi
  fi
fi

# ---- 1. Docker 自检 / 自动安装 --------------------------------------------
if ! command -v docker >/dev/null 2>&1; then
  warn "未检测到 docker。"
  if ask "是否现在自动安装 Docker（官方脚本 get.docker.com）？"; then
    curl -fsSL https://get.docker.com | sh
    systemctl enable --now docker 2>/dev/null || service docker start 2>/dev/null || true
    log "Docker 安装完成：$(docker --version 2>/dev/null || echo 未知)"
  else
    err "请先安装 Docker：https://docs.docker.com/engine/install/"; exit 1
  fi
fi
if docker compose version >/dev/null 2>&1; then
  DC="docker compose"
elif command -v docker-compose >/dev/null 2>&1; then
  DC="docker-compose"
else
  err "未检测到 docker compose（需 v2 插件）"; exit 1
fi
[ -f "$CFG" ] || { err "缺少 $CFG"; exit 1; }
[ -f "$COMPOSE_FILE" ] || { err "缺少 $COMPOSE_FILE"; exit 1; }

# ---- 2. 推断对外地址 ------------------------------------------------------
PUBLIC_BASE_URL="${PUBLIC_BASE_URL:-}"
if [ -z "$PUBLIC_BASE_URL" ]; then
  IP="$(curl -fsS --max-time 5 https://api.ipify.org 2>/dev/null || true)"
  [ -z "$IP" ] && IP="127.0.0.1"
  PUBLIC_BASE_URL="http://${IP}:${HTTP_PORT}"
  log "未指定 PUBLIC_BASE_URL，自动推断为 ${PUBLIC_BASE_URL}（可 export PUBLIC_BASE_URL=... 覆盖）"
fi
case "$PUBLIC_BASE_URL" in
  *127.0.0.1*|*localhost*) warn "当前对外地址是本地回环，若服务器未绑定公网IP，外部将无法访问。" ;;
esac

# ---- 3. 写入 public_base_url ---------------------------------------------
if grep -q "public_base_url" "$CFG"; then
  sed -i.bak -E "s#(public_base_url:[[:space:]]*).*#\1\"${PUBLIC_BASE_URL}\"#" "$CFG"
  log "已写入 server.public_base_url = ${PUBLIC_BASE_URL}"
fi

# ---- 4. 生成 JWT secret（仍为占位符时）-----------------------------------
if grep -q "CHANGE_ME_WITH_RANDOM_SECRET" "$CFG"; then
  SECRET="$(head -c 32 /dev/urandom | od -An -tx1 | tr -d ' \n')"
  sed -i.bak -E "s#(secret:[[:space:]]*).*#\1\"${SECRET}\"#" "$CFG"
  log "已生成随机 jwt.secret"
fi
rm -f "$CFG".bak

# ---- 5. 构建并启动 --------------------------------------------------------
log "开始构建并启动容器（首次较慢，请耐心等待）..."
$DC -f "$COMPOSE_FILE" up -d --build

# ---- 6. 健康检查 ----------------------------------------------------------
log "等待前端站点就绪..."
OK=0
for i in $(seq 1 60); do
  code="$(curl -s -o /dev/null -w '%{http_code}' --max-time 3 "http://127.0.0.1:${HTTP_PORT}/" || true)"
  if [ "$code" = "200" ] || [ "$code" = "304" ]; then
    log "前端已就绪：${PUBLIC_BASE_URL}/  （HTTP ${code}）"; OK=1; break
  fi
  sleep 2
done
[ "$OK" = "1" ] || err "前端未在预期时间内就绪，请查看： $DC logs -f frontend"

log "容器状态："
$DC -f "$COMPOSE_FILE" ps
log "完成。对外访问： ${PUBLIC_BASE_URL}"