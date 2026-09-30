.PHONY: run test fmt vet
run:
	go run ./cmd/server
test:
	go test ./...
fmt:
	gofmt -w .
vet:
	go vet ./...
