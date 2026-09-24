# LAF 前后端分离部署说明（单端口 8080）

## 一、部署形态
- 唯一对外入口：`http://<服务器公网IP或域名>:8080` → 前端 Nginx（容器内 80）
- `/api`、`/uploads` 由 Nginx 同源反向代理到后端容器 `backend:8080`（浏览器全程只接触 8080，无跨域）
- MySQL 仅容器内网可达，不对宿主机暴露端口

## 二、涉及文件
| 文件 | 作用 |
|---|---|
| `Dockerfile` | 后端镜像（Go 多阶段构建） |
| `frontend-demo/Dockerfile` | 前端镜像（Node 构建 + Nginx 托管） |
| `frontend-demo/nginx.conf` | SPA 回退 + `/api`、`/uploads` 反代 |
| `frontend-demo/.env.production` | 前端构建变量 `VITE_API_BASE_URL=/api/v1` |
| `config/config.docker.yaml` | 容器用配置（`database.host=mysql`） |
| `docker-compose.yml` | 三服务编排（mysql + backend + frontend） |
| `migrations/tables.sql` | 首次启动自动建表 |
| `deploy.sh` | 服务器端一键部署脚本（自动装 Docker / 加 swap / 写配置 / 健康检查） |

## 三、一键部署（推荐）
```bash
cd /opt/LAF
export PUBLIC_BASE_URL="http://<你的公网IP或域名>:8080"   # 可选，不填则自动探测公网IP
bash deploy.sh
```
首次运行若未装 Docker，脚本会询问并自动安装；2G 内存机器会询问是否加 2GB swap。
非交互执行：`ASSUME_YES=1 bash deploy.sh`

## 四、手动部署（等价步骤）
```bash
# 1. 改 config/config.docker.yaml：
#    server.public_base_url -> http://<公网IP>:8080
#    jwt.secret            -> openssl rand -hex 32 生成
docker compose up -d --build
docker compose ps
```

## 五、阿里云 ECS 专项：上线前必须开通/确认的项
（以下对应控制台路径，按需操作）

1. **公网可达（最关键）**：当前实例未绑定公网 IP、带宽 0 Mbps，部署后外网无法访问。
   - 控制台：ECS → 实例 → 「网络与安全组」→「绑定弹性公网IP」或「分配公网IP(固定/按量)」；
   - 或购买 EIP 后绑定；带宽建议 ≥ 1 Mbps；
   - 或用「域名 + 备案 + 解析到该 IP」。
2. **安全组放行 8080**：ECS → 实例 → 「网络与安全组」→ 安全组 → 入方向 → 添加规则：
   协议 TCP，端口范围 `8080/8080`，授权对象 `0.0.0.0/0`。（现有规则只放了 80/22/3389）
3. **SSH 登录信息**：ECS → 实例 → 「远程连接」，或「重置密码」设置 root 密码；若用密钥对则下载 `.pem`。
4. **Docker**：未安装则由 `deploy.sh` 自动安装；也可手动 `curl -fsSL https://get.docker.com | sh`。
5. **内存**：本机 2 GiB，建议加 2GB swap（`deploy.sh` 已内置自动检测与创建）。

## 六、注意事项
- **定位功能需要 HTTPS**：经 `http://<IP>`（非 localhost）访问时，浏览器会禁用 `navigator.geolocation`，一键定位不可用，仅能手动选点。需要该功能请配置域名 + TLS。
- 首次启动 MySQL 数据卷为空时会自动执行 `migrations/tables.sql`；之后重启不重复执行。
- 重建/更新：`docker compose up -d --build`；停止：`docker compose down`；连数据一起清除：`docker compose down -v`。
- 查看日志：`docker compose logs -f backend` / `docker compose logs -f frontend`。