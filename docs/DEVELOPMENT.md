# 开发指南

欢迎参与LLM Gateway的开发！本文档会帮助你快速搭建开发环境，了解项目结构和开发流程。

## 开发环境要求

- Go 1.22+
- Git
- Make（可选，方便执行编译、测试等命令）
- GolangCI-Lint（可选，代码质量检查）
- Pre-commit（可选，提交前自动检查）

## 环境搭建

### 1. 克隆代码
```bash
git clone https://github.com/your-org/llm-gateway.git
cd llm-gateway
```

### 2. 安装依赖
```bash
go mod download
```

### 3. 安装开发工具（可选）
```bash
# 安装golangci-lint
go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest

# 安装pre-commit
pip install pre-commit
pre-commit install
```

### 4. 准备配置文件
```bash
cp configs/config.example.yaml configs/config.yaml
# 编辑configs/config.yaml，填入你的LLM服务API Key等信息
```

## 项目结构

```
llm-gateway/
├── cmd/                     # 命令行入口
│   └── gateway/            # 主程序入口
│       └── main.go         # 主函数
├── configs/                 # 配置文件
│   └── config.example.yaml # 示例配置
├── internal/                # 内部代码，不对外暴露
│   ├── apiserver/          # HTTP API服务层
│   │   ├── handler.go      # 请求处理函数
│   │   ├── middleware.go   # 中间件
│   │   └── router.go       # 路由定义
│   ├── llm/                # LLM服务适配层
│   │   ├── openai/         # OpenAI适配
│   │   ├── claude/         # Claude适配
│   │   ├── deepseek/       # DeepSeek适配
│   │   ├── interface.go    # LLM服务统一接口
│   │   └── registry.go     # 服务注册工厂
│   ├── model/              # 数据模型定义
│   │   ├── api.go          # API请求/响应模型
│   │   ├── apikey.go       # API Key模型
│   │   ├── quota.go        # 配额模型
│   │   └── plugin.go       # 插件相关模型
│   ├── plugin/             # 插件系统
│   │   ├── auth/           # 认证插件
│   │   ├── ratelimit/      # 限流插件
│   │   ├── logging/        # 日志插件
│   │   ├── metrics/        # 监控插件
│   │   ├── tracing/        # 链路追踪插件
│   │   ├── interface.go    # 插件统一接口
│   │   └── manager.go      # 插件管理器
│   ├── service/            # 业务逻辑层
│   │   ├── apikey.go       # API Key服务
│   │   ├── quota.go        # 配额服务
│   │   └── router.go       # 路由服务
│   └── storage/            # 存储层
│       ├── memory/         # 内存存储实现
│       ├── redis/          # Redis存储实现
│       └── interface.go    # 存储统一接口
├── pkg/                    # 公共包，可以对外提供
│   ├── config/             # 配置加载
│   ├── errors/             # 统一错误定义
│   ├── logger/             # 日志组件
│   ├── metrics/            # 指标组件
│   ├── tracing/            # 链路追踪组件
│   └── utils/              # 通用工具函数
├── test/                   # 测试代码
│   ├── integration/        # 集成测试
│   ├── benchmark/          # 性能测试
│   └── e2e/                # 端到端测试
├── docs/                   # 文档
├── deploy/                 # 部署相关文件
├── examples/               # 示例代码
├── Makefile                # 构建脚本
├── go.mod                  # Go依赖
└── go.sum                  # 依赖校验
```

## 核心概念

### 1. LLM Service 接口
所有LLM厂商的适配都需要实现`LLMService`接口：
```go
type LLMService interface {
    Name() string
    SupportsModel(model string) bool
    ChatCompletion(ctx context.Context, req *model.ChatRequest) (*model.ChatResponse, error)
    ChatCompletionStream(ctx context.Context, req *model.ChatRequest) (<-chan *model.StreamResponse, error)
    Completion(ctx context.Context, req *model.CompletionRequest) (*model.CompletionResponse, error)
    CompletionStream(ctx context.Context, req *model.CompletionRequest) (<-chan *model.StreamResponse, error)
}
```

### 2. Plugin 接口
所有插件都需要实现`Plugin`接口：
```go
type Plugin interface {
    Name() string
    Init(config map[string]interface{}) error
    Close() error
    HandleRequest(ctx context.Context, req *model.LLMRequest) (*model.LLMRequest, error)
    HandleResponse(ctx context.Context, resp *model.LLMResponse) (*model.LLMResponse, error)
    HandleError(ctx context.Context, err error) error
}
```

### 3. Storage 接口
存储层需要实现`Storage`接口，支持多种存储后端：
```go
type Storage interface {
    // API Key相关
    CreateAPIKey(ctx context.Context, apikey *model.APIKey) error
    GetAPIKeyByID(ctx context.Context, id string) (*model.APIKey, error)
    GetAPIKeyByKey(ctx context.Context, key string) (*model.APIKey, error)
    UpdateAPIKey(ctx context.Context, apikey *model.APIKey) error
    DeleteAPIKey(ctx context.Context, id string) error
    ListAPIKeysByUserID(ctx context.Context, userID string, limit, offset int) ([]*model.APIKey, int64, error)

    // 配额相关
    GetQuota(ctx context.Context, userID string) (*model.Quota, error)
    UpdateQuota(ctx context.Context, quota *model.Quota) error
    IncrementUsage(ctx context.Context, userID string, model string, tokens int) error
    ResetQuota(ctx context.Context, userID string) error

    // 限流相关
    GetRateLimit(ctx context.Context, key string) (int, error)
    IncrementRateLimit(ctx context.Context, key string, period int64) (int, error)
}
```

## 开发流程

### 1. 选择/创建 Issue
- 对于新功能，先创建Issue描述需求和设计方案
- 对于Bug修复，先创建Issue描述复现步骤
- 回复Issue表明你正在处理，避免重复工作

### 2. 创建 Feature 分支
从main分支创建新的功能分支：
```bash
git checkout -b feature/your-feature-name main
# 或修复Bug的分支
git checkout -b fix/your-bug-fix main
```

### 3. 开发代码
遵循以下规范：
- 所有公共方法、结构体、常量必须有注释
- 代码格式遵循Go官方规范，使用`gofmt`格式化
- 新功能必须包含对应的单元测试
- 重大变更需要更新文档

### 4. 本地测试
```bash
# 运行单元测试
make test

# 运行代码检查
make lint

# 编译
make build

# 本地运行测试
make run
```

### 5. 提交代码
提交信息遵循[约定式提交](https://www.conventionalcommits.org/zh-hans/v1.0.0/)规范：
```
<类型>[可选 作用域]: <描述>

[可选 正文]

[可选 页脚]
```

类型说明：
- `feat`: 新功能
- `fix`: 修复Bug
- `docs`: 文档更新
- `style`: 代码格式调整（不影响代码运行）
- `refactor`: 重构（既不是新增功能，也不是修改Bug的代码变动）
- `perf`: 性能优化
- `test`: 测试相关
- `chore`: 构建/工具/依赖等相关的变动

示例：
```
feat(llm): 新增对通义千问的支持

- 实现通义千问的适配
- 支持流式和非流式响应
- 添加对应的单元测试

Closes #123
```

### 6. 创建 Pull Request
- PR标题遵循提交信息格式
- PR描述需要包含：
  - 变更的内容和目的
  - 相关的Issue链接
  - 测试情况
  - 任何需要注意的变更（不兼容变更、配置变更等）
- 等待CI检查通过
- 等待代码审查，根据反馈修改
- 审查通过后会被合并到main分支

## 开发指南

### 添加新的LLM厂商支持
步骤：
1. 在`internal/llm/`目录下创建新的厂商目录，例如`internal/llm/qwen/`
2. 实现`LLMService`接口
3. 在`init()`函数中注册服务工厂：
```go
func init() {
    llm.Register("qwen", func(cfg *llm.Config) (llm.LLMService, error) {
        return NewQwenService(cfg), nil
    })
}
```
4. 添加对应的单元测试
5. 在`docs/API.md`中添加支持的模型说明
6. 在`README.md`的支持厂商列表中添加

### 添加新的插件
步骤：
1. 在`internal/plugin/`目录下创建新的插件目录，例如`internal/plugin/content_audit/`
2. 实现`Plugin`接口
3. 在`init()`函数中注册插件：
```go
func init() {
    plugin.Register("content_audit", func() plugin.Plugin {
        return NewContentAuditPlugin()
    })
}
```
4. 在配置结构体中添加插件配置
5. 添加对应的单元测试
6. 在`docs/CONFIGURATION.md`中添加插件配置说明
7. 在`docs/ARCHITECTURE.md`的插件部分添加说明

### 添加新的存储后端
步骤：
1. 在`internal/storage/`目录下创建新的存储目录，例如`internal/storage/mysql/`
2. 实现`Storage`接口
3. 在存储工厂中注册：
```go
func NewStorage(cfg *config.StorageConfig) (Storage, error) {
    switch cfg.Type {
    case "memory":
        return memory.NewMemoryStorage(cfg), nil
    case "redis":
        return redis.NewRedisStorage(cfg), nil
    case "mysql":
        return mysql.NewMySQLStorage(cfg), nil
    default:
        return nil, errors.New("unsupported storage type")
    }
}
```
4. 添加对应的单元测试
5. 在`docs/CONFIGURATION.md`中添加存储配置说明

## 测试规范

### 单元测试
- 测试文件命名为`*_test.go`，和被测试文件放在同一目录
- 测试覆盖率要求：核心模块达到80%以上
- 测试用例要覆盖正常场景、边界场景、错误场景
- 使用`testify/assert`库进行断言，提高测试代码可读性

示例：
```go
func TestOpenAIService_ChatCompletion(t *testing.T) {
    // 准备测试数据
    cfg := &llm.Config{
        APIKey: "test-key",
    }
    svc := openai.NewOpenAIService(cfg)

    // 测试正常场景
    req := &model.ChatRequest{
        Model: "gpt-3.5-turbo",
        Messages: []model.ChatMessage{
            {Role: "user", Content: "Hello"},
        },
    }
    resp, err := svc.ChatCompletion(context.Background(), req)
    assert.NoError(t, err)
    assert.NotNil(t, resp)
    assert.Equal(t, "gpt-3.5-turbo", resp.Model)
}
```

### 集成测试
- 集成测试放在`test/integration/`目录下
- 测试需要外部依赖的场景，例如LLM服务调用、Redis存储等
- 使用Mock服务避免依赖外部服务

### 性能测试
- 性能测试放在`test/benchmark/`目录下
- 使用Go的基准测试框架，文件命名为`*_test.go`
- 测试性能优化效果，避免性能回退

示例：
```go
func BenchmarkChatCompletion(b *testing.B) {
    svc := openai.NewOpenAIService(&llm.Config{APIKey: "test-key"})
    req := &model.ChatRequest{
        Model: "gpt-3.5-turbo",
        Messages: []model.ChatMessage{
            {Role: "user", Content: "Hello"},
        },
    }

    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        _, err := svc.ChatCompletion(context.Background(), req)
        if err != nil {
            b.Fatal(err)
        }
    }
}
```

## 代码审查指南

审查者需要检查以下内容：
- 代码是否符合Go规范，格式是否正确
- 功能是否正确实现，是否满足需求
- 是否有对应的测试用例，测试是否覆盖主要场景
- 性能是否有明显下降
- 文档是否同步更新
- 是否引入安全风险
- 是否有不兼容的变更，是否在Release Note中说明

## 版本发布

版本号遵循[语义化版本号](https://semver.org/lang/zh-CN/)规范：
- 主版本号：不兼容的API变更
- 次版本号：向下兼容的功能新增
- 修订号：向下兼容的问题修正

发布流程：
1. 更新CHANGELOG.md，记录本次版本的变更内容
2. 打版本标签：`git tag v1.0.0`
3. 推送标签：`git push origin v1.0.0`
4. CI自动构建多平台二进制、Docker镜像，创建GitHub Release
5. 更新官方文档

## 常见问题

### Q：如何调试代码？
A：可以使用GoLand/VS Code的调试功能，在main函数处打断点调试。也可以在代码中添加`log.Debug()`打印调试信息。

### Q：如何添加新的配置项？
A：在`pkg/config/config.go`的结构体中添加对应的字段，然后在`configs/config.example.yaml`中添加示例配置，最后在`docs/CONFIGURATION.md`中添加说明。

### Q：如何处理LLM厂商的流式响应？
A：参考现有OpenAI/Claude的实现，使用bufio.Scanner读取响应流，逐行解析，转换为统一的StreamResponse格式，通过channel返回。

### Q：如何贡献文档？
A：文档在`docs/`目录下，都是Markdown格式，直接修改后提交PR即可。如果是新增文档，需要在`README.md`中添加链接。

## 社区交流
如果你在开发过程中遇到问题，可以通过以下方式交流：
- GitHub Issues：提交Bug和功能请求
- GitHub Discussions：讨论开发相关问题
- 微信群：添加微信备注"llm-gateway开发"加入开发者群
