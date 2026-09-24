# 卡在 "Pulling fs layer" 的解决办法（大镜像 blob 下载被阻断）

## 现在的问题（已确诊）
`docker images` 为空、`docker system df` 显示 `Images 0 / 0B`，
且 `mysql:8.0` 长时间停在 **`Pulling fs layer`**（十几分钟不进入 `Downloading/Extracting`）。

含义：
- Docker Hub 的**接口**是通的（`hello-world` 秒下 ✅）；
- 但**镜像层 blob 从另一个 CDN 主机下载时被限速/阻断**——大镜像卡在"列层"后无法下载。
- 这**不是脚本问题，也不是"下载慢"**，是网络对 blob 主机的阻断。

## 根治办法：给 Docker 配置镜像加速器（registry-mirrors）
镜像加速器会让你**整个 Docker 的拉取**（含 blob）都走国内镜像，不再直连被阻断的主机。

### 步骤
1. 托盘右键 Docker 图标 → **Settings**（设置）
2. 左侧选 **Docker Engine**（Docker 引擎）
3. 右侧是 JSON 配置。**保留原有其他字段**，加入 `registry-mirrors`，可参考项目里的
   `offline/daemon.json`（本项目已生成，直接用它的内容）：
```json
{
  "registry-mirrors": [
    "https://docker.m.daocloud.io",
    "https://docker.1ms.run",
    "https://dockerproxy.net",
    "https://docker.1panel.live",
    "https://hub.rat.dev"
  ]
}
```
4. 点 **Apply & Restart**，等 Docker 引擎完全重启（托盘图标变绿/稳定）。
5. 重新双击 `offline\build-and-save.bat`。

### 另一种等价做法（直接改配置文件）
截图里 `docker --help` 显示配置目录是 `C:\Users\Administrator\.docker`。
用记事本打开/新建 `C:\Users\Administrator\.docker\daemon.json`，粘贴上面的内容保存，
然后在托盘菜单选 **Restart** 重启 Docker。

## 验证加速器是否生效
```cmd
docker info
```
在输出里找到 `Registry Mirrors:`，应能看到上面配置的地址。看到即生效。

## 如果仍卡住：换镜像源
镜像源可用性会变动。把 `daocloud` 换成下面任一（可多填几个）：
- `https://docker.1ms.run`
- `https://dockerproxy.net`
- `https://hub.rat.dev`
- `https://docker.xuanyuan.me`
- 或你**阿里云/腾讯云专属地址**（`https://xxxx.mirror.aliyuncs.com`），放第一位最稳。
  阿里云获取：控制台 → 容器镜像服务 ACR → 镜像加速器。

## 手动逐条拉（配好加速器后，便于观察）
在 `E:\study\LAF` 打开 CMD：
```cmd
docker pull mysql:8.0
docker pull golang:1.26.5-alpine
docker pull node:20-alpine
docker pull nginx:1.27-alpine
docker pull alpine:3.20
docker compose build
docker save -o dist-offline\laf-images.tar laf-backend:latest laf-frontend:latest mysql:8.0
```
配好加速器后，这些 pull 应能很快看到 `Downloading / Extracting / Pull complete`。

## 其它排查
- **代理**：Settings → Resources → Proxies；公司网络无代理就关闭，避免误走失效代理。
- **Windows 容器**：确认启用 WSL2 后端 + Linux 容器（本项目的镜像是 Linux 的）。
