#!/bin/bash

# 测试脚本
set -e

echo "Running all tests..."
go test -v ./...

echo -e "\nRunning code lint..."
if command -v golangci-lint &> /dev/null; then
    golangci-lint run ./...
else
    echo "golangci-lint not installed, skipping lint check"
fi

echo -e "\nAll checks passed!"
