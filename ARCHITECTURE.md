# MiniSentry System Architecture

## 1. System Overview

MiniSentry is a comprehensive error tracking and monitoring system designed as a Sentry clone for JavaScript web applications. The system follows a modern three-tier architecture with clear separation of concerns.

### High-Level Architecture

```
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   Frontend      │    │   Backend       │    │   Database      │
│   (React SPA)   │◄──►│   (Go API)      │◄──►│  (PostgreSQL)   │
│                 │    │                 │    │                 │
│ • React 18      │    │ • Chi Router    │    │ • Primary DB    │
│ • TanStack      │    │ • JWT Auth      │    │ • Redis Cache   │
│ • Tailwind CSS  │    │ • Rate Limiting │    │ • Indexes       │
└─────────────────┘    └─────────────────┘    └─────────────────┘
        │                       │                       │
        └───────────────────────┼───────────────────────┘
                               │
                    ┌─────────────────┐
                    │   External      │
                    │   Services      │
                    │                 │
                    │ • Email Service │
                    │ • File Storage  │
                    │ • Monitoring    │
                    └─────────────────┘
```

### Component Relationships

- **Frontend**: Single Page Application serving the dashboard and error management interface
- **Backend**: RESTful API server handling authentication, error processing, and data management
- **Database**: PostgreSQL for persistent data storage with Redis for caching and sessions
- **External Services**: Email notifications, file storage for source maps, and system monitoring

## 2. Technology Stack

### Backend (Go)
- **Framework**: Chi Router v5 - Lightweight, fast, and composable
- **Database ORM**: GORM v2 - Feature-rich ORM with excellent PostgreSQL support
- **Authentication**: JWT with RS256 signing
- **Caching**: Redis for session management and hot data
- **Configuration**: Viper for environment-based configuration
- **Logging**: Structured logging with logrus
- **Testing**: Standard Go testing with testify assertions

**Justification**: Go provides excellent performance for high-throughput error ingestion, strong concurrency support, and fast compilation for development cycles.

### Frontend (React)
- **Framework**: React 18 with concurrent features
- **Routing**: TanStack Router v1 - Type-safe routing with excellent developer experience
- **State Management**: TanStack Query v5 for server state + Zustand for client state
- **UI Framework**: Tailwind CSS v3 with Headless UI components
- **Build Tool**: Vite for fast development and optimized builds
- **Testing**: Vitest + React Testing Library

**Justification**: This stack provides excellent developer experience, type safety, and performance while maintaining simplicity and reducing bundle size.

### Database
- **Primary**: PostgreSQL 15+ with JSONB support for flexible error metadata
- **Cache**: Redis 7+ for session storage and query caching
- **Search**: PostgreSQL full-text search with GIN indexes

**Justification**: PostgreSQL offers excellent JSON support for error metadata, strong consistency, and powerful indexing capabilities for error grouping and search.

## 3. Database Schema

### Core Tables

```sql
-- Users and Authentication
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

-- Organizations/Teams
CREATE TABLE organizations (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(255) NOT NULL,
    slug VARCHAR(100) UNIQUE NOT NULL,
    description TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Organization membership
CREATE TABLE organization_members (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    role VARCHAR(50) NOT NULL DEFAULT 'member', -- owner, admin, member
    joined_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    UNIQUE(organization_id, user_id)
);

-- Projects
CREATE TABLE projects (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    name VARCHAR(255) NOT NULL,
    slug VARCHAR(100) NOT NULL,
    description TEXT,
    platform VARCHAR(50) NOT NULL DEFAULT 'javascript',
    dsn VARCHAR(255) UNIQUE NOT NULL, -- Data Source Name for SDK
    public_key VARCHAR(255) NOT NULL,
    secret_key VARCHAR(255) NOT NULL,
    is_active BOOLEAN DEFAULT true,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    UNIQUE(organization_id, slug)
);

-- Error Issues (grouped errors)
CREATE TABLE issues (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    project_id UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    fingerprint VARCHAR(255) NOT NULL, -- Used for grouping
    title VARCHAR(500) NOT NULL,
    culprit VARCHAR(500), -- Function/file where error occurred
    type VARCHAR(100) NOT NULL, -- error, csp, default
    level VARCHAR(50) NOT NULL DEFAULT 'error', -- debug, info, warning, error, fatal
    status VARCHAR(50) DEFAULT 'unresolved', -- resolved, unresolved, ignored
    first_seen TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    last_seen TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    times_seen INTEGER DEFAULT 1,
    assignee_id UUID REFERENCES users(id),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    UNIQUE(project_id, fingerprint)
);

-- Individual error events
CREATE TABLE events (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    issue_id UUID NOT NULL REFERENCES issues(id) ON DELETE CASCADE,
    project_id UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    event_id VARCHAR(255) NOT NULL, -- Unique identifier from SDK
    timestamp TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    level VARCHAR(50) NOT NULL DEFAULT 'error',
    message TEXT,
    exception_type VARCHAR(255),
    exception_value TEXT,
    stack_trace JSONB,
    request_data JSONB, -- URL, method, headers, etc.
    user_context JSONB, -- User info, if available
    tags JSONB, -- Key-value pairs for filtering
    extra_data JSONB, -- Additional context
    fingerprint VARCHAR(255) NOT NULL,
    release_version VARCHAR(100),
    environment VARCHAR(100) DEFAULT 'production',
    server_name VARCHAR(255),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    UNIQUE(project_id, event_id)
);

-- Comments on issues
CREATE TABLE issue_comments (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    issue_id UUID NOT NULL REFERENCES issues(id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    content TEXT NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Issue activity log
CREATE TABLE issue_activities (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    issue_id UUID NOT NULL REFERENCES issues(id) ON DELETE CASCADE,
    user_id UUID REFERENCES users(id),
    type VARCHAR(100) NOT NULL, -- status_change, assignment, comment, etc.
    data JSONB,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Releases and deployments
CREATE TABLE releases (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    project_id UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    version VARCHAR(100) NOT NULL,
    ref VARCHAR(255), -- Git commit SHA
    url VARCHAR(500), -- Repository URL
    date_created TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    date_released TIMESTAMP WITH TIME ZONE,
    UNIQUE(project_id, version)
);

-- Performance indexes
CREATE INDEX idx_events_issue_id ON events(issue_id);
CREATE INDEX idx_events_project_timestamp ON events(project_id, timestamp DESC);
CREATE INDEX idx_events_fingerprint ON events(fingerprint);
CREATE INDEX idx_events_level ON events(level);
CREATE INDEX idx_events_environment ON events(environment);
CREATE INDEX idx_issues_project_status ON issues(project_id, status);
CREATE INDEX idx_issues_last_seen ON issues(last_seen DESC);
CREATE INDEX idx_issues_times_seen ON issues(times_seen DESC);

-- Full-text search indexes
CREATE INDEX idx_events_message_fts ON events USING GIN(to_tsvector('english', message));
CREATE INDEX idx_issues_title_fts ON issues USING GIN(to_tsvector('english', title));

-- JSONB indexes for fast queries
CREATE INDEX idx_events_tags ON events USING GIN(tags);
CREATE INDEX idx_events_user_context ON events USING GIN(user_context);
CREATE INDEX idx_events_stack_trace ON events USING GIN(stack_trace);
```

### Key Design Decisions

1. **UUID Primary Keys**: Better for distributed systems and API security
2. **JSONB Storage**: Flexible schema for error metadata while maintaining query performance
3. **Fingerprinting**: Errors are grouped by fingerprint (hash of stack trace + error type)
4. **Soft Deletes**: Issues and projects maintain history through status flags
5. **Time-based Partitioning**: Events table can be partitioned by month for better performance

## 4. API Design

### Authentication Endpoints

```
POST   /api/auth/login
POST   /api/auth/register
POST   /api/auth/logout
POST   /api/auth/refresh
POST   /api/auth/forgot-password
POST   /api/auth/reset-password
GET    /api/auth/me
```

### Organizations

```
GET    /api/organizations                    # List user's organizations
POST   /api/organizations                    # Create organization
GET    /api/organizations/{orgSlug}          # Get organization details
PUT    /api/organizations/{orgSlug}          # Update organization
DELETE /api/organizations/{orgSlug}          # Delete organization
GET    /api/organizations/{orgSlug}/members  # List members
POST   /api/organizations/{orgSlug}/members  # Invite member
PUT    /api/organizations/{orgSlug}/members/{userId} # Update member role
DELETE /api/organizations/{orgSlug}/members/{userId} # Remove member
```

### Projects

```
GET    /api/organizations/{orgSlug}/projects           # List projects
POST   /api/organizations/{orgSlug}/projects           # Create project
GET    /api/organizations/{orgSlug}/projects/{projectSlug} # Get project
PUT    /api/organizations/{orgSlug}/projects/{projectSlug} # Update project
DELETE /api/organizations/{orgSlug}/projects/{projectSlug} # Delete project
POST   /api/organizations/{orgSlug}/projects/{projectSlug}/keys/regenerate # Regenerate DSN
```

### Issues

```
GET    /api/projects/{projectId}/issues              # List issues with filtering
GET    /api/projects/{projectId}/issues/{issueId}    # Get issue details
PUT    /api/projects/{projectId}/issues/{issueId}    # Update issue (resolve, assign, etc.)
DELETE /api/projects/{projectId}/issues/{issueId}    # Delete issue
GET    /api/projects/{projectId}/issues/{issueId}/events # List events for issue
GET    /api/projects/{projectId}/issues/{issueId}/comments # List comments
POST   /api/projects/{projectId}/issues/{issueId}/comments # Add comment
```

### Events (Error Ingestion)

```
POST   /api/projects/{projectId}/store/     # Store error event (from SDK)
GET    /api/projects/{projectId}/events     # List events
GET    /api/events/{eventId}                # Get event details
```

### Example API Specifications

#### POST /api/auth/login
```json
Request:
{
  "email": "user@example.com",
  "password": "securepassword"
}

Response (200):
{
  "token": "eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9...",
  "refresh_token": "refresh_token_here",
  "user": {
    "id": "uuid",
    "email": "user@example.com",
    "name": "John Doe",
    "avatar_url": "https://..."
  }
}
```

#### GET /api/projects/{projectId}/issues
```json
Query Parameters:
- status: resolved|unresolved|ignored
- level: debug|info|warning|error|fatal
- environment: production|staging|development
- search: "search query"
- sort: first_seen|last_seen|times_seen|priority
- limit: 25 (default)
- cursor: "pagination_cursor"

Response (200):
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

#### POST /api/projects/{projectId}/store/
```json
Request (from Sentry SDK):
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
              "context_line": "const value = obj.x.y;",
              "pre_context": ["function handleClick() {", "  const obj = getObject();"],
              "post_context": ["  return value;", "}"]
            }
          ]
        }
      }
    ]
  },
  "request": {
    "url": "https://example.com/dashboard",
    "method": "GET",
    "headers": {"User-Agent": "Mozilla/5.0..."}
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

Response (200):
{
  "id": "event_uuid"
}
```

## 5. Frontend Architecture

### Component Structure

```
src/
├── components/           # Reusable UI components
│   ├── ui/              # Base UI components (Button, Input, etc.)
│   ├── forms/           # Form components
│   ├── layout/          # Layout components (Header, Sidebar, etc.)
│   └── charts/          # Data visualization components
├── pages/               # Page components
│   ├── auth/           # Authentication pages
│   ├── dashboard/      # Dashboard and overview
│   ├── issues/         # Issue management
│   ├── projects/       # Project management
│   └── settings/       # Settings and configuration
├── hooks/              # Custom React hooks
├── lib/                # Utility functions and configurations
│   ├── api.ts          # API client setup
│   ├── auth.ts         # Authentication utilities
│   └── utils.ts        # General utilities
├── stores/             # Zustand stores for client state
├── types/              # TypeScript type definitions
└── routes/             # TanStack Router configuration
```

### Routing Structure (TanStack Router)

```typescript
// routes/index.ts
import { createRouter } from '@tanstack/react-router'
import { routeTree } from './routeTree.gen'

export const router = createRouter({
  routeTree,
  context: {
    auth: undefined!, // Will be set by auth provider
  },
})

// Route structure:
// /                           - Dashboard redirect
// /login                      - Login page
// /register                   - Registration page
// /organizations/{orgSlug}/
//   ├── projects/             - Projects list
//   ├── projects/{projectSlug}/
//   │   ├── issues/           - Issues list
//   │   ├── issues/{issueId}/ - Issue details
//   │   ├── events/           - Events list
//   │   └── settings/         - Project settings
//   └── settings/             - Organization settings
```

### State Management Strategy

1. **Server State**: TanStack Query for all API interactions
2. **Client State**: Zustand for UI state (modals, filters, etc.)
3. **Authentication**: Context + localStorage for token management
4. **Form State**: React Hook Form for complex forms

```typescript
// Example Zustand store
import { create } from 'zustand'

interface UIStore {
  sidebarCollapsed: boolean
  currentProject: string | null
  toggleSidebar: () => void
  setCurrentProject: (projectId: string) => void
}

export const useUIStore = create<UIStore>((set) => ({
  sidebarCollapsed: false,
  currentProject: null,
  toggleSidebar: () => set((state) => ({ 
    sidebarCollapsed: !state.sidebarCollapsed 
  })),
  setCurrentProject: (projectId) => set({ 
    currentProject: projectId 
  }),
}))
```

### Data Flow

1. **Authentication Flow**: 
   - Login → Store JWT → Set auth context → Redirect to dashboard
   - Auto-refresh tokens using TanStack Query mutations

2. **Issue Management Flow**:
   - Fetch issues with filters → Display in list → Select issue → Show details
   - Real-time updates using polling (future: WebSocket)

3. **Error Ingestion Flow**:
   - SDK sends POST request → Backend processes → Updates issue counts
   - Frontend polls for new data or receives notifications

## 6. Security Architecture

### Authentication & Authorization

1. **JWT Token Strategy**:
   - Access tokens (15 minutes expiry) for API authentication
   - Refresh tokens (7 days expiry) stored as httpOnly cookies
   - RS256 signing with rotating keys

2. **API Security**:
   - Rate limiting: 100 requests/minute per IP, 1000/minute per authenticated user
   - CORS configuration for specific domains
   - Request validation using middleware
   - SQL injection prevention through parameterized queries

3. **Project-level Security**:
   - DSN (Data Source Name) authentication for error ingestion
   - Project-scoped permissions (read, write, admin)
   - Organization-based access control

### Security Headers

```go
// Middleware for security headers
func SecurityHeaders(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        w.Header().Set("X-Content-Type-Options", "nosniff")
        w.Header().Set("X-Frame-Options", "DENY")
        w.Header().Set("X-XSS-Protection", "1; mode=block")
        w.Header().Set("Strict-Transport-Security", "max-age=31536000; includeSubDomains")
        w.Header().Set("Content-Security-Policy", "default-src 'self'")
        next.ServeHTTP(w, r)
    })
}
```

### Data Protection

1. **Sensitive Data Handling**:
   - Password hashing using bcrypt with cost 12
   - PII data encryption for user context
   - Secure DSN generation and storage

2. **Database Security**:
   - Connection encryption (SSL)
   - Row-level security for multi-tenancy
   - Regular security updates and patches

## 7. Performance Considerations

### Backend Performance

1. **Database Optimization**:
   - Strategic indexing for common queries
   - Connection pooling (max 20 connections)
   - Query optimization with EXPLAIN ANALYZE
   - Read replicas for analytics queries

2. **Caching Strategy**:
   - Redis for frequently accessed data (user sessions, project metadata)
   - Application-level caching for expensive computations
   - HTTP caching headers for static assets

3. **Error Processing**:
   - Asynchronous error processing using worker queues
   - Bulk insertions for high-throughput scenarios
   - Efficient fingerprinting algorithm for error grouping

### Frontend Performance

1. **Bundle Optimization**:
   - Code splitting by route using TanStack Router
   - Tree shaking for unused code elimination
   - Dynamic imports for heavy components

2. **Data Loading**:
   - TanStack Query for efficient data fetching and caching
   - Infinite queries for paginated lists
   - Optimistic updates for better UX

3. **Rendering Performance**:
   - React.memo for expensive components
   - Virtual scrolling for large lists
   - Debounced search inputs

### Performance Targets

- **API Response Time**: < 200ms for 95th percentile
- **Error Ingestion**: 100 errors/second sustained throughput
- **Frontend Loading**: < 2 seconds initial page load
- **Database Queries**: < 50ms for common operations

## 8. Deployment Architecture

### Container Configuration

```dockerfile
# Backend Dockerfile
FROM golang:1.21-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o main ./cmd/server

FROM alpine:latest
RUN apk --no-cache add ca-certificates tzdata
WORKDIR /root/
COPY --from=builder /app/main .
COPY --from=builder /app/migrations ./migrations
EXPOSE 8080
CMD ["./main"]
```

```dockerfile
# Frontend Dockerfile
FROM node:18-alpine AS builder
WORKDIR /app
COPY package*.json ./
RUN npm ci --only=production
COPY . .
RUN npm run build

FROM nginx:alpine
COPY --from=builder /app/dist /usr/share/nginx/html
COPY nginx.conf /etc/nginx/nginx.conf
EXPOSE 80
CMD ["nginx", "-g", "daemon off;"]
```

### Docker Compose Setup

```yaml
version: '3.8'
services:
  postgres:
    image: postgres:15-alpine
    environment:
      POSTGRES_DB: minisentry
      POSTGRES_USER: postgres
      POSTGRES_PASSWORD: password
    volumes:
      - postgres_data:/var/lib/postgresql/data
    ports:
      - "5432:5432"

  redis:
    image: redis:7-alpine
    ports:
      - "6379:6379"
    volumes:
      - redis_data:/data

  backend:
    build: ./backend
    environment:
      DATABASE_URL: postgres://postgres:password@postgres:5432/minisentry?sslmode=disable
      REDIS_URL: redis://redis:6379
      JWT_SECRET: your-secret-key
    ports:
      - "8080:8080"
    depends_on:
      - postgres
      - redis

  frontend:
    build: ./frontend
    ports:
      - "3000:80"
    depends_on:
      - backend

volumes:
  postgres_data:
  redis_data:
```

### Environment Configuration

```bash
# Backend Environment Variables
DATABASE_URL=postgres://user:pass@localhost:5432/minisentry
REDIS_URL=redis://localhost:6379
JWT_SECRET=your-256-bit-secret
JWT_ISSUER=minisentry
CORS_ORIGINS=http://localhost:3000,https://yourdomain.com
RATE_LIMIT_REQUESTS=100
RATE_LIMIT_WINDOW=60s
EMAIL_SMTP_HOST=smtp.example.com
EMAIL_SMTP_PORT=587
EMAIL_FROM=noreply@yourdomain.com
```

### Production Deployment

1. **Infrastructure**:
   - Load balancer (nginx/HAProxy) for high availability
   - Database clustering with primary/replica setup
   - Redis cluster for session storage
   - CDN for static asset delivery

2. **Monitoring**:
   - Application metrics (Prometheus + Grafana)
   - Error tracking (dogfooding with MiniSentry itself)
   - Log aggregation (ELK stack or similar)
   - Health checks and alerting

3. **Scaling Strategy**:
   - Horizontal scaling of API servers
   - Database read replicas for analytics
   - Redis cluster for session distribution
   - CDN for global performance

This architecture provides a solid foundation for a production-ready error tracking system that can scale with your needs while maintaining excellent performance and security standards.