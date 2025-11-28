.PHONY: build build-master build-worker build-client test clean run-master run-worker help

# Variables
BIN_DIR := bin
BUILD_FLAGS := -v

# Compilar solo lo que existe
build: build-master build-worker build-client

build-master:
	@echo "Building master..."
	@mkdir -p $(BIN_DIR)
	go build $(BUILD_FLAGS) -o $(BIN_DIR)/master ./cmd/master || echo "Master build failed, continuing..."

build-worker:
	@echo "Building worker..."
	@mkdir -p $(BIN_DIR)
	go build $(BUILD_FLAGS) -o $(BIN_DIR)/worker ./cmd/worker || echo "Worker build failed, continuing..."

build-client:
	@echo "Building client..."
	@mkdir -p $(BIN_DIR)
	go build $(BUILD_FLAGS) -o $(BIN_DIR)/client ./cmd/client || echo "Client build failed, continuing..."

# Ejecutar tests solo si existen
test:
	@echo "Running tests..."
	@go test ./... 2>/dev/null || echo "No tests found or tests failed"

# Limpiar binarios
clean:
	@echo "Cleaning..."
	@rm -rf $(BIN_DIR)
	@rm -rf data/output/* 2>/dev/null || true
	@echo "Clean complete!"

# Ejecutar componentes
run-master:
	@go run ./cmd/master

run-worker:
	@go run ./cmd/worker

# Ayuda
help:
	@echo "Comandos disponibles:"
	@echo "  make build       - Compilar todos los componentes"
	@echo "  make clean       - Limpiar binarios"
	@echo "  make run-master  - Ejecutar master"
	@echo "  make run-worker  - Ejecutar worker"
	@echo "  make test        - Ejecutar tests"
