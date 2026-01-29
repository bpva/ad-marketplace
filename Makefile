MAKEFLAGS += --no-print-directory

TOOLS_BIN := $(PWD)/tools/bin
export PATH := $(TOOLS_BIN):$(PATH)
export GOBIN := $(TOOLS_BIN)

GOLINES_VERSION := v0.14.0
MOCKGEN_VERSION := v0.6.0
GOFUMPT_VERSION := v0.7.0
GOLANGCI_LINT_VERSION := v2.1.6

.PHONY: up down wipe dev-logs logs-back logs-front lint deploy generate test-integration deps fmt fmt-go fmt-gofumpt fmt-lines

# Bootstrap
up:
	docker compose up --build -d

down:
	docker compose down

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

fmt-go:
	go fmt ./...

fmt-gofumpt:
	gofumpt -w .

fmt-lines:
	golines -w -m 100 .

fmt: fmt-go fmt-gofumpt fmt-lines

lint:
	golangci-lint run ./...
	cd frontend && npm run lint

deploy:
	./deploy.sh

generate:
	go generate ./...
	$(MAKE) fmt

test-integration:
	go test -v -tags integration ./integration-tests/...
