#!/bin/bash
# MiniSentry 一键部署脚本
# 用法: bash deploy.sh

set -e

echo "========================================="
echo "  MiniSentry 一键部署"
echo "========================================="

# 1. 生成 .env（如果不存在）
if [ ! -f .env ]; then
    JWT_SECRET=$(openssl rand -hex 32)
    cat > .env << EOF
POSTGRES_PASSWORD=$(openssl rand -hex 16)
REDIS_PASSWORD=$(openssl rand -hex 16)
JWT_SECRET=$JWT_SECRET
JWT_ISSUER=minisentry
FRONTEND_URL=http://localhost:3000
CORS_ORIGINS=*
EOF
    echo "[OK] 已生成 .env 配置文件"
else
    echo "[OK] .env 已存在，跳过"
fi

# 2. 启动服务
echo ""
echo "[*] 启动 Docker 服务..."
docker compose -f docker-compose.yml up -d --build

echo ""
echo "[*] 等待服务就绪..."
sleep 10

# 3. 检查状态
if curl -s http://localhost:8080/health > /dev/null 2>&1; then
    echo "[OK] 后端 API 已启动: http://localhost:8080"
else
    echo "[!] 后端启动中，请稍等..."
fi

if curl -s http://localhost:3000 > /dev/null 2>&1; then
    echo "[OK] 前端界面已启动: http://localhost:3000"
else
    echo "[!] 前端启动中，请稍等..."
fi

echo ""
echo "========================================="
echo "  本地访问"
echo "  前端: http://localhost:3000"
echo "  API:  http://localhost:8080/health"
echo "========================================="
echo ""
echo "下一步 — 内网穿透（选一个）："
echo ""
echo "  方案1 Cloudflare Tunnel（免费，需域名）"
echo "    cloudflared tunnel --url http://localhost:3000"
echo ""
echo "  方案2 ngrok（免费 1 条隧道）"
echo "    ngrok http 3000"
echo ""
