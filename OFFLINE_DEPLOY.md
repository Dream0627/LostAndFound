# LAF 离线部署（服务器无外网 · 单端口 8080）

> 适用场景：目标服务器无法访问外网（已确认 `curl https://get.docker.com` 超时）。
> 核心思路：在**能上网的机器**上构建并导出镜像 → 传到服务器 → 离线导入运行。
> 目标机器：内网 `http://172.22.0.139:8080`（前端为唯一入口，后端不对外暴露端口）。

## 为什么不能直接在线部署
在线部署需要服务器出网：安装 Docker、拉取基础镜像、`npm install`、`go mod download`。
本服务器出网被阻断，故必须走离线镜像路线。**内网访问不受影响**。

## 必备基础镜像（构建机会自动拉取）
| 镜像 | 用途 | 来源 |
|---|---|---|
| `golang:1.26.5-alpine` | 后端编译阶段 | 构建机 |
| `alpine:3.20` | 后端运行阶段 | 构建机 |
| `node:20-alpine` | 前端构建阶段 | 构建机 |
| `nginx:1.27-alpine` | 前端运行阶段 | 构建机 |
| `mysql:8.0` | 数据库 | 构建机（随包导出） |

---

## 步骤 A：在“有外网 + Docker”的机器上构建并导出

### 方式一（推荐）：Windows 一键脚本
在「项目根目录」双击：

```
offline\build-and-save.bat
```

脚本会自动：校验 Docker → 拉取基础镜像 → 构建前后端镜像 → 导出镜像包。
产物：`dist-offline\laf-images.tar`

### 方式二：命令行（Git Bash / WSL / Linux / macOS）
```bash
# 在项目根目录
bash offline/build-and-save.sh
# 产物：dist-offline/laf-images.tar
```

> 说明：镜像包通常数百 MB（含后端 alpine 运行层、前端 nginx 层、`mysql:8.0`）。
> 多阶段构建只保留运行层，构建用的 golang / node 层**不会**打进产物。

---

## 步骤 B：把项目拷到服务器
因服务器无公网，用阿里云控制台 **Workbench 的「上传文件」** 把项目压缩包传到服务器：

```bash
# 服务器上（Workbench 终端）
mkdir -p /opt && cd /opt
command -v unzip >/dev/null && unzip -o /root/LAF.zip -d /opt/ || python3 -m zipfile -e /root/LAF.zip /opt/
cd /opt/LAF
```

打包提示（在构建机上）：
- **必须包含**：`dist-offline/laf-images.tar`、`docker-compose.yml`、`config/config.docker.yaml`、`migrations/`、`offline/`、后端源码与 `frontend-demo/` 源码。
- **可排除**：`frontend-demo/node_modules`、`frontend-demo/dist`、`.git`、`.idea`、`.vscode`。

Windows 打包命令示例（在项目根目录，PowerShell）：
```powershell
Compress-Archive -Path * -DestinationPath ..\LAF.zip -Force
```
（`node_modules` 若存在，请先删除或改用 7-Zip 排除后再压缩。）

---

## 步骤 C：服务器离线导入并启动
```bash
cd /opt/LAF
bash offline/load-and-run.sh
```
脚本会：写入内网 `public_base_url` 与随机 JWT → `docker load` 导入镜像 → `docker compose up -d --no-build` → 健康检查。

---

## 步骤 D：验证
```bash
docker compose ps                 # 三容器均 Up
curl -I http://127.0.0.1:8080     # 200 / 304
```
内网访问：`http://172.22.0.139:8080`

---

## 前提：服务器需已安装 Docker（离线）
若服务器**未装 Docker**，离线环境无法用 `get.docker.com` 在线脚本，可选：

- **推荐**：临时给实例绑定弹性公网 IP（EIP，按量计费），
  `curl -fsSL https://get.docker.com | sh` 装好后**解绑释放 EIP**；
- 或使用 Docker 官方离线 deb 包（`docker-ce` + `containerd.io` + `docker-ce-cli` +
  `docker-buildx-plugin` + `docker-compose-plugin`）逐包 `dpkg -i` 安装
  （需自行下载匹配 Ubuntu 24.04 amd64 的包）。

> 先检查服务器是否已有 Docker：`docker --version`。
> 若有输出则可跳过本节。

---

## 常见问题
- **本机 `docker pull` 很慢/失败**：配置 Docker Desktop 的镜像加速器（registry mirror），或换网络重试。
- **`docker compose build` 报 go/npm 网络错误**：本项目 Dockerfile 已配置国内源
  （`GOPROXY=https://goproxy.cn`、`npm registry=https://registry.npmmirror.com`），一般可直连。
- **上传包过大**：镜像包属正常体积；可用分卷压缩或先单独上传 `laf-images.tar` 再上传其余源码。
- **启动后页面 502**：多为后端未起（MySQL 初始化较慢），等待 1–2 分钟后刷新；
  用 `docker compose logs backend` 查看。
