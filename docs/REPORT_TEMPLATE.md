# BÁO CÁO BÀI TẬP CÁ NHÂN: BUILDING BACK-END WITH GOLANG

**Đề tài:** Hệ thống quản lý công việc (Task Management API)

**Họ tên:** [Họ tên của bạn]
**Mã số học viên:** [MSSV của bạn]
**Lớp:** [Lớp của bạn]
**Ngày nộp:** [Ngày nộp]

---

## 1. MỤC TIÊU VÀ PHẠM VI BÀI LÀM

### 1.1 Mục tiêu
Xây dựng REST API hoàn chỉnh cho hệ thống quản lý công việc (Task Management) với các tính năng:
- Quản lý User, Project, Task, Comment
- Authentication & Authorization (JWT, Role-based)
- Real-time notifications (WebSocket)
- Caching, Rate limiting, Observability
- Containerization & CI/CD

### 1.2 Phạm vi
- Backend API (không bao gồm frontend)
- Database: PostgreSQL
- Cache: Redis
- Deploy: Docker + Render/Fly.io
- Documentation: Swagger + Postman

---

## 2. MÔ TẢ HỆ THỐNG

### 2.1 Đề tài chọn
**Hệ thống quản lý công việc (Task Management)** - Cho phép người dùng tạo project, chia thành các task, comment và nhận thông báo realtime.

### 2.2 Tính năng chính
| Module | Tính năng |
|--------|-----------|
| Auth | Register, Login, JWT token, Bcrypt password |
| User | Profile CRUD, Role (user/admin) |
| Project | CRUD, owner-based authorization |
| Task | CRUD, status/priority, filter, pagination, Redis cache |
| Comment | CRUD, author-only update/delete, background notification |
| Real-time | WebSocket broadcast comment notifications |
| Admin | Protected admin endpoint |

### 2.3 Công nghệ sử dụng
| Category | Technology |
|----------|------------|
| Language | Go 1.23 |
| Framework | Gin |
| Database | PostgreSQL 16 + GORM |
| Cache | Redis 7 + go-redis/v9 |
| Auth | JWT (golang-jwt/v5), Bcrypt |
| Real-time | Gorilla WebSocket |
| Testing | testify, httptest |
| Docs | Swagger (swaggo), Postman |
| CI/CD | GitHub Actions |
| Container | Docker multi-stage, docker-compose |

---

## 3. KIẾN TRÚC ỨNG DỤNG

### 3.1 Sơ đồ Layer Architecture
```
┌─────────────────────────────────────────────────────────────┐
│                      HTTP Layer (Gin)                       │
│  ┌─────────────┐ ┌─────────────┐ ┌─────────────┐            │
│  │   Auth      │ │  Projects   │ │   Tasks     │  ...       │
│  │  Handlers   │ │  Handlers   │ │  Handlers   │            │
│  └──────┬──────┘ └──────┬──────┘ └──────┬──────┘            │
└─────────┼───────────────┼───────────────┼────────────────────┘
          ▼               ▼               ▼
┌─────────────────────────────────────────────────────────────┐
│                    Service Layer                            │
│  ┌─────────────┐ ┌─────────────┐ ┌─────────────┐            │
│  │ AuthService │ │UserService  │ │TaskService  │  ...       │
│  └──────┬──────┘ └──────┬──────┘ └──────┬──────┘            │
└─────────┼───────────────┼───────────────┼────────────────────┘
          ▼               ▼               ▼
┌─────────────────────────────────────────────────────────────┐
│                   Repository Layer                          │
│  ┌─────────────┐ ┌─────────────┐ ┌─────────────┐            │
│  │UserRepository│ │ProjectRepo  │ │TaskRepository│ (Cache)  │
│  └──────┬──────┘ └──────┬──────┘ └──────┬──────┘            │
└─────────┼───────────────┼───────────────┼────────────────────┘
          ▼               ▼               ▼
┌─────────────────────────────────────────────────────────────┐
│                      Data Layer                             │
│         PostgreSQL (GORM)         Redis (Cache)             │
└─────────────────────────────────────────────────────────────┘
```

### 3.2 Luồng Request/Response
1. Client gửi HTTP request → Gin Router
2. Middleware chain: Recovery → Logger → Metrics → RateLimit → CORS → Auth (nếu protected)
3. Handler bind & validate input
4. Handler gọi Service layer (business logic)
5. Service gọi Repository (data access)
6. Repository query PostgreSQL / Redis
7. Response trả về qua Handler → JSON chuẩn `{success, data/error}`

### 3.3 Middleware Pipeline
```
Request
  │
  ├─► Recovery (panic handling)
  ├─► Logger (method, path, status, latency)
  ├─► Metrics (Prometheus counters/histograms)
  ├─► RateLimit (token bucket 10req/s per IP)
  ├─► CORS (allow all origins)
  ├─► Auth (JWT validation → set user_id, role)
  └─► Handler
```

---

## 4. THIẾT KẾ DATABASE

### 4.1 ERD (Entity Relationship Diagram)
```
┌─────────────┐       ┌─────────────┐       ┌─────────────┐       ┌─────────────┐
│    User     │       │   Project   │       │    Task     │       │  Comment    │
├─────────────┤       ├─────────────┤       ├─────────────┤       ├─────────────┤
│ id (PK)     │◄──────│ id (PK)     │◄──────│ id (PK)     │◄──────│ id (PK)     │
│ name        │ 1:N   │ name        │ 1:N   │ title       │ 1:N   │ body        │
│ email (UK)  │       │ description │       │ description │       │ task_id(FK) │
│ password_hash│       │ owner_id(FK)│       │ status      │       │ author_idFK)│
│ role        │       │ created_at  │       │ priority    │       │ created_at  │
│ created_at  │       │ updated_at  │       │ project_idFK)│      │ deleted_at  │
│ updated_at  │       │ deleted_at  │       │ assignee_id │       └─────────────┘
│ deleted_at  │       └─────────────┘       │ created_at  │
└─────────────┘                             │ updated_at  │
                                            │ deleted_at  │
                                            └─────────────┘
```

### 4.2 Bảng chính & Mối quan hệ
| Bảng | Mô tả | Quan hệ |
|------|-------|---------|
| `users` | Người dùng hệ thống | 1:N Project (owner), 1:N Comment (author) |
| `projects` | Dự án chứa các task | N:1 User (owner), 1:N Task |
| `tasks` | Công việc trong project | N:1 Project, 1:N Comment |
| `comments` | Bình luận trên task | N:1 Task, N:1 User (author) |

### 4.3 Index & Constraints
- `users.email`: UNIQUE INDEX
- `projects.owner_id`: INDEX (FK)
- `tasks.project_id`: INDEX (FK)
- `tasks.status`: INDEX (filter)
- `comments.task_id`: INDEX (FK)
- `comments.author_id`: INDEX (FK)
- Soft delete: `deleted_at` trên tất cả bảng

---

## 5. MÔ TẢ CHI TIẾT CÁC ENDPOINT

### 5.1 Authentication (Public)
| Method | Path | Request Body | Response | Status |
|--------|------|--------------|----------|--------|
| POST | `/api/v1/auth/register` | `{name, email, password}` | User | 201 |
| POST | `/api/v1/auth/login` | `{email, password}` | `{token, user}` | 200 |

### 5.2 User Profile (Protected)
| Method | Path | Request Body | Response | Status |
|--------|------|--------------|----------|--------|
| GET | `/api/v1/users/me` | - | User | 200 |
| PUT | `/api/v1/users/me` | `{name, email, password?}` | User | 200 |
| DELETE | `/api/v1/users/me` | - | - | 204 |

### 5.3 Projects (Protected, Owner-only)
| Method | Path | Query/Body | Response | Status |
|--------|------|------------|----------|--------|
| GET | `/api/v1/projects` | `page, limit, sort` | Project[] | 200 |
| POST | `/api/v1/projects` | `{name, description}` | Project | 201 |
| GET | `/api/v1/projects/:id` | - | Project | 200 |
| PUT | `/api/v1/projects/:id` | `{name, description}` | Project | 200 |
| DELETE | `/api/v1/projects/:id` | - | - | 204 |

### 5.4 Tasks (Protected, Owner-only via Project)
| Method | Path | Query/Body | Response | Status |
|--------|------|------------|----------|--------|
| GET | `/api/v1/tasks` | `page, limit, sort, status` | Task[] | 200 |
| POST | `/api/v1/tasks` | `{title, description, status, priority, project_id}` | Task | 201 |
| GET | `/api/v1/tasks/:id` | - | Task | 200 |
| PUT | `/api/v1/tasks/:id` | `{title, description, status, priority}` | Task | 200 |
| DELETE | `/api/v1/tasks/:id` | - | - | 204 |

### 5.5 Comments (Protected, Author-only for update/delete)
| Method | Path | Query/Body | Response | Status |
|--------|------|------------|----------|--------|
| GET | `/api/v1/tasks/:id/comments` | `page, limit, sort` | Comment[] | 200 |
| POST | `/api/v1/comments` | `{body, task_id}` | Comment | 201 |
| GET | `/api/v1/comments/:id` | - | Comment | 200 |
| PUT | `/api/v1/comments/:id` | `{body}` | Comment | 200 |
| DELETE | `/api/v1/comments/:id` | - | - | 204 |

### 5.6 Admin & System
| Method | Path | Auth | Response | Status |
|--------|------|------|----------|--------|
| GET | `/health` | No | `{status: ok}` | 200 |
| GET | `/metrics` | No | Prometheus metrics | 200 |
| GET | `/swagger/*` | No | Swagger UI | 200 |
| GET | `/api/v1/admin/status` | Admin | `{message}` | 200 |
| GET | `/ws?token=` | JWT query | WebSocket | 101 |

### 5.7 Response Format
**Success:**
```json
{
  "success": true,
  "data": { ... }
}
```
**Error:**
```json
{
  "success": false,
  "error": { "message": "error description" }
}
```

---

## 6. AUTHENTICATION & MIDDLEWARE

### 6.1 JWT Flow
```
1. Client POST /auth/login {email, password}
2. Server: verify bcrypt → create JWT (HS256)
   Payload: {user_id, role, exp, iat}
3. Return: {token, user}
4. Client: Authorization: Bearer <token>
5. Middleware Auth:
   - Parse header "Bearer "
   - Validate signature & expiry
   - Set context: user_id, role
6. Handler uses context values
```

### 6.2 Middleware Chi tiết
| Middleware | Chức năng | File |
|------------|-----------|------|
| `Recovery` | Catch panic, return 500 | Gin built-in |
| `Logger` | Log method, path, status, latency | `internal/middleware/middleware.go` |
| `Metrics` | Prometheus: request_count, request_duration | `internal/middleware/metrics.go` |
| `RateLimit` | Token bucket 10req/s, burst 20 per IP | `internal/middleware/middleware.go` |
| `CORS` | Allow all origins, methods, headers | `gin-contrib/cors` |
| `Auth` | JWT validate, set user_id/role | `internal/middleware/middleware.go` |
| `AdminOnly` | Check role == "admin" | `internal/middleware/middleware.go` |

### 6.3 Password Security
- Bcrypt cost: DefaultCost (10)
- Không lưu plain text
- Hash khi register & update password

---

## 7. TESTING

### 7.1 Unit Test Coverage
| Package | Test Cases | Coverage |
|---------|------------|----------|
| `internal/handler` | 3 (Health, PathID, Bind) | ~15% |
| `internal/middleware` | 6 (Auth, AdminOnly, CORS) | ~35% |
| `internal/service` | 1 (Password hash/compare) | ~10% |
| **Tổng** | **10 test cases** | **~20%** |

### 7.2 Chạy test
```bash
go test ./... -v
go test ./... -cover
```

### 7.3 Test Strategy
- Handler: httptest + Gin TestMode
- Middleware: httptest + token helper
- Service: Password package isolated test
- Integration: testcontainers (chưa implement)

---

## 8. DOCKER & DEPLOY

### 8.1 Dockerfile (Multi-stage)
```dockerfile
# Stage 1: Builder
FROM golang:1.23-alpine AS builder
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags='-s -w' -o /app/server ./cmd/server

# Stage 2: Runtime
FROM alpine:3.20
RUN adduser -D -H appuser
USER appuser
COPY --from=builder /app/server /app/server
EXPOSE 8080
ENTRYPOINT ["/app/server"]
```

### 8.2 docker-compose.yml
```yaml
services:
  app:
    build: .
    ports: ["8080:8080"]
    env_file: .env
    depends_on:
      db: {condition: service_healthy}
      redis: {condition: service_started}
  db:
    image: postgres:16-alpine
    environment:
      POSTGRES_DB: tasks
      POSTGRES_USER: postgres
      POSTGRES_PASSWORD: 12345
    healthcheck: pg_isready
  redis:
    image: redis:7-alpine
```

### 8.3 CI/CD Pipeline (GitHub Actions)
```yaml
# .github/workflows/ci.yml
jobs:
  quality:
    runs-on: ubuntu-latest
    steps:
      - checkout
      - setup-go (1.23)
      - gofmt check
      - go vet
      - go test
      - go build
```

### 8.4 Deploy Render
1. Connect GitHub repo → New Web Service → Docker
2. Add Managed PostgreSQL + Redis
3. Set Environment Variables:
   - `DATABASE_URL` (from Render PG)
   - `REDIS_URL` (from Render Redis)
   - `JWT_SECRET` (strong random string)
   - `JWT_EXPIRES_HOURS=24`
4. Deploy → Auto-build & run

**Link deploy:** `https://<your-app>.onrender.com`

---

## 9. KHÓ KHĂN GẶP PHẢI VÀ CÁCH GIẢI QUYẾT

| Khó khăn | Giải pháp |
|----------|-----------|
| Docker daemon không chạy trên Windows | Khởi động Docker Desktop, chờ daemon ready |
| Swagger import package `docs` không tìm thấy | Sửa `.dockerignore` bỏ qua `docs`, copy vào build context |
| Rate limit in-memory không scale multi-instance | Acceptable cho bài tập; production dùng Redis-based |
| GORM Preload("Tasks") gây N+1 | Chỉ load khi cần; có thể dùng Select field hoặc lazy load |
| WebSocket token qua query param | Thay vì header (browser limitation), validate trong Hub.Handle |
| Background worker drop notification khi full | Log warning, có thể tăng buffer hoặc dùng persistent queue |

---

## 10. LESSONS LEARNED (BÀI HỌC RÚT RA)

1. **Layered Architecture** giúp tách biệt concerns, dễ test, maintain
2. **Middleware pattern** trong Gin mạnh mẽ cho cross-cutting concerns
3. **GORM AutoMigrate** tiện cho dev, nhưng production nên dùng migration tool (golang-migrate)
4. **Redis caching** đơn giản nhưng hiệu quả: invalidate trên write path
5. **Docker multi-stage** giảm size image đáng kể (alpine ~20MB vs golang ~1GB)
6. **Graceful shutdown** quan trọng cho production: drain connections, finish jobs
7. **Swagger annotations** verbose nhưng tự động generate docs chuẩn OpenAPI 3.0
8. **GitHub Actions** miễn phí cho public repo, đủ cho CI cơ bản
9. **Environment variables** cho config: không hardcode, không commit .env
10. **Test-driven mindset** từ đầu giúp code ít bug hơn, dễ refactor

---

## 11. LINK THAM CHIẾU

| Resource | Link |
|----------|------|
| **GitHub Repository** | https://github.com/anday06/building-back-end-with-Golang |
| **API Deploy (Render)** | https://<your-app>.onrender.com |
| **Swagger UI (Local)** | http://localhost:8080/swagger/index.html |
| **Swagger UI (Deploy)** | https://<your-app>.onrender.com/swagger/index.html |
| **Postman Collection** | `docs/postman_collection.json` (import vào Postman) |
| **Video Demo (YouTube)** | [Link video unlisted] |

---

## PHỤ LỤC: POSTMAN COLLECTION IMPORT

1. Mở Postman → Import → File → Chọn `docs/postman_collection.json`
2. Set environment variable `baseUrl` = `http://localhost:8080` (local) hoặc deploy URL
3. Chạy theo thứ tự: Register → Login (auto save token) → CRUD Project → CRUD Task → Comment → WebSocket

---

*Kết thúc báo cáo*