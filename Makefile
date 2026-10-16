.PHONY: run test swagger

back-local:
	go run ./backend/cmd/app

test:
	go test ./...

swagger:
	go -C backend run github.com/swaggo/swag/cmd/swag@v1.16.6 init -g cmd/app/main.go -o docs --parseInternal --outputTypes go,json,yaml --quiet

env-up:
	docker compose up --build -d
