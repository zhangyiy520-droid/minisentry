# MiniSentry Development Guide

This guide covers the development workflow, architecture, and contributing guidelines for MiniSentry.

## 📋 Table of Contents

- [Development Setup](#development-setup)
- [Architecture Overview](#architecture-overview)
- [Backend Development](#backend-development)
- [Frontend Development](#frontend-development)
- [Database Schema](#database-schema)
- [API Documentation](#api-documentation)
- [Testing Strategy](#testing-strategy)
- [Contributing Guidelines](#contributing-guidelines)
- [Performance Considerations](#performance-considerations)
- [Deployment](#deployment)

## 🛠️ Development Setup

### Prerequisites

- Go 1.21+
- Node.js 18+
- Docker and Docker Compose
- PostgreSQL 15+ (or Docker)
- Redis (optional, or Docker)

### Quick Setup

```bash
# Clone repository
git clone <repository>
cd minisentry

# Copy environment file
cp .env.example .env

# Start development environment
make dev

# Or start services individually
make db-up     # Start database services
make backend   # Start backend
make frontend  # Start frontend
```

### Development Environment URLs

- **Frontend**: http://localhost:3000
- **Backend API**: http://localhost:8080
- **Database**: localhost:5433
- **Redis**: localhost:6379

### Environment Configuration

Edit `.env` file for local development:

```bash
# Database
DATABASE_URL=postgres://postgres:password@localhost:5433/minisentry?sslmode=disable

# Redis (optional)
REDIS_URL=redis://localhost:6379

# JWT
JWT_SECRET=dev-secret-key-change-in-production
JWT_ISSUER=minisentry

# Server
HOST=0.0.0.0
PORT=8080
FRONTEND_URL=http://localhost:3000

# CORS
CORS_ORIGINS=http://localhost:3000,http://localhost:5173
```

## 🏗️ Architecture Overview

MiniSentry follows a clean architecture pattern with clear separation of concerns:

### System Components

```
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   Frontend      │    │   Backend       │    │   Database      │
│   (React SPA)   │◄──►│   (Go API)      │◄──►│  (PostgreSQL)   │
│                 │    │                 │    │                 │
│ • TanStack      │    │ • Chi Router    │    │ • Primary DB    │
│ • TypeScript    │    │ • GORM          │    │ • Redis Cache   │
│ • Tailwind CSS  │    │ • JWT Auth      │    │ • Migrations    │
└─────────────────┘    └─────────────────┘    └─────────────────┘
```

### Backend Architecture Layers

```
┌─────────────────┐
│   HTTP Layer    │ ← Handlers, Middleware, Routing
├─────────────────┤
│  Service Layer  │ ← Business Logic, Validation
├─────────────────┤
│   Data Layer    │ ← Models, Repository Pattern
├─────────────────┤
│ Database Layer  │ ← GORM, PostgreSQL, Redis
└─────────────────┘
```

### Key Design Principles

1. **Separation of Concerns**: Clear boundaries between layers
2. **Dependency Injection**: Services are injected into handlers
3. **Interface-based Design**: Easy testing and mocking
4. **Error Handling**: Consistent error responses
5. **Security First**: Authentication and authorization at every level

## 🔧 Backend Development

### Project Structure

```
backend/
├── cmd/server/              # Application entry point
├── internal/
│   ├── config/             # Configuration management
│   ├── database/           # Database connection & migrations
│   ├── models/             # Data models
│   ├── services/           # Business logic
│   ├── handlers/           # HTTP handlers
│   ├── middleware/         # HTTP middleware
│   └── dto/               # Data Transfer Objects
├── migrations/             # Database migrations
└── test/                  # Test utilities
```

### Adding a New Feature

1. **Define Model** (`internal/models/`)
```go
type NewModel struct {
    Base
    Name        string `json:"name" gorm:"not null"`
    Description string `json:"description"`
}
```

2. **Create Service** (`internal/services/`)
```go
type NewModelService struct {
    db *database.DB
}

func NewNewModelService(db *database.DB) *NewModelService {
    return &NewModelService{db: db}
}

func (s *NewModelService) Create(model *models.NewModel) error {
    return s.db.Create(model).Error
}
```

3. **Create Handler** (`internal/handlers/`)
```go
type NewModelHandler struct {
    service *services.NewModelService
}

func NewNewModelHandler(service *services.NewModelService) *NewModelHandler {
    return &NewModelHandler{service: service}
}

func (h *NewModelHandler) Create(w http.ResponseWriter, r *http.Request) {
    // Implementation
}
```

4. **Register Routes** (in `cmd/server/main.go`)
```go
newModelHandler := handlers.NewNewModelHandler(newModelService)
newModelHandler.RegisterRoutes(r, authMiddleware)
```

### Database Migrations

Create new migration files in `backend/migrations/`:

```sql
-- 003_add_new_table.up.sql
CREATE TABLE new_models (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(255) NOT NULL,
    description TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX idx_new_models_name ON new_models(name);
```

```sql
-- 003_add_new_table.down.sql
DROP TABLE IF EXISTS new_models;
```

### Running Backend Tests

```bash
cd backend

# Run all tests
go test ./...

# Run with verbose output
go test -v ./...

# Run specific package
go test ./internal/services

# Run with coverage
go test -cover ./...
```

### Backend Development Commands

```bash
# Format code
go fmt ./...

# Lint code (requires golangci-lint)
golangci-lint run

# Update dependencies
go mod tidy

# Build binary
go build -o bin/server ./cmd/server

# Run locally
./bin/server
```

## ⚛️ Frontend Development

### Project Structure

```
frontend/src/
├── components/             # Reusable components
│   ├── ui/                # Base UI components
│   ├── forms/             # Form components
│   ├── layout/            # Layout components
│   └── [feature]/         # Feature-specific components
├── pages/                 # Page components
├── hooks/                 # Custom React hooks
├── lib/                   # Utilities
│   ├── api.ts            # API client
│   ├── auth.tsx          # Authentication context
│   └── utils.ts          # Helper functions
├── stores/               # Zustand stores
├── types/                # TypeScript definitions
└── routes/               # Router configuration
```

### Adding a New Page

1. **Create Page Component** (`src/pages/`)
```tsx
import React from 'react';
import { useQuery } from '@tanstack/react-query';
import { api } from '../lib/api';

export function NewPage() {
  const { data, isLoading } = useQuery({
    queryKey: ['new-data'],
    queryFn: () => api.get('/new-endpoint')
  });

  if (isLoading) return <div>Loading...</div>;

  return (
    <div>
      <h1>New Page</h1>
      {/* Content */}
    </div>
  );
}
```

2. **Add Route** (`src/routes/`)
```tsx
import { Route } from '@tanstack/react-router';
import { NewPage } from '../pages/NewPage';

export const newRoute = new Route({
  getParentRoute: () => rootRoute,
  path: '/new',
  component: NewPage,
});
```

3. **Update Router** (include in route tree)

### State Management

- **Server State**: TanStack Query for API data
- **Client State**: Zustand for UI state
- **Authentication**: React Context

Example Zustand store:
```tsx
import { create } from 'zustand';

interface UIStore {
  sidebarOpen: boolean;
  setSidebarOpen: (open: boolean) => void;
}

export const useUIStore = create<UIStore>((set) => ({
  sidebarOpen: true,
  setSidebarOpen: (open) => set({ sidebarOpen: open }),
}));
```

### API Integration

```tsx
import { useMutation, useQuery } from '@tanstack/react-query';
import { api } from '../lib/api';

// Query hook
export function useProjects() {
  return useQuery({
    queryKey: ['projects'],
    queryFn: () => api.get('/api/v1/projects'),
  });
}

// Mutation hook
export function useCreateProject() {
  return useMutation({
    mutationFn: (data) => api.post('/api/v1/projects', data),
    onSuccess: () => {
      queryClient.invalidateQueries(['projects']);
    },
  });
}
```

### Frontend Development Commands

```bash
cd frontend

# Install dependencies
npm install

# Start development server
npm run dev

# Build for production
npm run build

# Run tests
npm test

# Lint code
npm run lint

# Type check
npm run type-check
```

## 🗄️ Database Schema

### Core Tables

#### Users
```sql
CREATE TABLE users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    email VARCHAR(255) UNIQUE NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    name VARCHAR(255) NOT NULL,
    avatar_url VARCHAR(500),
    is_active BOOLEAN DEFAULT true,
    email_verified BOOLEAN DEFAULT false,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);
```

#### Organizations
```sql
CREATE TABLE organizations (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(255) NOT NULL,
    slug VARCHAR(100) UNIQUE NOT NULL,
    description TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE TABLE organization_members (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    role VARCHAR(50) NOT NULL DEFAULT 'member',
    joined_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    UNIQUE(organization_id, user_id)
);
```

#### Projects
```sql
CREATE TABLE projects (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    name VARCHAR(255) NOT NULL,
    slug VARCHAR(100) NOT NULL,
    description TEXT,
    platform VARCHAR(50) NOT NULL DEFAULT 'javascript',
    dsn VARCHAR(255) UNIQUE NOT NULL,
    public_key VARCHAR(255) NOT NULL,
    secret_key VARCHAR(255) NOT NULL,
    is_active BOOLEAN DEFAULT true,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    UNIQUE(organization_id, slug)
);
```

#### Issues & Events
```sql
CREATE TABLE issues (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    project_id UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    fingerprint VARCHAR(255) NOT NULL,
    title VARCHAR(500) NOT NULL,
    culprit VARCHAR(500),
    type VARCHAR(100) NOT NULL,
    level VARCHAR(50) NOT NULL DEFAULT 'error',
    status VARCHAR(50) DEFAULT 'unresolved',
    first_seen TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    last_seen TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    times_seen INTEGER DEFAULT 1,
    assignee_id UUID REFERENCES users(id),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    UNIQUE(project_id, fingerprint)
);

CREATE TABLE events (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    issue_id UUID NOT NULL REFERENCES issues(id) ON DELETE CASCADE,
    project_id UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    event_id VARCHAR(255) NOT NULL,
    timestamp TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    level VARCHAR(50) NOT NULL DEFAULT 'error',
    message TEXT,
    exception_type VARCHAR(255),
    exception_value TEXT,
    stack_trace JSONB,
    request_data JSONB,
    user_context JSONB,
    tags JSONB,
    extra_data JSONB,
    fingerprint VARCHAR(255) NOT NULL,
    release_version VARCHAR(100),
    environment VARCHAR(100) DEFAULT 'production',
    server_name VARCHAR(255),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    UNIQUE(project_id, event_id)
);
```

### Performance Indexes

```sql
-- Issues indexes
CREATE INDEX idx_issues_project_status ON issues(project_id, status);
CREATE INDEX idx_issues_last_seen ON issues(last_seen DESC);
CREATE INDEX idx_issues_times_seen ON issues(times_seen DESC);

-- Events indexes
CREATE INDEX idx_events_issue_id ON events(issue_id);
CREATE INDEX idx_events_project_timestamp ON events(project_id, timestamp DESC);
CREATE INDEX idx_events_fingerprint ON events(fingerprint);
CREATE INDEX idx_events_level ON events(level);
CREATE INDEX idx_events_environment ON events(environment);

-- JSONB indexes
CREATE INDEX idx_events_tags ON events USING GIN(tags);
CREATE INDEX idx_events_user_context ON events USING GIN(user_context);
CREATE INDEX idx_events_stack_trace ON events USING GIN(stack_trace);

-- Full-text search indexes
CREATE INDEX idx_events_message_fts ON events USING GIN(to_tsvector('english', message));
CREATE INDEX idx_issues_title_fts ON issues USING GIN(to_tsvector('english', title));
```

## 📡 API Documentation

### Authentication

All API endpoints except health checks and error ingestion require JWT authentication.

**JWT Token Format:**
```json
{
  "sub": "user_id",
  "iss": "minisentry",
  "exp": 1234567890,
  "iat": 1234567890,
  "email": "user@example.com"
}
```

**Authorization Header:**
```
Authorization: Bearer <jwt_token>
```

### Response Format

**Success Response:**
```json
{
  "id": "uuid",
  "name": "Example",
  "created_at": "2024-01-01T00:00:00Z"
}
```

**Error Response:**
```json
{
  "error": "Bad Request",
  "message": "Validation failed",
  "details": {
    "field": "email",
    "issue": "required"
  }
}
```

### Authentication Endpoints

#### POST /api/v1/auth/register
Register a new user.

**Request:**
```json
{
  "name": "John Doe",
  "email": "john@example.com",
  "password": "securepassword"
}
```

**Response (201):**
```json
{
  "token": "jwt_token",
  "refresh_token": "refresh_token",
  "user": {
    "id": "uuid",
    "email": "john@example.com",
    "name": "John Doe"
  }
}
```

#### POST /api/v1/auth/login
Authenticate user and return JWT token.

**Request:**
```json
{
  "email": "john@example.com",
  "password": "securepassword"
}
```

**Response (200):**
```json
{
  "token": "jwt_token",
  "refresh_token": "refresh_token",
  "user": {
    "id": "uuid",
    "email": "john@example.com",
    "name": "John Doe"
  }
}
```

### Organization Endpoints

#### GET /api/v1/organizations
List user's organizations.

**Response (200):**
```json
[
  {
    "id": "uuid",
    "name": "Acme Corp",
    "slug": "acme-corp",
    "description": "Main organization",
    "role": "owner",
    "created_at": "2024-01-01T00:00:00Z"
  }
]
```

#### POST /api/v1/organizations
Create a new organization.

**Request:**
```json
{
  "name": "New Organization",
  "description": "Organization description"
}
```

**Response (201):**
```json
{
  "id": "uuid",
  "name": "New Organization",
  "slug": "new-organization",
  "description": "Organization description",
  "role": "owner",
  "created_at": "2024-01-01T00:00:00Z"
}
```

### Project Endpoints

#### GET /api/v1/organizations/{org_id}/projects
List organization projects.

**Query Parameters:**
- `limit` (int): Number of results (default: 25)
- `offset` (int): Pagination offset

**Response (200):**
```json
[
  {
    "id": "uuid",
    "name": "Web App",
    "slug": "web-app",
    "platform": "javascript",
    "dsn": "http://localhost:8080/api/project-id/store/",
    "is_active": true,
    "created_at": "2024-01-01T00:00:00Z"
  }
]
```

#### POST /api/v1/organizations/{org_id}/projects
Create a new project.

**Request:**
```json
{
  "name": "Mobile App",
  "description": "iOS/Android app",
  "platform": "react-native"
}
```

**Response (201):**
```json
{
  "id": "uuid",
  "name": "Mobile App",
  "slug": "mobile-app",
  "platform": "react-native",
  "dsn": "http://localhost:8080/api/project-id/store/",
  "public_key": "public_key",
  "is_active": true,
  "created_at": "2024-01-01T00:00:00Z"
}
```

### Issue Endpoints

#### GET /api/v1/projects/{project_id}/issues
List project issues with filtering.

**Query Parameters:**
- `status`: `resolved` | `unresolved` | `ignored`
- `level`: `debug` | `info` | `warning` | `error` | `fatal`
- `environment`: Filter by environment
- `search`: Full-text search in issue titles
- `sort`: `first_seen` | `last_seen` | `times_seen`
- `limit`: Number of results (default: 25)
- `cursor`: Pagination cursor

**Response (200):**
```json
{
  "data": [
    {
      "id": "uuid",
      "title": "TypeError: Cannot read property 'x' of undefined",
      "culprit": "app.js in handleClick",
      "level": "error",
      "status": "unresolved",
      "first_seen": "2024-01-01T10:00:00Z",
      "last_seen": "2024-01-01T15:30:00Z",
      "times_seen": 42,
      "assignee": {
        "id": "uuid",
        "name": "Jane Doe",
        "email": "jane@example.com"
      }
    }
  ],
  "pagination": {
    "next_cursor": "cursor_string",
    "has_more": true
  }
}
```

#### GET /api/v1/issues/{issue_id}
Get issue details.

**Response (200):**
```json
{
  "id": "uuid",
  "title": "TypeError: Cannot read property 'x' of undefined",
  "culprit": "app.js in handleClick",
  "level": "error",
  "status": "unresolved",
  "first_seen": "2024-01-01T10:00:00Z",
  "last_seen": "2024-01-01T15:30:00Z",
  "times_seen": 42,
  "assignee": null,
  "project": {
    "id": "uuid",
    "name": "Web App"
  },
  "latest_event": {
    "id": "uuid",
    "timestamp": "2024-01-01T15:30:00Z",
    "message": "TypeError: Cannot read property 'x' of undefined",
    "stack_trace": [...],
    "user_context": {...},
    "tags": {...}
  }
}
```

### Error Ingestion

#### POST /api/{project_id}/store/
Ingest error events (Sentry-compatible endpoint).

**Request:**
```json
{
  "event_id": "event_uuid",
  "timestamp": "2024-01-01T10:00:00Z",
  "level": "error",
  "platform": "javascript",
  "exception": {
    "values": [
      {
        "type": "TypeError",
        "value": "Cannot read property 'x' of undefined",
        "stacktrace": {
          "frames": [
            {
              "filename": "app.js",
              "function": "handleClick",
              "lineno": 42,
              "colno": 15,
              "context_line": "const value = obj.x.y;"
            }
          ]
        }
      }
    ]
  },
  "request": {
    "url": "https://example.com/dashboard",
    "method": "GET"
  },
  "user": {
    "id": "user123",
    "email": "user@example.com"
  },
  "tags": {
    "browser": "Chrome",
    "os": "Windows"
  },
  "environment": "production",
  "release": "v1.0.0"
}
```

**Response (200):**
```json
{
  "id": "event_uuid"
}
```

## 🧪 Testing Strategy

### Backend Tests

1. **Unit Tests**: Test individual functions and methods
2. **Integration Tests**: Test API endpoints and database interactions
3. **Service Tests**: Test business logic

```bash
# Run tests
cd backend
go test ./...

# Run with coverage
go test -cover ./...

# Run specific test
go test -v ./internal/services -run TestUserService
```

### Frontend Tests

1. **Component Tests**: Test individual React components
2. **Hook Tests**: Test custom hooks
3. **Integration Tests**: Test user flows

```bash
# Run tests
cd frontend
npm test

# Run with coverage
npm run test:coverage

# Run specific test
npm test -- UserForm.test.tsx
```

### End-to-End Tests

The integration test suite (`test-integration.js`) covers:
- Complete user workflow
- API endpoint testing
- Error ingestion and processing
- Issue management

```bash
# Run integration tests
npm install axios
node test-integration.js
```

### Test Data

Use the test client (`examples/js-client.html`) for manual testing:
1. Open in browser
2. Configure with your project DSN
3. Trigger various error types
4. Verify in dashboard

## 🤝 Contributing Guidelines

### Code Style

**Go:**
- Use `gofmt` for formatting
- Follow standard Go conventions
- Use meaningful variable names
- Write tests for new functionality

**TypeScript/React:**
- Use Prettier for formatting
- Follow React/TypeScript best practices
- Use functional components with hooks
- Write component tests

### Commit Messages

Follow conventional commits:
```
feat: add user authentication
fix: resolve database connection issue
docs: update API documentation
test: add integration tests for projects
```

### Pull Request Process

1. **Fork** the repository
2. **Create** a feature branch: `git checkout -b feature/new-feature`
3. **Make** your changes
4. **Test** your changes: `make test && node test-integration.js`
5. **Commit** with descriptive messages
6. **Push** to your fork: `git push origin feature/new-feature`
7. **Create** a Pull Request

### Code Review Checklist

- [ ] Code follows style guidelines
- [ ] Tests are included and passing
- [ ] Documentation is updated
- [ ] No breaking changes (or properly documented)
- [ ] Security considerations addressed
- [ ] Performance impact considered

## ⚡ Performance Considerations

### Database Performance

1. **Indexing**: Ensure proper indexes for common queries
2. **Connection Pooling**: Limit database connections
3. **Query Optimization**: Use EXPLAIN ANALYZE for slow queries
4. **Pagination**: Use cursor-based pagination for large datasets

### Backend Performance

1. **Caching**: Use Redis for frequently accessed data
2. **Async Processing**: Use worker queues for heavy operations
3. **Rate Limiting**: Protect against abuse
4. **Connection Reuse**: Use HTTP keep-alive

### Frontend Performance

1. **Code Splitting**: Split bundles by route
2. **Lazy Loading**: Load components on demand
3. **Memoization**: Use React.memo for expensive components
4. **Query Optimization**: Use TanStack Query effectively

### Monitoring

1. **Metrics**: Track API response times, error rates
2. **Logging**: Structured logging with appropriate levels
3. **Health Checks**: Monitor service health
4. **Resource Usage**: Monitor CPU, memory, disk usage

## 🚀 Deployment

### Development Deployment

```bash
# Start development environment
make dev

# View logs
make logs

# Stop services
make down
```

### Production Deployment

See [README.md](README.md#production-deployment) for production deployment instructions.

### Environment Variables

Production environment variables:
- Set secure `JWT_SECRET`
- Configure production database URL
- Set up proper CORS origins
- Configure email service (if implemented)
- Set up monitoring and logging

### Security Checklist

- [ ] JWT secret is securely generated
- [ ] Database credentials are secure
- [ ] HTTPS is enabled in production
- [ ] CORS is properly configured
- [ ] Rate limiting is enabled
- [ ] Input validation is comprehensive
- [ ] SQL injection protection is in place
- [ ] XSS protection is enabled

## 📚 Additional Resources

- [Go Documentation](https://golang.org/doc/)
- [React Documentation](https://react.dev/)
- [TanStack Query](https://tanstack.com/query)
- [TanStack Router](https://tanstack.com/router)
- [PostgreSQL Documentation](https://www.postgresql.org/docs/)
- [Docker Documentation](https://docs.docker.com/)

---

For questions or support, please create an issue or refer to the main [README.md](README.md) for troubleshooting guidance.