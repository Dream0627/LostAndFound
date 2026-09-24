#!/usr/bin/env bash
# ============================================================================
# 【在“有外网 + 已装 Docker”的机器上执行】
# 作用：构建前后端镜像，并连同 mysql 基础镜像一起导出为离线 tar 包。
# 产物：dist-offline/laf-images.tar  （连同本项目一起拷到服务器）
# 用法：在项目根目录执行  bash offline/build-and-save.sh
# ============================================================================
set -euo pipefail
cd "$(dirname "$0")/.."   # 回到项目根目录

OUT_DIR="dist-offline"
TAR="$OUT_DIR/laf-images.tar"
mkdir -p "$OUT_DIR"

echo "[1/3] 拉取基础镜像（mysql / golang / node / nginx / alpine）..."
docker pull mysql:8.0
docker pull golang:1.26.5-alpine
docker pull node:20-alpine
docker pull nginx:1.27-alpine
docker pull alpine:3.20

echo "[2/3] 构建前后端镜像..."
docker compose build

echo "[3/3] 导出镜像到 $TAR （体积较大，请耐心等待）..."
docker save -o "$TAR" laf-backend:latest laf-frontend:latest mysql:8.0

echo "完成。产物："
ls -lh "$TAR"
echo
echo "下一步：把整个项目目录（含 $OUT_DIR，但可排除 frontend-demo/node_modules）"
echo "上传到服务器，然后在服务器执行  bash offline/load-and-run.sh"