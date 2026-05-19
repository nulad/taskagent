.PHONY: build run test lint docker-build compose-up compose-down compose-logs compose-smoke

build:
	go build ./...

run:
	go run ./cmd/server

test:
	go test ./...

lint:
	go vet ./...

docker-build:
	docker build -t taskagent:local .

compose-up:
	docker compose up --build

compose-down:
	docker compose down

compose-logs:
	docker compose logs -f taskagent

compose-smoke:
	@echo "Delegating to smoke verification (see TASK-005)"
	@echo "Run 'make compose-logs' separately to inspect service logs."
