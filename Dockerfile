# =============================================================================
# StudentHub · 多阶段构建
#   阶段 1: 构建前端 (Node 20 + pnpm)
#   阶段 2: 编译后端 (Go 1.25)
#   阶段 3: 运行时镜像 (Alpine 3.20)  - 集成后端二进制 + 前端 dist
# 镜像内布局:
#   /app/server                    后端二进制
#   /app/configs/config.yaml       配置
#   /app/frontend/dist/            前端静态资源 (Gin NoRoute SPA fallback)
#   /app/data/                     SQLite + WAL/SHM (持久化卷)
#   /app/storage/                  上传文件 (持久化卷)
#   /app/logs/                     日志输出 (持久化卷)
#   /app/deploy/entrypoint.sh      启动脚本
# =============================================================================

# -----------------------------------------------------------------------------
# Stage 1 · 前端构建
# -----------------------------------------------------------------------------
FROM node:20-alpine AS frontend-builder

WORKDIR /build/frontend

# 单独拷贝 lock 文件以充分利用 Docker 缓存
COPY frontend/package.json frontend/pnpm-lock.yaml ./

# 安装 pnpm (使用 npm 全局安装, 避开 corepack 在国内网络下的兼容问题)
# 固定 9.15.4: 锁文件为 lockfileVersion 9.0 (pnpm 9.x 时代), 兼容 Node 20
# pnpm@latest (11.x) 需要 Node >= 22.13, 而 node:20-alpine 内置 v20.20.2
RUN npm config set registry https://registry.npmmirror.com \
    && npm install -g pnpm@9.15.4 \
    && pnpm config set registry https://registry.npmmirror.com \
    && pnpm install --no-frozen-lockfile

# 拷贝源码并构建
COPY frontend/ ./
RUN pnpm run build

# -----------------------------------------------------------------------------
# Stage 2 · 后端编译
# -----------------------------------------------------------------------------
FROM golang:1.25-alpine AS backend-builder

WORKDIR /build/backend

# 单独下载依赖以利用缓存
# 设置国内 Go 模块代理 (阿里云), 加速国内服务器构建
ENV GOPROXY=https://mirrors.aliyun.com/goproxy,direct
COPY backend/go.mod backend/go.sum ./
RUN go mod download

# 拷贝源码并静态编译 (CGO 禁用，SQLite 走 modernc.org/sqlite 替代)
# 注: 当前项目使用 gorm.io/driver/sqlite, 其底层为 mattn/go-sqlite3, 依赖 CGO
#     因此这里保留 CGO 并安装 gcc; 若未来切换为纯 Go driver 可改为 CGO_ENABLED=0
# 使用阿里云 Alpine 镜像源, 国内服务器构建速度提升 10x+
RUN sed -i 's/dl-cdn.alpinelinux.org/mirrors.aliyun.com/g' /etc/apk/repositories \
    && apk add --no-cache gcc musl-dev
COPY backend/ ./
RUN CGO_ENABLED=1 GOOS=linux go build \
        -trimpath \
        -ldflags="-s -w" \
        -o /out/server \
        ./cmd/server

# -----------------------------------------------------------------------------
# Stage 3 · 运行时镜像
# -----------------------------------------------------------------------------
FROM alpine:3.20

# 安装运行时基础依赖: 时区数据 / ca-certificates / wget (健康检查)
# tzdata 保留在镜像内 (~2MB), 避免删除后 zoneinfo 失效
# 使用阿里云 Alpine 镜像源加速
RUN sed -i 's/dl-cdn.alpinelinux.org/mirrors.aliyun.com/g' /etc/apk/repositories \
    && apk add --no-cache \
        ca-certificates \
        tzdata \
        wget \
    && cp /usr/share/zoneinfo/Asia/Shanghai /etc/localtime \
    && echo "Asia/Shanghai" > /etc/timezone

# 创建非 root 用户运行服务 (安全基线)
RUN addgroup -S app && adduser -S app -G app

WORKDIR /app

# 拷贝后端二进制
COPY --from=backend-builder /out/server /app/server

# 拷贝配置
COPY backend/configs/ /app/configs/

# 拷贝前端构建产物 (Gin NoRoute 会回退到 /app/frontend/dist/index.html)
COPY --from=frontend-builder /build/frontend/dist /app/frontend/dist

# 拷贝启动脚本并去除 Windows 换行符 (\r\n -> \n)
COPY deploy/entrypoint.sh /app/deploy/entrypoint.sh
RUN sed -i 's/\r$//' /app/deploy/entrypoint.sh \
    && chmod +x /app/deploy/entrypoint.sh

# 创建运行时数据目录并授权
RUN mkdir -p /app/data /app/storage /app/logs \
    && chown -R app:app /app

USER app

# Gin 默认监听 :8080
EXPOSE 8080

# 容器内健康检查 (基于 /api/v1/healthz)
HEALTHCHECK --interval=30s --timeout=5s --start-period=30s --retries=3 \
    CMD wget -qO- http://127.0.0.1:8080/api/v1/healthz >/dev/null 2>&1 || exit 1

ENTRYPOINT ["/app/deploy/entrypoint.sh"]
