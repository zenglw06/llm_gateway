.PHONY: all build test clean run fmt lint tidy build-all docker-build

# 项目信息
APP_NAME := llm-gateway
VERSION := 1.0.0
BUILD_TIME := $(shell date +%Y%m%d%H%M%S)
GIT_COMMIT := $(shell git rev-parse HEAD 2>/dev/null || echo "unknown")
GO_VERSION := $(shell go version | awk '{print $$3}')

# 编译参数
LDFLAGS := -X main.version=$(VERSION) \
           -X main.buildTime=$(BUILD_TIME) \
           -X main.gitCommit=$(GIT_COMMIT) \
           -X main.goVersion=$(GO_VERSION)

# 默认目标
all: build

# 编译主程序
build:
	@mkdir -p bin
	go build -ldflags "$(LDFLAGS)" -o bin/$(APP_NAME) ./cmd/gateway
	@echo "Build completed: bin/$(APP_NAME)"

# 运行所有测试
test:
	go test -v ./...

# 运行测试并生成覆盖率报告
test-coverage:
	@mkdir -p coverage
	go test -coverprofile=coverage/coverage.out ./...
	go tool cover -html=coverage/coverage.out -o coverage/coverage.html
	@echo "Coverage report generated: coverage/coverage.html"

# 清理编译产物
clean:
	rm -rf bin/ coverage/
	go clean

# 直接运行程序
run: build
	./bin/$(APP_NAME) -c configs/config.yaml

# 格式化代码
fmt:
	go fmt ./...

# 代码检查
lint:
	golangci-lint run ./...

# 整理依赖
tidy:
	go mod tidy

# 交叉编译所有平台版本
build-all:
	@mkdir -p bin
	# Linux amd64
	GOOS=linux GOARCH=amd64 go build -ldflags "$(LDFLAGS)" -o bin/$(APP_NAME)-linux-amd64 ./cmd/gateway
	# Linux arm64
	GOOS=linux GOARCH=arm64 go build -ldflags "$(LDFLAGS)" -o bin/$(APP_NAME)-linux-arm64 ./cmd/gateway
	# macOS amd64
	GOOS=darwin GOARCH=amd64 go build -ldflags "$(LDFLAGS)" -o bin/$(APP_NAME)-darwin-amd64 ./cmd/gateway
	# macOS arm64
	GOOS=darwin GOARCH=arm64 go build -ldflags "$(LDFLAGS)" -o bin/$(APP_NAME)-darwin-arm64 ./cmd/gateway
	# Windows amd64
	GOOS=windows GOARCH=amd64 go build -ldflags "$(LDFLAGS)" -o bin/$(APP_NAME)-windows-amd64.exe ./cmd/gateway
	@echo "All platforms build completed"

# 构建Docker镜像
docker-build:
	docker build -t $(APP_NAME):$(VERSION) .
