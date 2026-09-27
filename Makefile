.PHONY: all lint test clean build

# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOCET=$(GOCMD) test
GOGET=$(GOCMD) get
GOFMT=$(GOCMD) fmt
BINARY_NAME=auth-service
GOLANGCI_LINT=$(shell command -v golangci-lint 2>/dev/null || if [ -n "$$(go env GOBIN)" ]; then printf '%s/golangci-lint' "$$(go env GOBIN)"; else printf '%s/bin/golangci-lint' "$$(go env GOPATH)"; fi)

all: lint test clean

lint:
	@echo "running Lint check"
	$(GOFMT) ./...
	go vet ./...
	@if [ -x "$(GOLANGCI_LINT)" ]; then \
	  		"$(GOLANGCI_LINT)" run; \
  	else \
  	  	echo "golangci-lint not installed. Run 'go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest'"; \
  	fi

test:
	@echo "Running tests..."
	$(GOCMD) test -v -race -cover ./...

migrate:
	@echo "Running database migrations..."
	for f in migrations/*.sql; do \
        docker exec -i postgres_auth psql -U postgres -d auth_db < "$$f"; \
    done

clean:
	@echo "Cleaning up..."
	@rm -rf bin/
	$(GOCMD) clean