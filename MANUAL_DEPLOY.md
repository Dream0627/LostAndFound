# LAF 手动部署步骤（阿里云 ECS · 单端口 8080）

## 〇、服务器现状（据控制台 / 终端）
- 实例 `Dream`，Ubuntu 24.04 LTS x86_64，2 vCPU / 2 GiB，磁盘约 39 GiB
- 私网 IP `172.22.0.139`；**无公网 IP**，经控制台「远程连接(Workbench)」登录为 root
- 安全组入方向已放行 `22` 与 `8080`

---

## 第 0 步 ⚠️ 先确认服务器是否有外网（必做）
在线安装 Docker、以及构建镜像时的 `npm install` / `go mod download` 都需要外网。
在服务器终端执行：

```bash
curl -sS --max-time 8 -o /dev/null -w "%{http_code}\n" https://get.docker.com
```

- 输出 `200` / `301` / `302` → 有外网，继续第 1 步
- 超时 / 报错 → **无外网**。需先在阿里云控制台给实例**绑定公网 IP（EIP，按量计费，带宽 ≥ 1 Mbps）**或配置 NAT 网关，否则无法在线部署。
- 提示：你的启动日志出现 `Failed to connect to https://changelogs.ubuntu.com`，很可能当前就**没有外网**，请务必先跑上面这条命令确认。

---

## 第 1 步 上传项目到服务器
项目在本机 `E:\study\LAF`。因无公网 IP，本机不能直接 scp，推荐用控制台上传：

1. 本机把项目打包（**务必排除** `frontend-demo/node_modules`、`frontend-demo/dist`）
2. 在 Workbench 终端界面用「上传文件」把压缩包传到 `/root`
3. 在服务器解压：

```bash
mkdir -p /opt && cd /opt
# 压缩包是 zip：
command -v unzip >/dev/null && unzip -o /root/LAF.zip -d /opt/ \
  || python3 -m zipfile -e /root/LAF.zip /opt/
# 压缩包是 tar.gz：
# tar -xzf /root/LAF.tar.gz -C /opt/
cd /opt/LAF
```

（若你绑定了 EIP，也可直接从本机 `scp -r E:\study\LAF root@<公网IP>:/opt/`）

---

## 第 2 步 安装 Docker

```bash
curl -fsSL https://get.docker.com | sh
systemctl enable --now docker
docker --version && docker compose version
```

（也可跳过此步，`deploy.sh` 会询问并自动安装。）

---

## 第 3 步 一键部署

```bash
cd /opt/LAF
export PUBLIC_BASE_URL="http://172.22.0.139:8080"
bash deploy.sh
```

脚本会自动：检测内存/swap（2 GiB 会提示创建 2GB swap）→ 检测/安装 Docker → 写入对外地址与随机 JWT → 构建并启动 → 健康检查。

---

## 第 4 步 验证

```bash
docker compose ps                # 三个容器均 Up / running
curl -I http://127.0.0.1:8080    # 返回 200 / 304
```

内网访问：`http://172.22.0.139:8080`

---

## 常用运维

```bash
docker compose logs -f backend     # 后端日志
docker compose logs -f frontend    # 前端日志
docker compose up -d --build       # 改代码后重建
docker compose down                # 停止（保留数据）
docker compose down -v             # 停止并删除数据卷
```

---

## 注意事项
- **无外网**：无法在线安装 Docker / 构建镜像，需先绑定 EIP 或配置 NAT。
- **定位功能**：经 `http://<IP>`（非 localhost / HTTPS）访问时，浏览器会禁用定位，只能手动选地点；自动定位需 HTTPS。
- **内存**：2 GiB 建议加 2GB swap（`deploy.sh` 已内置自动检测）。
- **换行符**：`deploy.sh` 必须保持 LF。若在 Windows 上编辑过，先执行 `sed -i 's/\r$//' deploy.sh`。
