#!/bin/bash
# =============================================================================
# StudentHub · CentOS 7.9 一键部署脚本
# -----------------------------------------------------------------------------
# 使用方式 (在 CentOS 7.9 服务器上执行):
#   1. 将整个 StudentHub 项目目录上传到 /opt/studenthub
#   2. chmod +x deploy/centos_deploy.sh && sudo bash deploy/centos_deploy.sh
#
# 前置条件:
#   - CentOS 7.9 x86_64
#   - Docker 已安装并启动 (sudo systemctl start docker)
#   - 至少 2GB 可用内存 (构建阶段需要)
#   - 至少 5GB 可用磁盘
# =============================================================================

set -e

# -------- 颜色输出 --------
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m'

log_info()  { echo -e "${GREEN}[INFO]${NC}  $*"; }
log_warn()  { echo -e "${YELLOW}[WARN]${NC}  $*"; }
log_error() { echo -e "${RED}[ERROR]${NC} $*"; }

# -------- 配置变量 --------
PROJECT_DIR="/opt/studenthub"
COMPOSE_FILE="${PROJECT_DIR}/docker-compose.yml"
ENV_FILE="${PROJECT_DIR}/deploy/.env"

echo "=================================================="
echo "  StudentHub · CentOS 7.9 部署脚本"
echo "  $(date '+%Y-%m-%d %H:%M:%S')"
echo "=================================================="

# =============================================================================
# Step 1 · 环境检查
# =============================================================================
log_info "[1/6] 检查运行环境..."

# 1.1 检查是否 root
if [ "$(id -u)" -ne 0 ]; then
    log_error "请使用 root 或 sudo 运行此脚本"
    exit 1
fi

# 1.2 检查 Docker 是否运行
if ! systemctl is-active --quiet docker 2>/dev/null; then
    log_warn "Docker 未运行, 尝试启动..."
    systemctl start docker
    sleep 2
    if ! systemctl is-active --quiet docker; then
        log_error "Docker 启动失败, 请检查 systemctl status docker"
        exit 1
    fi
fi
log_info "Docker 版本: $(docker --version)"

# 1.2.5 配置 Docker 镜像加速 (阿里云)
DAEMON_JSON="/etc/docker/daemon.json"
if [ ! -f "$DAEMON_JSON" ] || ! grep -q "mirrors" "$DAEMON_JSON" 2>/dev/null; then
    log_warn "未配置 Docker 镜像加速, 正在设置阿里云加速器..."
    mkdir -p /etc/docker
    if [ -f "$DAEMON_JSON" ]; then
        # 已有文件但无 mirrors, 备份后重写
        cp "$DAEMON_JSON" "${DAEMON_JSON}.bak.$(date +%s)"
    fi
    cat > "$DAEMON_JSON" << 'EOF'
{
  "registry-mirrors": [
    "https://registry.aliyuncs.com",
    "https://hub-mirror.c.163.com",
    "https://mirror.baidubce.com"
  ],
  "log-driver": "json-file",
  "log-opts": {
    "max-size": "50m",
    "max-file": "10"
  }
}
EOF
    systemctl daemon-reload
    systemctl restart docker
    sleep 3
    log_info "Docker 镜像加速已配置并重启服务"
else
    log_info "Docker 镜像加速已配置"
fi

# 1.3 检查 Docker Compose (支持 v1 和 v2)
COMPOSE_CMD=""
if docker compose version >/dev/null 2>&1; then
    COMPOSE_CMD="docker compose"
    log_info "Docker Compose: v2 (plugin)"
elif docker-compose --version >/dev/null 2>&1; then
    COMPOSE_CMD="docker-compose"
    log_info "Docker Compose: v1 (standalone)"
else
    log_error "未找到 docker compose 或 docker-compose, 正在安装..."
    # CentOS 7 安装 docker-compose v1 (Python pip)
    curl -L "https://github.com/docker/compose/releases/download/1.29.2/docker-compose-$(uname -s)-$(uname -m)" \
        -o /usr/local/bin/docker-compose
    chmod +x /usr/local/bin/docker-compose
    COMPOSE_CMD="docker-compose"
    log_info "Docker Compose v1 安装完成"
fi

# 1.4 检查项目目录
if [ ! -f "$COMPOSE_FILE" ]; then
    log_error "未找到 ${COMPOSE_FILE}, 请确认项目已上传到 ${PROJECT_DIR}"
    exit 1
fi

# 1.5 检查 .env 文件
if [ ! -f "$ENV_FILE" ]; then
    log_warn "未找到 ${ENV_FILE}, 从 .env.example 生成..."
    cp "${PROJECT_DIR}/deploy/.env.example" "$ENV_FILE"
    log_warn "!!! 请立即编辑 ${ENV_FILE}, 修改 JWT_SECRET 和 CRYPTOX_KEY !!!"
    log_warn "生成命令: openssl rand -base64 48  # JWT_SECRET"
    log_warn "生成命令: openssl rand -base64 32  # CRYPTOX_KEY"
    echo ""
    read -r -p "已编辑好 .env 文件? (y/n) " -n 1 -r
    echo
    if [[ ! $REPLY =~ ^[Yy]$ ]]; then
        log_error "请先配置 deploy/.env 后再执行"
        exit 1
    fi
fi

# 1.6 检查内存 (构建至少需要 2GB)
TOTAL_MEM=$(grep MemTotal /proc/meminfo | awk '{print $2}')
TOTAL_MEM_MB=$((TOTAL_MEM / 1024))
if [ "$TOTAL_MEM_MB" -lt 1800 ]; then
    log_warn "服务器内存仅 ${TOTAL_MEM_MB}MB, Docker 构建可能失败"
    log_warn "建议: 在本地高配机器构建镜像, 推送至阿里云 ACR, 再在此服务器 pull"
    echo ""
    read -r -p "内存不足, 是否仍继续尝试构建? (y/n) " -n 1 -r
    echo
    if [[ ! $REPLY =~ ^[Yy]$ ]]; then
        exit 1
    fi
fi

# =============================================================================
# Step 2 · 防火墙 / 安全组检查
# =============================================================================
log_info "[2/6] 检查端口可达性..."

# 从 .env 读取 HOST_PORT
HOST_PORT=$(grep -E '^HOST_PORT=' "$ENV_FILE" | cut -d= -f2)
HOST_PORT=${HOST_PORT:-8080}

# CentOS 7 firewalld
if systemctl is-active --quiet firewalld 2>/dev/null; then
    if ! firewall-cmd --list-ports 2>/dev/null | grep -q "${HOST_PORT}/tcp"; then
        log_warn "firewalld 未放行端口 ${HOST_PORT}, 正在添加..."
        firewall-cmd --zone=public --add-port="${HOST_PORT}/tcp" --permanent
        firewall-cmd --reload
        log_info "已放行端口 ${HOST_PORT}"
    else
        log_info "firewalld 已放行端口 ${HOST_PORT}"
    fi
else
    log_info "firewalld 未运行, 跳过"
fi

# 提醒阿里云安全组
log_warn "!!! 请确保阿里云安全组已放行入方向 TCP ${HOST_PORT} !!!"

# =============================================================================
# Step 3 · 停止旧容器 (如果存在)
# =============================================================================
log_info "[3/6] 停止旧容器..."

cd "$PROJECT_DIR"

if $COMPOSE_CMD ps 2>/dev/null | grep -q "Up"; then
    log_info "检测到运行中的旧容器, 正在停止..."
    $COMPOSE_CMD down --remove-orphans
fi

# =============================================================================
# Step 4 · 构建镜像
# =============================================================================
log_info "[4/6] 构建 Docker 镜像 (可能需要 5-15 分钟)..."

# 清理旧的悬虚镜像以释放磁盘空间
docker image prune -f >/dev/null 2>&1 || true

BUILD_START=$(date +%s)
$COMPOSE_CMD build 2>&1 | while IFS= read -r line; do
    echo "  $line"
done
BUILD_END=$(date +%s)
BUILD_TIME=$((BUILD_END - BUILD_START))
log_info "镜像构建完成, 耗时 ${BUILD_TIME}s"

# =============================================================================
# Step 5 · 启动服务
# =============================================================================
log_info "[5/6] 启动 StudentHub 服务..."

$COMPOSE_CMD up -d --remove-orphans

# 等待容器启动（数据库迁移 + seed 数据需要时间）
sleep 10

# =============================================================================
# Step 6 · 健康检查
# =============================================================================
log_info "[6/6] 健康检查..."

HEALTH_URL="http://127.0.0.1:${HOST_PORT}/api/v1/healthz"
MAX_WAIT=120
WAITED=0

while [ "$WAITED" -lt "$MAX_WAIT" ]; do
    if curl -sf "$HEALTH_URL" >/dev/null 2>&1; then
        log_info "健康检查通过! (${WAITED}s)"
        break
    fi
    sleep 5
    WAITED=$((WAITED + 5))
    echo "  等待中... ${WAITED}s"
done

if [ "$WAITED" -ge "$MAX_WAIT" ]; then
    log_error "健康检查超时 (${MAX_WAIT}s), 请检查日志:"
    log_error "  $COMPOSE_CMD logs --tail=50"
    $COMPOSE_CMD logs --tail=30
    exit 1
fi

# =============================================================================
# 部署完成
# =============================================================================
SERVER_IP=$(curl -s ifconfig.me 2>/dev/null || echo "YOUR_SERVER_IP")

echo ""
echo "=================================================="
echo -e "  ${GREEN}StudentHub 部署成功!${NC}"
echo "=================================================="
echo "  访问地址: http://${SERVER_IP}:${HOST_PORT}"
echo "  健康检查: ${HEALTH_URL}"
echo ""
echo "  常用命令:"
echo "    查看日志:   cd ${PROJECT_DIR} && ${COMPOSE_CMD} logs -f"
echo "    重启服务:   cd ${PROJECT_DIR} && ${COMPOSE_CMD} restart"
echo "    停止服务:   cd ${PROJECT_DIR} && ${COMPOSE_CMD} down"
echo "    查看状态:   cd ${PROJECT_DIR} && ${COMPOSE_CMD} ps"
echo ""
echo "  日志文件: ${PROJECT_DIR}/backend/logs/"
echo "  数据库:   ${PROJECT_DIR}/backend/data/studenthub.db (Docker卷)"
echo ""
echo "  !!! 重要提醒 !!!"
echo "  1. 建议配置 Nginx 反向代理, 启用 HTTPS"
echo "  2. 定期备份 Docker 卷: docker run --rm -v studenthub_data:/data -v \$(pwd):/backup alpine tar czf /backup/studenthub_data_\$(date +%F).tgz -C /data ."
echo "  3. 首次使用请登录管理员账号创建角色和用户"
echo "=================================================="
