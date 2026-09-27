#!/usr/bin/env bash
# ============================================================================
# LAF 前后端分离 · 一键部署脚本
# 端口规划：8080=后端 API（直接对外），9090=临时前端 frontend-demo。
# 用法：把项目上传到服务器后，在项目根目录执行：  bash deploy.sh
# 作用：内存/swap 自检 -> Docker 自检(可自动安装) ->
#       生成 JWT secret -> 构建启动 -> 健康检查
# 非交互：ASSUME_YES=1 bash deploy.sh
# 注意：需在 Linux 运行，使用 LF 换行（从 Windows 拷贝时勿被转成 CRLF）。
# ============================================================================
set -euo pipefail
cd "$(dirname "$0")"

BACKEND_PORT="${BACKEND_PORT:-8080}"
FRONTEND_PORT="${FRONTEND_PORT:-9090}"
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
# 注意：低内存机器无 swap 时 Go 编译/链接会吃光内存导致整机换页假死、SSH 无响应
# （1.7G 机器实测：包编译限并行后仍在链接阶段 wa 91%），故这里【自动创建】4G swap，
# 不再依赖可能被随手跳过的交互提问。
# 确实要跳过（例如自有 swap 方案）： SKIP_SWAP=1 bash deploy.sh
MEM_MB=$(awk '/MemTotal/{printf "%d",$2/1024}' /proc/meminfo 2>/dev/null || echo 0)
SWAP_MB=$(awk '/SwapTotal/{printf "%d",$2/1024}' /proc/meminfo 2>/dev/null || echo 0)
if [ "$MEM_MB" -gt 0 ] && [ "$MEM_MB" -lt 2560 ] && [ "$SWAP_MB" -lt 2048 ]; then
  if [ "${SKIP_SWAP:-0}" = "1" ]; then
    warn "内存 ${MEM_MB}MB、swap ${SWAP_MB}MB，且已指定 SKIP_SWAP=1：跳过 swap 创建，构建可能 OOM 假死。"
  elif [ ! -f /swapfile ]; then
    warn "检测到内存 ${MEM_MB}MB 且 swap 仅 ${SWAP_MB}MB，构建前自动创建 4GB swap ..."
    fallocate -l 4G /swapfile 2>/dev/null || dd if=/dev/zero of=/swapfile bs=1M count=4096
    chmod 600 /swapfile; mkswap /swapfile >/dev/null; swapon /swapfile
    grep -q '/swapfile' /etc/fstab || echo '/swapfile none swap sw 0 0' >> /etc/fstab
    log "已启用 4GB swap（已写入 /etc/fstab，重启后依然生效）"
  else
    warn "/swapfile 已存在但当前 swap 不足，尝试重新启用"
    swapon /swapfile 2>/dev/null || true
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

# .env 保存 MySQL 口令（compose 用 ${MYSQL_*} 注入），不提交 git。
ENV_FILE=".env"
if [ ! -f "$ENV_FILE" ]; then
  if [ -f ".env.example" ]; then
    cp .env.example "$ENV_FILE"
    warn "未找到 .env，已从 .env.example 生成一份。"
    warn "请编辑 $ENV_FILE 填入强口令，并保证 MYSQL_PASSWORD 与 $CFG 的 database.password 一致；"
    warn "若 MySQL 数据卷已初始化过，新口令不会自动生效，需与旧口令保持一致或手动改密。"
    [ "$ASSUME_YES" = "1" ] || ask "已了解，继续使用当前 .env 部署？" || exit 1
  else
    err "缺少 .env 与 .env.example：docker compose 需要 MYSQL_ROOT_PASSWORD / MYSQL_PASSWORD"; exit 1
  fi
fi

# ---- 2. 推断对外地址 ------------------------------------------------------
PUBLIC_BASE_URL="${PUBLIC_BASE_URL:-}"
if [ -z "$PUBLIC_BASE_URL" ]; then
  IP="$(curl -fsS --max-time 5 https://api.ipify.org 2>/dev/null || true)"
  [ -z "$IP" ] && IP="127.0.0.1"
  PUBLIC_BASE_URL="http://${IP}"
  log "未指定 PUBLIC_BASE_URL，自动推断为 ${PUBLIC_BASE_URL}（可 export PUBLIC_BASE_URL=... 覆盖，不要带端口）"
fi
case "$PUBLIC_BASE_URL" in
  *127.0.0.1*|*localhost*) warn "当前对外地址是本地回环，若服务器未绑定公网IP，外部将无法访问。" ;;
esac


# ---- 3. 生成 JWT secret（仍为占位符时）-----------------------------------
if grep -q "CHANGE_ME_WITH_RANDOM_SECRET" "$CFG"; then
  SECRET="$(head -c 32 /dev/urandom | od -An -tx1 | tr -d ' \n')"
  sed -i.bak -E "s#(secret:[[:space:]]*).*#\1\"${SECRET}\"#" "$CFG"
  log "已生成随机 jwt.secret"
fi
rm -f "$CFG".bak

# ---- 4. 构建并启动 --------------------------------------------------------
log "开始构建并启动容器（首次较慢，请耐心等待）..."
$DC -f "$COMPOSE_FILE" up -d --build

# ---- 5. 健康检查 ----------------------------------------------------------
log "等待后端 API 就绪..."
OK=0
for i in $(seq 1 60); do
  code="$(curl -s -o /dev/null -w '%{http_code}' --max-time 3 "http://127.0.0.1:${BACKEND_PORT}/api/v1/geo/locations" || true)"
  if [ "$code" = "200" ]; then
    log "后端已就绪：${PUBLIC_BASE_URL}:${BACKEND_PORT}  （HTTP ${code}）"; OK=1; break
  fi
  sleep 2
done
[ "$OK" = "1" ] || err "后端未在预期时间内就绪，请查看： $DC logs -f backend"

log "等待临时前端站点就绪..."
OK=0
for i in $(seq 1 30); do
  code="$(curl -s -o /dev/null -w '%{http_code}' --max-time 3 "http://127.0.0.1:${FRONTEND_PORT}/" || true)"
  if [ "$code" = "200" ] || [ "$code" = "304" ]; then
    log "临时前端已就绪：${PUBLIC_BASE_URL}:${FRONTEND_PORT}/  （HTTP ${code}）"; OK=1; break
  fi
  sleep 2
done
[ "$OK" = "1" ] || err "临时前端未在预期时间内就绪，请查看： $DC logs -f frontend"

log "容器状态："
$DC -f "$COMPOSE_FILE" ps
log "完成。后端 API: ${PUBLIC_BASE_URL}:${BACKEND_PORT}  临时前端: ${PUBLIC_BASE_URL}:${FRONTEND_PORT}"
log "提醒：若启用 5173 正式前端，请把 ${PUBLIC_BASE_URL}:5173 加入 ${CFG} 的 server.cors_allow_origins 并重启后端。"