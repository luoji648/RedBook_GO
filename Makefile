.PHONY: build run test vet

build:
	go build -o bin/redbook-go ./cmd/server

run:
	go run ./cmd/server

test:
	go test ./...

vet:
	go vet ./...
