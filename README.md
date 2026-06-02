# MiniSentry

基于 Go + React 的错误追踪与监控系统，Sentry 的轻量级替代方案。

**线上环境**：[broke-clearance-judges-tag.trycloudflare.com](https://broke-clearance-judges-tag.trycloudflare.com)

## 功能

- **实时错误追踪**：自动捕获并分组前端错误
- **Sentry 兼容 API**：可直接替换 Sentry SDK 上报地址
- **多租户架构**：组织 + 项目两级隔离，基于角色的权限控制
- **用户认证**：JWT 安全认证，支持注册/登录/Token 刷新
- **问题管理**：解决、忽略、分配 issue，支持评论与批量操作
- **性能监控**：页面加载耗时、自定义指标追踪
- **搜索与过滤**：全文搜索 + 多维度筛选（状态/级别/环境）
- **统计总览**：仪表盘概览卡片 + Top 项目排行，30s 自动刷新
- **结构化日志**：slog JSON 格式，可直接接入 ELK/Loki
- **环境指示**：非 production 环境 UI 顶部彩色警告条

## 技术架构

| 层 | 技术栈 |
|----|--------|
| **后端** | Go 1.21+ · Chi Router · GORM · slog |
| **前端** | React 18 · TanStack Router/Query · Tailwind CSS |
| **数据库** | PostgreSQL 15+ · Redis |
| **部署** | Docker Compose（开发/生产双配置） |

## 接入 SDK

```javascript
MiniSentry.init({
    dsn: 'https://broke-clearance-judges-tag.trycloudflare.com/api/<project-id>/store/',
    environment: 'production',
    release: '1.0.0'
});

try {
    riskyOperation();
} catch (error) {
    MiniSentry.captureException(error);
}
```

## 本地开发

### 环境要求

- Docker & Docker Compose（推荐）
- Go 1.21+ / Node.js 18+（裸机开发）
- PostgreSQL 15+

### Docker 启动

```bash
git clone https://github.com/zhangyiy520-droid/minisentry.git
cd minisentry
cp .env.example .env

make dev     # 全部服务 + 热重载
make up      # 后台模式
```

### 开发命令

```bash
make help       # 所有命令
make test       # 全部测试
make logs       # 服务日志
make clean      # 清理容器和卷
```

## 生产部署

```bash
docker-compose -f docker-compose.prod.yml up -d
```

关键环境变量：

```bash
DATABASE_URL=postgres://postgres:password@db:5432/minisentry?sslmode=disable
REDIS_URL=redis://redis:6379
JWT_SECRET=<256-bit-secret>
JWT_ISSUER=minisentry
FRONTEND_URL=https://<your-frontend-domain>
CORS_ORIGINS=https://<your-frontend-domain>
```

## API 概要

```
# 认证
POST   /api/v1/auth/register|login|refresh|logout
GET    /api/v1/auth/profile
PUT    /api/v1/auth/profile

# 组织
GET|POST    /api/v1/organizations
GET|PUT|DELETE /api/v1/organizations/{id}
GET|POST    /api/v1/organizations/{id}/members

# 项目
GET|POST    /api/v1/organizations/{org_id}/projects
GET|PUT|DELETE /api/v1/projects/{id}
POST        /api/v1/projects/{id}/keys/regenerate

# Issue
GET    /api/v1/projects/{id}/issues       # 支持筛选
GET    /api/v1/projects/{id}/issues/stats
GET|PUT /api/v1/issues/{id}
POST   /api/v1/issues/{id}/comments
POST   /api/v1/issues/bulk-update

# 统计 & 上报
GET    /api/v1/stats/overview             # 全局统计
POST   /api/{project_id}/store/           # 错误事件上报（Sentry 兼容）
```

## 项目结构

```
minisentry/
├── backend/
│   ├── cmd/server/          # 入口
│   └── internal/
│       ├── config/          # 配置管理
│       ├── database/        # 数据库 & 迁移
│       ├── models/          # 数据模型
│       ├── services/        # 业务逻辑
│       ├── handlers/        # HTTP 处理器
│       ├── middleware/       # 认证/CORS/安全
│       └── dto/             # 数据传输对象
├── frontend/src/
│   ├── components/          # 可复用组件
│   ├── pages/               # 页面
│   ├── hooks/               # 自定义 Hooks
│   ├── lib/                 # API 客户端/认证
│   ├── stores/              # 状态管理
│   └── types/               # TypeScript 类型
├── examples/                # SDK 示例
├── docker-compose.yml       # 开发环境
├── docker-compose.prod.yml  # 生产环境
└── Makefile
```

## 特性亮点

- **优雅关机**：SIGINT/SIGTERM 信号处理，10 秒请求排空超时
- **结构化日志**：slog JSON 格式，可直接接入 ELK/Loki
- **精简启动**：ASCII banner 替代冗长路由列表

## License

ISC
