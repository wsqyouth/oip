#!/bin/bash

set -e

echo "Building dpmain..."
mkdir -p bin
go build -o bin/dpmain-apiserver ./cmd/apiserver
echo "✓ Build completed: bin/dpmain-apiserver"
