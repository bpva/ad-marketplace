.PHONY: up down dev-logs lint deploy

up:
	docker-compose up --build

down:
	docker-compose down

dev-logs:
	docker-compose logs -f

lint:
	golangci-lint run ./...
	cd frontend && npm run lint

deploy:
	./deploy.sh
