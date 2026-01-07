#!/bin/bash

set -e

echo "Building dpsync..."
mkdir -p bin
go build -o bin/dpsync-worker ./cmd/worker
echo "Build completed: bin/dpsync-worker"
