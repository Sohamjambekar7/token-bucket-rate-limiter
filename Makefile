.PHONY: test run redis-up load
test:
	go test -race ./...
run:
	go run ./cmd/server
redis-up:
	docker compose up --build
load:
	k6 run scripts/load.js
