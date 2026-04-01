.PHONY: build run test lint

build:
	go build ./...

run:
	go run ./cmd/server

test:
	go test ./...

lint:
	go vet ./...