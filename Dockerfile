# syntax=docker/dockerfile:1

# ============================================================================
# LAF 后端镜像（多阶段构建）
# 阶段一用官方 Go 镜像编译出静态可执行文件；
# 阶段二只保留二进制与运行期资源，让最终镜像尽量小，
# 且不把整套工具链和源码带进生产环境。
# ============================================================================

# ---- 构建阶段 ----
FROM golang:1.26.5-alpine AS builder

WORKDIR /src

# 先只拷贝依赖清单，让 go mod download 这一层能被缓存（源码改动不触发重新下载依赖）。
COPY go.mod go.sum ./
# 国内拉依赖走镜像源，避免 proxy.golang.org 不可达；构建环境可直连时删掉此行即可。
RUN go env -w GOPROXY=https://goproxy.cn,direct && go mod download

# 再拷贝其余源码进行编译。
COPY . .

# CGO_ENABLED=0：本项目依赖均为纯 Go（gin / gorm / go-sql-driver 等），可静态编译，
# 运行阶段用 alpine 也无需 libc 兼容层；-s -w 去掉符号表与调试信息以减小体积。
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /out/laf .

# ---- 运行阶段 ----
FROM alpine:3.20

# ca-certificates：如后续需要外呼 HTTPS；tzdata：日志时间与 DSN 的 loc=Local 会用到时区。
RUN apk add --no-cache ca-certificates tzdata \
	&& ln -sf /usr/share/zoneinfo/Asia/Shanghai /etc/localtime

WORKDIR /app

# 二进制本体。
COPY --from=builder /out/laf /app/laf

# 运行期需要的非代码资源：
#  - config/：config.Load() 固定读取 config/config.yaml，这里放一份 .example 作模板；
#            真正的 config.yaml 建议由编排/挂载注入，不要打进镜像（避免密钥入库）。
#  - migrations/：建库建表 SQL，供首次初始化数据库时使用。
COPY config/config.example.yaml /app/config/config.example.yaml
COPY migrations /app/migrations

# 帖子图片写盘目录（engine.Static 把 ./uploads 映射为 /uploads 静态资源）。
RUN mkdir -p /app/uploads/posts

EXPOSE 8080

# 启动；配置从 /app/config/config.yaml 读取（由挂载或构建期提供）。
CMD ["/app/laf"]
