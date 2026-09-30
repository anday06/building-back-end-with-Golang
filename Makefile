.PHONY: run test fmt vet docker-up docker-down docker-logs
run:
	go run ./cmd/server
test:
	go test ./...
fmt:
	gofmt -w .
vet:
	go vet ./...
docker-up:
	docker compose up --build -d
docker-down:
	docker compose down
docker-logs:
	docker compose logs -f app
