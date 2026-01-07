PROJECT_NAME=oip_backend
MODULES=common dpmain dpsync

.PHONY: all
all: build

.PHONY: help
help:
	@echo "========================================="
	@echo "  $(PROJECT_NAME) - Makefile"
	@echo "========================================="
	@echo "Usage:"
	@echo "  make build         - 构建所有模块"
	@echo "  make tidy          - 运行 go mod tidy（所有模块）"
	@echo "  make test          - 运行测试（所有模块）"
	@echo "  make fmt           - 格式化代码（所有模块）"
	@echo "  make clean         - 清理构建产物"
	@echo "  make run-dpmain    - 启动 dpmain API 服务"
	@echo "  make run-dpsync    - 启动 dpsync Worker 服务"
	@echo ""

.PHONY: tidy
tidy:
	@echo "Running go mod tidy for all modules..."
	@for module in $(MODULES); do \
		echo "  -> $$module"; \
		cd $$module && go mod tidy && cd ..; \
	done
	@echo "✓ All modules tidied"

.PHONY: build
build: tidy
	@echo "Building all modules..."
	@echo "  -> common (no binary)"
	@echo "  -> dpmain"
	@cd dpmain && make build
	@echo "  -> dpsync"
	@cd dpsync && make build
	@echo "✓ Build completed"

.PHONY: test
test:
	@echo "Running tests for all modules..."
	@for module in $(MODULES); do \
		echo "  -> $$module"; \
		cd $$module && go test -v ./... && cd ..; \
	done
	@echo "✓ All tests passed"

.PHONY: fmt
fmt:
	@echo "Formatting code for all modules..."
	@for module in $(MODULES); do \
		echo "  -> $$module"; \
		cd $$module && go fmt ./... && cd ..; \
	done
	@echo "✓ All code formatted"

.PHONY: clean
clean:
	@echo "Cleaning build artifacts..."
	@cd dpmain && make clean
	@cd dpsync && make clean
	@echo "✓ Clean completed"

.PHONY: run-dpmain
run-dpmain:
	@echo "Starting dpmain API Server..."
	@cd dpmain && make run

.PHONY: run-dpsync
run-dpsync:
	@echo "Starting dpsync Worker..."
	@cd dpsync && make run

.PHONY: docker-compose-up
docker-compose-up:
	@echo "Starting infrastructure (MySQL, Redis, Lmstfy)..."
	@docker-compose up -d
	@echo "✓ Infrastructure started"

.PHONY: docker-compose-down
docker-compose-down:
	@echo "Stopping infrastructure..."
	@docker-compose down
	@echo "✓ Infrastructure stopped"
