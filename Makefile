.PHONY: up down dev-logs lint deploy generate test-integration

up:
	docker compose up --build

down:
	docker compose down

dev-logs:
	docker compose logs -f

lint:
	golangci-lint run ./...
	cd frontend && npm run lint

deploy:
	./deploy.sh

generate:
	go generate ./...
	go fmt ./...

test-integration:
	go test -v -tags integration ./integration-tests/...
