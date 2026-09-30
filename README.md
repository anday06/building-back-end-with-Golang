# Task Management API

REST API for Task Management built with Go, Gin, PostgreSQL, Redis, and Docker.

## Features

- **Authentication**: JWT-based auth with register/login, bcrypt password hashing
- **Authorization**: Role-based access control (admin/user), protected routes
- **Resources**: Projects, Tasks, Comments with full CRUD
- **Real-time**: WebSocket notifications for comments
- **Caching**: Redis caching for task lists with auto-invalidation
- **Rate Limiting**: Token bucket per IP
- **Observability**: Structured logging, Prometheus metrics, health check
- **Graceful Shutdown**: Signal handling with 10s timeout
- **Background Jobs**: Worker pool for async notifications
- **Soft Delete**: GORM soft delete on all resources
- **Pagination/Filtering/Sorting**: On all list endpoints
- **API Docs**: Swagger UI at `/swagger/index.html`
- **CI/CD**: GitHub Actions (fmt, vet, test, build)

## Tech Stack

| Layer | Technology |
|-------|------------|
| Language | Go 1.23 |
| Framework | Gin |
| Database | PostgreSQL 16 + GORM |
| Cache | Redis 7 + go-redis |
| Auth | JWT (golang-jwt/jwt/v5) + bcrypt |
| Real-time | Gorilla WebSocket |
| Docs | Swagger (swaggo) |
| Testing | testify, httptest |
| CI/CD | GitHub Actions |
| Container | Docker multi-stage, docker-compose |

## Project Structure

```
.
├── cmd/server              # Application entry point
├── internal/
│   ├── config              # Configuration loading
│   ├── database            # DB connection & migration
│   ├── handler             # HTTP handlers (Gin)
│   ├── middleware          # Auth, logging, rate limit, CORS, metrics
│   ├── models              # GORM models
│   ├── repository          # Data access layer
│   ├── service             # Business logic
│   ├── job                 # Background worker pool
│   └── realtime            # WebSocket hub
├── pkg/
│   ├── password            # Bcrypt helpers
│   ├── response            # Standard JSON responses
│   └── token               # JWT create/parse
├── docs/                   # Swagger generated files
├── docker-compose.yml      # Local dev stack
├── Dockerfile              # Multi-stage build
├── Makefile                # Common commands
└── .github/workflows/ci.yml
```

## Quick Start

### Prerequisites
- Docker Desktop
- Go 1.23+ (for local dev without Docker)

### Using Docker Compose (Recommended)

```bash
# Clone repo
git clone https://github.com/anday06/building-back-end-with-Golang.git
cd building-back-end-with-Golang

# Start all services (app, postgres, redis)
docker compose up --build

# API available at http://localhost:8080
# Swagger UI at http://localhost:8080/swagger/index.html
```

### Local Development (without Docker)

```bash
# Start PostgreSQL & Redis separately
docker run -d --name postgres -e POSTGRES_PASSWORD=12345 -e POSTGRES_DB=tasks -p 5432:5432 postgres:16-alpine
docker run -d --name redis -p 6379:6379 redis:7-alpine

# Copy env and adjust if needed
cp .env.example .env

# Run migrations & start server
go run ./cmd/server
```

## Environment Variables

| Variable | Default | Description |
|----------|---------|-------------|
| `PORT` | `8080` | HTTP port |
| `DATABASE_URL` | `host=localhost user=postgres password=postgres dbname=tasks port=5432 sslmode=disable` | Postgres DSN |
| `REDIS_URL` | `localhost:6379` | Redis address |
| `JWT_SECRET` | `development-secret` | JWT signing secret |
| `JWT_EXPIRES_HOURS` | `24` | Token expiry |

## API Endpoints

### Public
| Method | Path | Description |
|--------|------|-------------|
| GET | `/health` | Health check |
| GET | `/metrics` | Prometheus metrics |
| GET | `/swagger/*` | Swagger UI |
| POST | `/api/v1/auth/register` | Register |
| POST | `/api/v1/auth/login` | Login |

### Protected (require `Authorization: Bearer <token>`)
| Method | Path | Description |
|--------|------|-------------|
| GET | `/api/v1/users/me` | Get current user |
| PUT | `/api/v1/users/me` | Update current user |
| DELETE | `/api/v1/users/me` | Delete current user |
| GET | `/api/v1/projects` | List projects |
| POST | `/api/v1/projects` | Create project |
| GET | `/api/v1/projects/:id` | Get project |
| PUT | `/api/v1/projects/:id` | Update project |
| DELETE | `/api/v1/projects/:id` | Delete project |
| GET | `/api/v1/tasks` | List tasks (filter: `?status=todo`) |
| POST | `/api/v1/tasks` | Create task |
| GET | `/api/v1/tasks/:id` | Get task |
| PUT | `/api/v1/tasks/:id` | Update task |
| DELETE | `/api/v1/tasks/:id` | Delete task |
| GET | `/api/v1/tasks/:id/comments` | List task comments |
| POST | `/api/v1/comments` | Create comment |
| GET | `/api/v1/comments/:id` | Get comment |
| PUT | `/api/v1/comments/:id` | Update comment (author only) |
| DELETE | `/api/v1/comments/:id` | Delete comment (author only) |
| GET | `/api/v1/admin/status` | Admin only |
| GET | `/ws?token=<jwt>` | WebSocket realtime |

### Query Parameters (List endpoints)
- `page` (default: 1)
- `limit` (default: 20, max: 100)
- `sort` (default: `created_at DESC`, use `oldest` for ASC)
- `status` (tasks only): `todo`, `in_progress`, `done`

## Example Usage

```bash
# Register
curl -X POST http://localhost:8080/api/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{"name":"John","email":"john@example.com","password":"secret123"}'

# Login
curl -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"john@example.com","password":"secret123"}'

# Use token
TOKEN="eyJhbGciOiJIUzI1NiIs..."

# Create project
curl -X POST http://localhost:8080/api/v1/projects \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"name":"My Project","description":"Project description"}'

# List tasks with pagination
curl -G http://localhost:8080/api/v1/tasks \
  -H "Authorization: Bearer $TOKEN" \
  -d page=1 -d limit=10 -d status=todo
```

## Testing

```bash
# Run all tests
go test ./...

# Run with coverage
go test -cover ./...

# Specific package
go test ./internal/handler -v
```

## Docker

```bash
# Build image
docker build -t task-api .

# Run container (needs external DB/Redis)
docker run -p 8080:8080 \
  -e DATABASE_URL="host=host.docker.internal user=postgres password=12345 dbname=tasks port=5432 sslmode=disable" \
  -e REDIS_URL="host.docker.internal:6379" \
  -e JWT_SECRET="prod-secret" \
  task-api
```

## Deploy to Render (Free)

1. Push repo to GitHub
2. Create Render account → New Web Service → Connect GitHub repo
3. Environment: **Docker**
4. Add Environment Variables:
   - `DATABASE_URL` (from Render PostgreSQL)
   - `REDIS_URL` (from Render Redis)
   - `JWT_SECRET` (generate strong secret)
   - `JWT_EXPIRES_HOURS=24`
5. Deploy → Get `https://your-app.onrender.com`

## CI/CD Pipeline

GitHub Actions (`.github/workflows/ci.yml`) runs on every push/PR:
- `gofmt` check
- `go vet`
- `go test ./...`
- `go build ./cmd/server`

## Swagger Documentation

After running locally or deployed:
- **Local**: http://localhost:8080/swagger/index.html
- **Production**: `https://your-domain/swagger/index.html`

## License

MIT