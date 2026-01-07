PROGRAM=dpsync-worker
BIN_DIR=bin

.PHONY: all
all: build

.PHONY: build
build: tidy
	@echo "Building $(PROGRAM)..."
	@mkdir -p $(BIN_DIR)
	@go build -o $(BIN_DIR)/$(PROGRAM) ./cmd/worker
	@echo "Build completed: $(BIN_DIR)/$(PROGRAM)"

.PHONY: run
run: build
	@echo "Starting Worker..."
	@./$(BIN_DIR)/$(PROGRAM)

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
	@echo "Clean completed."
