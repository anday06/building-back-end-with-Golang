# Task Management API

REST API cho hệ thống quản lý công việc, xây dựng bằng Go 1.21+, Gin, GORM, PostgreSQL, Redis và JWT.

## Chạy local

```powershell
Copy-Item .env.example .env
go mod tidy
go run ./cmd/server
```

Hoặc chạy đầy đủ dependency bằng Docker:

```powershell
Copy-Item .env.example .env
docker compose up --build
```

Docker Compose cũng có default development values và có thể chạy trực tiếp bằng `docker compose up --build` khi chưa tạo `.env`; hãy dùng `.env` riêng khi cần đổi secret hoặc thông tin database.

API chạy tại `http://localhost:8080`. Health check: `GET /health`.

## Endpoint chính

- `POST /api/v1/auth/register` - đăng ký `{name,email,password}`
- `POST /api/v1/auth/login` - đăng nhập `{email,password}`, nhận JWT
- `GET|PUT|DELETE /api/v1/users/me` - xem, cập nhật hoặc xóa tài khoản hiện tại
- `GET|POST|PUT|DELETE /api/v1/projects[/:id]`
- `GET|POST|PUT|DELETE /api/v1/tasks[/:id]`
- `GET /api/v1/tasks/:task_id/comments`
- `POST /api/v1/comments` - tạo comment `{body,task_id}`
- `GET|PUT|DELETE /api/v1/comments/:id`
- `GET /api/v1/tasks?status=todo`

Các route users/project/task/comment yêu cầu header `Authorization: Bearer <token>`. Token chứa `user_id`, `role`, thời điểm phát hành và thời điểm hết hạn; middleware chỉ chấp nhận chữ ký HS256 với đúng `JWT_SECRET`.

## Authentication và middleware

- `POST /api/v1/auth/register`: kiểm tra input bằng validator, băm password bằng bcrypt và tạo user role `user`.
- `POST /api/v1/auth/login`: xác thực email/password và trả JWT.
- `middleware.Auth`: bảo vệ toàn bộ route `/api/v1` private.
- `middleware.Logger`: ghi method, path, HTTP status và thời gian xử lý request.
- CORS: cho phép frontend gửi request với các method CRUD và header Authorization.
- `middleware.AdminOnly`: phân quyền role `admin`; kiểm tra bằng `GET /api/v1/admin/status`.
  Route `GET /api/v1/admin/status` yêu cầu JWT có claim `role=admin`; user thường nhận `403`.

## Kiến trúc

Request đi qua CORS, logging, recovery và JWT middleware trước handler. Handler chịu trách nhiệm HTTP/validation, repository chịu trách nhiệm truy vấn GORM, PostgreSQL lưu dữ liệu và Redis cache danh sách task trong 1 phút.

## Kiểm thử

```powershell
go test ./...
go vet ./...
```

## API documentation

Import `docs/postman_collection.json` vào Postman. Collection có sẵn các request register, login, CRUD Project, CRUD Task và CRUD Comment. Request login tự lưu JWT vào biến `token`; cập nhật `projectId`, `taskId` và `commentId` theo dữ liệu trả về khi chạy demo.

## Deploy bằng Render

Repo đã có `render.yaml` để tạo web service Docker, PostgreSQL và Redis. Trên Render chọn **New > Blueprint**, kết nối GitHub repository và deploy. `DATABASE_URL`, `REDIS_URL` và `JWT_SECRET` được cấp qua biến môi trường; không commit secret thật. Sau khi deploy, kiểm tra `GET https://<your-service>.onrender.com/health` rồi đổi biến `baseUrl` trong Postman.

## Coverage rubric

- Core API: User, Project, Task và Comment có các endpoint cần thiết; Project, Task và Comment có list/create/get/update/delete.
- Architecture: `cmd`, `internal/{config,database,handler,middleware,models,repository,service}` và `pkg`.
- Optimization: Redis cache cho task list, per-client rate limiting, request logging, health check và Docker multi-stage build.
- Delivery: GitHub Actions chạy format check, `go vet`, test và build; Postman collection dùng cho demo API.

## Nộp bài

Bổ sung họ tên, mã học viên, lớp, link GitHub, link deploy và video demo vào báo cáo cá nhân. Không commit file `.env` hoặc secret thật.
