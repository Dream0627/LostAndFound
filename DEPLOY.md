# LAF 前后端分离部署说明（前端 5173 / 后端 8080）

## 一、部署形态
- 后端 API：`http://<服务器公网IP或域名>:8080`（直接对外，`/uploads` 图片也在此端口）
- 临时前端（frontend-demo）：`http://<服务器公网IP或域名>:9090`（容器内 Caddy 同源反代后端，仅临时使用）
- 正式前端（5173）：由前端同学构建后部署到前端服务器，前端 Caddy 将 `/api`、`/uploads` 反代到 8080 后端
- MySQL 仅容器内网可达，不对宿主机暴露端口

## 二、访问方式（Caddy 反向代理，无 CORS）
浏览器只同源访问前端，由前端服务器上的 Caddy 把 API 与图片反代到后端，后端无需（也不再）配置 CORS：

```caddyfile
# 前端服务器 Caddyfile（HTTP 5173；填域名则 Caddy 自动签 HTTPS 证书）
:5173 {
	root * /var/www/laf-frontend        # 前端构建产物(dist)目录

	@backend path /api/* /uploads/*
	handle @backend {
		reverse_proxy <后端IP或域名>:8080   # 保持路径，后端路由前缀为 /api/v1
	}

	handle {
		try_files {path} /index.html       # SPA 路由回退，避免刷新 404
		file_server
	}
}
```

- 前端构建时 `VITE_API_BASE_URL=/api/v1`（相对路径，见 `frontend-demo/.env.production`），**不要**填后端完整地址。
- 后端已移除 CORS 中间件：浏览器直连后端地址的跨域请求会被浏览器拦截，前端必须走 Caddy 反代。
- 前提：后端 8080 端口需在安全组中对前端服务器 IP 放行（Caddy 才能连上后端）。

## 三、涉及文件
| 文件 | 作用 |
|---|---|
| `Dockerfile` | 后端镜像（Go 多阶段构建） |
| `frontend-demo/Dockerfile` | 临时前端镜像（Node 构建 + Caddy 托管） |
| `frontend-demo/Caddyfile` | SPA 回退 + `/api`、`/uploads` 反代 |
| `config/config.docker.yaml` | 容器用配置（`database.host=mysql`，无 CORS） |
| `.env.example` → `.env` | MySQL 口令（compose 自动读取；`.env` 已 gitignore，不上传公开仓库） |
| `docker-compose.yml` | 三服务编排（mysql + backend + frontend），口令通过 `${MYSQL_*}` 注入 |
| `migrations/tables.sql` | 首次启动自动建表 |
| `deploy.sh` | 服务器端一键部署脚本（自动装 Docker / 加 swap / 写配置 / 健康检查） |

## 四、一键部署（推荐）
```bash
cd /opt/LAF
# 1) 配置 MySQL 口令（deploy.sh 也会在缺失时自动从模板生成 .env 并提示）：
cp .env.example .env && vi .env   # MYSQL_PASSWORD 必须与下方 config.docker.yaml 的 database.password 一致
# 2) 编辑 config/config.docker.yaml：jwt.secret 改随机串
bash deploy.sh
```
首次运行若未装 Docker，脚本会询问并自动安装；2G 内存机器会询问是否加 2GB swap。
非交互执行：`ASSUME_YES=1 bash deploy.sh`

## 五、手动部署（等价步骤）
```bash
# 1. 改 config/config.docker.yaml：
#    jwt.secret            -> openssl rand -hex 32 生成
docker compose up -d --build
docker compose ps
```

## 六、阿里云 ECS 专项：上线前必须开通/确认的项
1. **公网可达**：ECS 实例需绑定公网 IP / EIP，带宽 ≥ 1 Mbps；或用「域名 + 备案 + 解析到该 IP」。
2. **安全组放行**：ECS → 实例 → 「网络与安全组」→ 安全组 → 入方向，添加规则：
   - `8080/8080`（后端 API，前端 Caddy 反代访问必经）
   - `5173/5173`（正式前端部署到本机时）
   - `9090/9090`（临时前端，可只在需要时开放）
   授权对象按需：`0.0.0.0/0` 或指定开发同学 IP 段。
3. **SSH 登录信息**：ECS → 实例 → 「远程连接」，或「重置密码」设置 root 密码；若用密钥对则下载 `.pem`。
4. **Docker**：未安装则由 `deploy.sh` 自动安装；也可手动 `curl -fsSL https://get.docker.com | sh`。
5. **内存**：2 GiB 机器建议加 2GB swap（`deploy.sh` 已内置自动检测与创建）。

## 七、注意事项
- **定位功能需要 HTTPS**：经 `http://<IP>`（非 localhost）访问时，浏览器会禁用 `navigator.geolocation`，一键定位不可用，仅能手动选点。需要该功能请配置域名 + TLS。
- 首次启动 MySQL 数据卷为空时会自动执行 `migrations/tables.sql`；之后重启不重复执行。
- 重建/更新：`docker compose up -d --build`；停止：`docker compose down`；连数据一起清除：`docker compose down -v`。
- 查看日志：`docker compose logs -f backend` / `docker compose logs -f frontend`。
- 正式前端（5173）构建时，`VITE_API_BASE_URL` 应指向后端完整地址，如 `http://<公网IP>:8080/api/v1`。
