MODULE=common

.PHONY: all
all: tidy

.PHONY: tidy
tidy:
	@echo "Running go mod tidy for $(MODULE)..."
	@go mod tidy

.PHONY: test
test:
	@echo "Running tests for $(MODULE)..."
	@go test -v ./...

.PHONY: fmt
fmt:
	@echo "Formatting code for $(MODULE)..."
	@go fmt ./...
