PROGRAM=dpmain-apiserver
CONSUMER=dpmain-consumer
CALLBACK_CONSUMER=dpmain-callback-consumer
BIN_DIR=bin

# Tools
WIRE := $(shell command -v wire 2> /dev/null)

.PHONY: all
all: build

.PHONY: build
build: tidy wire
	@echo "Building $(PROGRAM)..."
	@mkdir -p $(BIN_DIR)
	@go build -o $(BIN_DIR)/$(PROGRAM) ./cmd/apiserver
	@echo "✓ Build completed: $(BIN_DIR)/$(PROGRAM)"

.PHONY: build-callback-consumer
build-callback-consumer: tidy
	@echo "Building $(CALLBACK_CONSUMER)..."
	@mkdir -p $(BIN_DIR)
	@go build -o $(BIN_DIR)/$(CALLBACK_CONSUMER) ./cmd/callback_consumer
	@echo "✓ Build completed: $(BIN_DIR)/$(CALLBACK_CONSUMER)"

.PHONY: build-all
build-all: build build-callback-consumer
	@echo "✓ All builds completed"

.PHONY: wire
wire:
ifndef WIRE
	@echo "Installing wire..."
	@go install github.com/google/wire/cmd/wire@latest
endif
	@echo "Generating dependency injection code..."
	@cd cmd/apiserver && $(HOME)/go/bin/wire ./...

.PHONY: run
run: build
	@echo "Starting API Server..."
	@./$(BIN_DIR)/$(PROGRAM)

.PHONY: run-callback-consumer
run-callback-consumer: build-callback-consumer
	@echo "Starting Callback Consumer..."
	@./$(BIN_DIR)/$(CALLBACK_CONSUMER)

.PHONY: test
test:
	@echo "Running tests..."
	@go test -v ./...

.PHONY: fmt
fmt:
	@echo "Formatting code..."
	@go fmt ./...

.PHONY: tidy
tidy:
	@echo "Running go mod tidy..."
	@go mod tidy

.PHONY: clean
clean:
	@echo "Cleaning build artifacts..."
	@rm -rf $(BIN_DIR)
	@echo "✓ Clean completed."

.PHONY: arch
arch:
	@echo "========================================="
	@echo "  DDD Production-Grade Architecture"
	@echo "========================================="
	@echo ""
	@echo "Domains Layer:"
	@echo "  entity/   - etorder, etaccount, etprimitive"
	@echo "  repo/     - rporder, rpaccount (interfaces)"
	@echo "  services/ - svorder, svdiagnosis"
	@echo "  modules/  - mdorder, mdaccount (orchestration)"
	@echo "  apimodel/ - request/response DTOs"
	@echo ""
	@echo "Infra Layer:"
	@echo "  persistence/ - mysql, redis (repo implementations)"
	@echo "  mq/          - lmstfy client"
	@echo ""
	@echo "Server Layer:"
	@echo "  handlers/    - order, account"
	@echo "  routers/     - route registration"
	@echo "  middlewares/ - cors, logger, error handler"
	@echo ""
	@echo "Pkg Layer:"
	@echo "  errorx/ - business errors"
	@echo "  ginx/   - gin extensions"
	@echo "  logger/ - logging"
	@echo ""
