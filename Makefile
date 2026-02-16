MAKEFLAGS += --no-print-directory

TOOLS_BIN := $(PWD)/tools/bin
export PATH := $(TOOLS_BIN):$(PATH)
export GOBIN := $(TOOLS_BIN)

GOLINES_VERSION := v0.14.0
MOCKGEN_VERSION := v0.6.0
GOFUMPT_VERSION := v0.7.0
GOLANGCI_LINT_VERSION := v2.1.6
SWAG_VERSION := v1.16.6

.PHONY: up down wipe dev-logs logs-back logs-front lint deploy generate test-integration deps fmt fmt-go fmt-gofumpt fmt-lines fmt-fe swagger gen-types docs build-frontend seed tg-login

# Bootstrap
up:
	docker compose up --build -d

down:
	docker compose down

build-frontend:
	cd frontend && npm ci && npm run build

wipe:
	docker compose down -v

dev-logs:
	docker compose logs -f

logs-back:
	docker compose logs -f backend

logs-front:
	docker compose logs -f frontend

# Tooling
deps:
	@mkdir -p $(TOOLS_BIN)
	go install github.com/golangci/golines@$(GOLINES_VERSION)
	go install go.uber.org/mock/mockgen@$(MOCKGEN_VERSION)
	go install mvdan.cc/gofumpt@$(GOFUMPT_VERSION)
	go install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@$(GOLANGCI_LINT_VERSION)
	go install github.com/swaggo/swag/cmd/swag@$(SWAG_VERSION)

fmt-go:
	go fmt ./...

fmt-gofumpt:
	gofumpt -w .

fmt-lines:
	golines -w -m 100 --no-reformat-tags .

fmt-fe:
	cd frontend && npm run format

fmt: fmt-go fmt-gofumpt fmt-lines fmt-fe

lint:
	golangci-lint run ./...
	cd frontend && npm run lint

deploy:
	./deploy.sh

swagger:
	swag init -g internal/http/app/doc.go -o docs --parseDependency --st

gen-types:
	cd frontend && npx swagger2openapi ../docs/swagger.json -o ../docs/openapi.json
	cd frontend && npx openapi-typescript ../docs/openapi.json -o src/types/api.ts
	cd frontend && npm run format -- src/types/api.ts

generate:
	go generate ./...
	$(MAKE) swagger
	$(MAKE) gen-types
	$(MAKE) fmt

docs:
	@echo "Serving API docs at http://localhost:57771"
	docker run --rm -p 57771:8080 -e SWAGGER_JSON=/spec/swagger.json -v $(PWD)/docs:/spec swaggerapi/swagger-ui

test-integration:
	go test -v -tags integration ./integration-tests/...

seed:
	go run ./cmd/seed

tg-login:
	docker run --rm -i --env-file .env -v $(PWD):/app -w /app golang:1.25.6-alpine3.23 sh -lc "go run ./cmd/tglogin"
