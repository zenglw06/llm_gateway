# 架构设计

## 设计理念

LLM Gateway的设计遵循以下原则：

1. **协议兼容优先**：100%兼容OpenAI API协议，降低用户迁移成本
2. **插件化扩展**：核心功能稳定，通过插件扩展能力，避免核心代码臃肿
3. **高性能低损耗**：最小化网关的性能损耗，不影响LLM服务本身的响应速度
4. **高可用性**：内置熔断、降级、重试机制，确保服务稳定性
5. **可观测性**：完整的日志、Metrics、链路追踪，方便问题排查
6. **云原生友好**：支持容器化、K8s部署、配置热重载等云原生特性

## 整体架构

```
┌─────────────────────────────────────────────────────────────────┐
│                        API Gateway Layer                        │
├──────────┬──────────┬──────────┬──────────┬──────────┬──────────┤
│  Auth    │  Rate    │  Logging │ Metrics  │ Tracing  │ Custom   │
│ Plugin   │ Limit    │ Plugin   │ Plugin   │ Plugin   │ Plugins  │
└──────────┴──────────┴──────────┴──────────┴──────────┴──────────┘
                                │
                                ▼
┌─────────────────────────────────────────────────────────────────┐
│                          Router Layer                           │
├──────────────────┬──────────────────┬───────────────────────────┤
│ Model Routing    │ Load Balancing   │ Fallback / Failover       │
└──────────────────┴──────────────────┴───────────────────────────┘
                                │
                                ▼
┌─────────────────────────────────────────────────────────────────┐
│                         LLM Adapter Layer                       │
├──────────┬──────────┬──────────┬──────────┬──────────┬──────────┤
│ OpenAI   │ Claude   │ DeepSeek │ Qwen     │ ERNIE    | Custom   │
│ Adapter  │ Adapter  │ Adapter  │ Adapter  │ Adapter  | Adapters │
└──────────┴──────────┴──────────┴──────────┴──────────┴──────────┘
                                │
                                ▼
┌─────────────────────────────────────────────────────────────────┐
│                         Infrastructure                          │
├──────────┬──────────┬──────────┬──────────┬──────────┬──────────┤
│ Storage  │ Logging  │ Metrics  │ Tracing  │ Config   │ Cache    │
└──────────┴──────────┴──────────┴──────────┴──────────┴──────────┘
```

## 核心模块详解

### 1. API Gateway 层

API Gateway层是请求的入口，负责处理HTTP协议、参数校验、插件执行。

#### 核心组件：
- **HTTP Server**：基于Gin框架，高性能HTTP请求处理
- **参数校验**：自动校验请求参数的合法性，返回统一的错误格式
- **插件执行引擎**：按顺序执行插件链，支持请求前、响应后、错误时三个执行点
- **协议转换**：统一处理SSE流式响应格式，兼容不同厂商的流式接口

#### 插件执行流程：
```
Request → [Auth Plugin] → [Rate Limit Plugin] → [Logging Plugin] → [Metrics Plugin] → ... → Route
Response ← [Auth Plugin] ← [Rate Limit Plugin] ← [Logging Plugin] ← [Metrics Plugin] ← ... ← LLM Response
Error ← [Error Handling Plugins] ← Any error during processing
```

### 2. Router 层

Router层负责请求的路由调度，选择合适的LLM服务处理请求。

#### 核心能力：
- **模型匹配**：根据请求的model参数匹配对应的LLM服务
- **负载均衡**：支持轮询、加权轮询、最少连接等负载均衡策略
- **故障转移**：当某个LLM服务不可用时，自动切换到备用服务
- **流量染色**：支持按用户、租户、特征等进行流量灰度
- **成本优化**：自动选择成本最低的可用服务处理请求

#### 路由规则示例：
```yaml
router:
  rules:
    - match:
        model: "gpt-4*"
        user_id: "vip-*"
      target:
        service: "openai"
        priority: 100
    - match:
        model: "gpt-4*"
      target:
        service: "azure-openai"
        priority: 50
    - default:
        service: "openai"
```

### 3. LLM Adapter 层

LLM Adapter层是LLM服务的适配层，统一不同厂商的API差异。

#### 统一接口定义：
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

#### 适配器职责：
- **请求转换**：将统一的请求格式转换为对应厂商的API请求格式
- **签名认证**：处理各厂商的API签名和认证逻辑
- **响应转换**：将厂商的响应格式转换为统一的OpenAI兼容格式
- **错误处理**：统一处理各厂商的错误码，转换为标准的错误格式
- **流式处理**：处理SSE流式响应，统一流式数据格式

### 4. Infrastructure 层

基础设施层提供基础的技术支撑能力。

#### 核心组件：
- **存储层**：支持内存、Redis等存储，用于保存API Key、配额、限流数据等
  - 内存存储适合单实例部署，性能最高
  - Redis存储适合分布式部署，支持多实例共享状态
- **日志系统**：基于Zap的高性能日志，支持结构化输出、上下文传递
- **Metrics**：基于Prometheus的指标收集，内置核心业务指标
- **链路追踪**：基于OpenTelemetry的全链路追踪，支持Jaeger、Zipkin
- **配置中心**：支持配置热重载，多配置源（文件、环境变量、远程配置）
- **缓存层**：支持请求级缓存，降低重复请求的LLM调用成本

## 数据流详解

### 非流式请求处理流程：

```
1. 客户端发送HTTP请求到API Gateway
2. HTTP Server解析请求，进行参数校验
3. 执行请求前置插件链：
   - Auth插件：验证API Key有效性，获取用户信息
   - RateLimit插件：检查用户配额和限流规则
   - Logging插件：记录请求开始日志
   - Metrics插件：统计请求量指标
   - 其他自定义插件
4. Router层根据请求模型和路由规则选择合适的LLM服务
5. LLM Adapter转换请求格式，调用对应厂商的API
6. 收到LLM服务响应，转换为统一格式
7. 执行响应后置插件链：
   - Metrics插件：统计请求延迟、Token用量
   - Logging插件：记录响应日志
   - 其他自定义插件
8. 返回响应给客户端
```

### 流式请求处理流程：

```
1-5. 与非流式请求流程相同
6. LLM Adapter发起流式请求，获取响应流
7. 启动协程异步读取流式响应，逐块转换为统一的SSE格式
8. 边转换边发送给客户端，同时统计Token用量和延迟
9. 流式响应结束后，执行响应后置插件链，更新用量和指标
10. 关闭连接
```

### 错误处理流程：

```
任何环节发生错误时：
1. 中断正常处理流程
2. 执行错误处理插件链：
   - Metrics插件：统计错误量指标
   - Logging插件：记录错误日志
   - Tracing插件：记录错误信息到链路
   - 降级插件：如果配置了降级策略，返回降级响应
3. 返回统一格式的错误响应给客户端
```

## 部署架构

### 单实例部署（适合开发、测试、小规模场景）

```
┌─────────────┐
│   Clients   │
└──────┬──────┘
       │
       ▼
┌─────────────┐
│ LLM Gateway │
└──────┬──────┘
       │
       ▼
┌─────────────┐
│ LLM Services│
└─────────────┘
```

### 高可用部署（适合生产环境）

```
┌─────────────┐
│    Nginx    │ 负载均衡
└──────┬──────┘
       │
       ▼
┌──────┴──────┐
│ LLM Gateway │ 实例1
├─────────────┤
│ LLM Gateway │ 实例2
├─────────────┤
│ LLM Gateway │ 实例N
└──────┬──────┘
       │
       ▼
┌─────────────┐
│   Redis     │ 共享存储（限流、配额、会话）
└─────────────┘
       │
       ▼
┌─────────────┐
│ LLM Services│
└─────────────┘
```

### 云原生部署（Kubernetes）

```
┌─────────────┐
│ Ingress     │
└──────┬──────┘
       │
       ▼
┌─────────────┐
│ Service     │
└──────┬──────┘
       │
       ▼
┌─────────────┐
│ Deployment  │ 多实例部署，自动扩缩容
└──────┬──────┘
       │
       ▼
┌─────────────┐
│ Redis Cache │
└─────────────┘
```

## 性能设计

### 性能优化点：
1. **零拷贝**：流式响应处理使用零拷贝技术，减少内存拷贝
2. **连接池**：LLM服务的HTTP连接池复用，减少连接建立开销
3. **异步处理**：非核心逻辑（日志、指标上报等）异步处理，不阻塞主流程
4. **内存复用**：使用sync.Pool复用对象，减少GC压力
5. **高效序列化**：使用高性能JSON序列化库，降低序列化开销
6. **无锁设计**：核心路径尽量减少锁的使用，提高并发性能

### 性能指标：
- 网关本身的延迟开销：< 5ms
- 最大QPS：> 2000（8核16G配置）
- 内存占用：< 512MB（正常负载）
- CPU使用率：< 70%（2000 QPS时）
