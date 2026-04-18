# 配置参考

本文档详细说明LLM Gateway的所有配置项。

## 配置文件格式

配置文件使用YAML格式，默认路径为`configs/config.yaml`，可以通过`-c`参数指定其他路径。

支持通过环境变量覆盖配置项，格式为：`配置项路径转为大写，用下划线连接`。例如：
- `server.port` → `SERVER_PORT`
- `llm.services.openai.api_key` → `LLM_SERVICES_OPENAI_API_KEY`
- `plugin.ratelimit.default.limit` → `PLUGIN_RATELIMIT_DEFAULT_LIMIT`

## 完整配置示例

```yaml
# 服务配置
server:
  port: 8080                  # 服务监听端口
  mode: release               # 运行模式：debug/release/test
  read_timeout: 60            # 请求读取超时时间（秒）
  write_timeout: 60           # 响应写入超时时间（秒）
  idle_timeout: 120           # 连接空闲超时时间（秒）
  max_header_bytes: 1048576   # 最大请求头大小（字节）
  enable_pprof: false         # 是否开启pprof性能分析端点
  enable_metrics: true        # 是否开启/metrics端点

# 日志配置
log:
  level: info                 # 日志级别：debug/info/warn/error/fatal
  format: json                # 日志格式：json/text
  output_path: stdout         # 输出路径：stdout/stderr/文件路径
  max_size: 100               # 日志文件最大大小（MB）
  max_backups: 10             # 最大保留日志文件数
  max_age: 30                 # 日志文件最大保留天数（天）
  compress: true              # 是否压缩历史日志
  caller_skip: 2              # 调用栈跳过层数
  debug: false                # 是否开启Debug模式，打印调用栈信息

# 认证配置
auth:
  jwt_secret: "your-jwt-secret-key" # JWT签名密钥，生产环境请修改为强随机字符串
  jwt_expire: 168             # API Key默认过期时间（小时，默认7天）
  enable_api_key_auth: true   # 是否开启API Key认证
  enable_jwt_auth: true       # 是否开启JWT Token认证
  allow_anonymous: false      # 是否允许匿名访问（不推荐开启）

# LLM服务配置
llm:
  default_service: "openai"   # 默认使用的LLM服务，当模型不匹配时使用
  connect_timeout: 10         # 连接LLM服务的超时时间（秒）
  response_header_timeout: 30 # 等待LLM响应头的超时时间（秒）
  max_idle_conns: 100         # 最大空闲连接数
  max_idle_conns_per_host: 50 # 每个主机的最大空闲连接数
  idle_conn_timeout: 90       # 空闲连接超时时间（秒）
  disable_compression: false  # 是否禁用压缩
  tls_insecure_skip_verify: false # 是否跳过TLS证书验证（不推荐）

  services:                   # LLM服务列表，可以配置多个
    openai:
      type: openai            # 服务类型：openai/claude/deepseek等
      api_key: "sk-xxxxxxxx"  # API Key
      base_url: "https://api.openai.com/v1" # API基础地址，可以配置为代理地址
      timeout: 60             # 请求超时时间（秒）
      max_retries: 2          # 最大重试次数
      retry_delay: 1          # 重试间隔（秒）
      proxy: ""               # 代理地址，例如：http://127.0.0.1:7890
      enabled: true           # 是否启用该服务
      weight: 100             # 负载均衡权重
      priority: 1             # 优先级，数字越小优先级越高
      models: []              # 支持的模型列表，不配置则自动根据类型匹配
      exclude_models: []      # 排除的模型列表

    claude:
      type: claude
      api_key: "sk-ant-xxxxxxxx"
      base_url: "https://api.anthropic.com/v1"
      timeout: 120
      max_retries: 2
      enabled: true
      weight: 100

    deepseek:
      type: deepseek
      api_key: "sk-xxxxxxxx"
      base_url: "https://api.deepseek.com/v1"
      timeout: 60
      max_retries: 2
      enabled: true

# 插件配置
plugin:
  enabled_plugins:            # 启用的插件列表，按顺序执行
    - auth
    - ratelimit
    - logging
    - metrics
    # - tracing
    # - cache
    # - content_audit
    # - cost_statistics

  # 认证插件配置
  auth:
    enabled: true
    cache_ttl: 60             # API Key缓存时间（秒）
    enable_ip_whitelist: false # 是否启用IP白名单验证

  # 限流插件配置
  ratelimit:
    enabled: true
    default:                  # 默认限流规则
      strategy: token_bucket  # 限流算法：token_bucket/fixed_window/sliding_window
      limit_type: request     # 限流类型：request/token
      limit: 100              # 限制数量
      period: 1s              # 时间窗口
      burst: 200              # 令牌桶突发容量
      priority: 0             # 规则优先级
    rules:                    # 自定义限流规则
      - match:
          user_id: "vip-*"    # 支持通配符匹配
        limit: 1000
        period: 1m
        priority: 10
      - match:
          model: "gpt-4*"
        limit: 50
        period: 1m
        limit_type: token
        priority: 20

  # 日志插件配置
  logging:
    enabled: true
    log_request: true         # 是否记录请求内容
    log_response: false       # 是否记录响应内容（生产环境不建议开启，可能包含敏感数据）
    log_headers: false        # 是否记录请求头
    exclude_paths:            # 不记录日志的路径
      - /health
      - /metrics
    max_body_size: 10240      # 最大记录的请求/响应体大小（字节）

  # 监控插件配置
  metrics:
    enabled: true
    path: "/metrics"          # Metrics端点路径
    enable_go_metrics: true   # 是否收集Go运行时指标
    enable_process_metrics: true # 是否收集进程指标
    buckets:                  # 延迟直方图桶配置（秒）
      - 0.01
      - 0.05
      - 0.1
      - 0.5
      - 1
      - 2.5
      - 5
      - 10
      - 30

  # 链路追踪插件配置
  tracing:
    enabled: false
    endpoint: "http://jaeger:14268/api/traces" # Jaeger/OTLP端点
    service_name: "llm-gateway" # 服务名称
    sampler_type: "probabilistic" # 采样类型：always/never/probabilistic/ratelimiting
    sampler_param: 0.1        # 采样参数，概率采样时是采样率0-1
    enable_baggage: false     # 是否启用Baggage
    max_tag_value_length: 256 # 最大标签值长度

  # 缓存插件配置
  cache:
    enabled: false
    type: memory              # 缓存类型：memory/redis
    ttl: 300                  # 默认缓存过期时间（秒）
    max_size: 10000           # 最大缓存条目数（内存缓存时）
    key_prefix: "llm_cache:"  # 缓存键前缀
    enable_param_hash: true   # 是否对请求参数进行哈希作为缓存键
    cache_by_user: true       # 是否按用户隔离缓存
    cache_methods:            # 需要缓存的方法
      - "chat.completions"
      - "completions"
    exclude_models:           # 不缓存的模型
      - "gpt-4*"

# 存储配置
storage:
  type: memory                # 存储类型：memory/redis
  prefix: "llm_gateway:"      # 存储键前缀

  # Redis配置（type=redis时需要）
  redis:
    addr: "localhost:6379"    # Redis地址
    password: ""              # Redis密码
    db: 0                     # Redis数据库编号
    pool_size: 100            # 连接池大小
    min_idle_conns: 10        # 最小空闲连接数
    max_retries: 3            # 最大重试次数
    dial_timeout: 5           # 连接超时时间（秒）
    read_timeout: 3           # 读取超时时间（秒）
    write_timeout: 3          # 写入超时时间（秒）
    idle_timeout: 300         # 空闲连接超时时间（秒）

# 路由配置
router:
  enabled: true
  default_strategy: "weighted_round_robin" # 默认负载均衡策略：round_robin/weighted_round_robin/least_conn/random
  enable_fallback: true       # 是否启用故障转移
  max_fallback_retries: 2     # 最大故障转移重试次数
  fallback_threshold: 0.5     # 故障转移阈值，错误率超过这个值则自动切换
  health_check_interval: 10   # 健康检查间隔（秒）
  rules:                      # 路由规则
    - match:
        model: "gpt-4*"
        user_id: "internal-*"
      target:
        service: "azure-openai"
        priority: 100
    - match:
        model: "gpt-3.5-turbo*"
      target:
        service: "openai"
        priority: 50
```

## 配置项详细说明

### server 服务配置

| 配置项 | 类型 | 默认值 | 说明 |
|--------|------|--------|------|
| port | int | 8080 | 服务监听的TCP端口 |
| mode | string | release | 运行模式：debug模式会打印更多调试信息，release模式是生产环境推荐模式 |
| read_timeout | int | 60 | 读取客户端请求的最大超时时间，单位秒 |
| write_timeout | int | 60 | 向客户端发送响应的最大超时时间，单位秒 |
| idle_timeout | int | 120 | 空闲连接的超时时间，单位秒 |
| max_header_bytes | int | 1048576 | 请求头的最大大小，单位字节 |
| enable_pprof | bool | false | 是否开启pprof性能分析端点，开启后可以通过`/debug/pprof`访问 |
| enable_metrics | bool | true | 是否开启Prometheus metrics端点，路径是`/metrics` |

### log 日志配置

| 配置项 | 类型 | 默认值 | 说明 |
|--------|------|--------|------|
| level | string | info | 日志级别：debug < info < warn < error < fatal |
| format | string | json | 日志输出格式：json是结构化输出，便于日志系统收集；text是人类可读格式 |
| output_path | string | stdout | 日志输出路径，可以是stdout、stderr或者文件路径 |
| max_size | int | 100 | 日志文件的最大大小，单位MB，超过后自动轮转 |
| max_backups | int | 10 | 最多保留的历史日志文件数 |
| max_age | int | 30 | 历史日志文件的最大保留天数 |
| compress | bool | true | 是否压缩历史日志文件，使用gzip压缩 |
| debug | bool | false | 是否开启Debug模式，会打印调用栈和文件名行号信息，对性能有影响 |

### auth 认证配置

| 配置项 | 类型 | 默认值 | 说明 |
|--------|------|--------|------|
| jwt_secret | string | "" | JWT签名密钥，必须配置，生产环境请使用强随机字符串 |
| jwt_expire | int | 168 | API Key的默认过期时间，单位小时，默认7天 |
| enable_api_key_auth | bool | true | 是否开启API Key认证，开启后请求需要携带`Authorization: Bearer <api_key>` |
| enable_jwt_auth | bool | true | 是否开启JWT Token认证 |
| allow_anonymous | bool | false | 是否允许匿名访问，生产环境不要开启 |

### llm LLM服务配置

#### 全局配置

| 配置项 | 类型 | 默认值 | 说明 |
|--------|------|--------|------|
| default_service | string | "openai" | 默认的LLM服务，当请求的模型不匹配任何服务时使用 |
| connect_timeout | int | 10 | 连接LLM服务的超时时间，单位秒 |
| response_header_timeout | int | 30 | 等待LLM服务响应头的超时时间，单位秒 |
| max_idle_conns | int | 100 | 全局最大空闲HTTP连接数 |
| max_idle_conns_per_host | int | 50 | 每个主机的最大空闲HTTP连接数 |
| idle_conn_timeout | int | 90 | 空闲连接的超时时间，单位秒 |
| tls_insecure_skip_verify | bool | false | 是否跳过TLS证书验证，用于自签名证书的场景，不安全，不推荐开启 |

#### services 服务列表配置

每个LLM服务的配置：

| 配置项 | 类型 | 默认值 | 说明 |
|--------|------|--------|------|
| type | string | - | 服务类型：必须是已支持的类型，例如openai、claude、deepseek等 |
| api_key | string | - | 该服务的API Key，必须配置 |
| base_url | string | - | API的基础地址，可以配置为反向代理地址，默认使用各厂商的官方地址 |
| timeout | int | 60 | 该服务的请求超时时间，单位秒 |
| max_retries | int | 2 | 请求失败的最大重试次数，0表示不重试 |
| retry_delay | int | 1 | 重试的间隔时间，单位秒 |
| proxy | string | "" | 代理服务器地址，例如：http://127.0.0.1:7890 |
| enabled | bool | true | 是否启用该服务 |
| weight | int | 100 | 负载均衡的权重，权重越高分配的流量越多 |
| priority | int | 1 | 服务优先级，数字越小优先级越高，优先级高的服务会被优先选择 |
| models | array | [] | 该服务支持的模型列表，支持通配符，不配置则自动根据服务类型匹配默认模型 |
| exclude_models | array | [] | 该服务排除的模型列表，支持通配符 |

### plugin 插件配置

#### auth 认证插件

| 配置项 | 类型 | 默认值 | 说明 |
|--------|------|--------|------|
| enabled | bool | true | 是否启用认证插件 |
| cache_ttl | int | 60 | API Key信息的缓存时间，单位秒，减少存储查询 |
| enable_ip_whitelist | bool | false | 是否启用IP白名单验证，开启后只有在白名单中的IP才能访问 |

#### ratelimit 限流插件

| 配置项 | 类型 | 默认值 | 说明 |
|--------|------|--------|------|
| enabled | bool | true | 是否启用限流插件 |
| default.strategy | string | "token_bucket" | 默认限流算法：token_bucket（令牌桶，支持突发）、fixed_window（固定窗口）、sliding_window（滑动窗口，最精确但性能稍差） |
| default.limit_type | string | "request" | 限流类型：request（按请求次数）、token（按消耗的Token数量） |
| default.limit | int | 100 | 时间窗口内的最大限制数量 |
| default.period | string | "1s" | 时间窗口：支持s（秒）、m（分钟）、h（小时）、d（天） |
| default.burst | int | 200 | 令牌桶的最大突发容量，只有token_bucket算法支持 |
| rules | array | [] | 自定义限流规则，可以按用户、模型等维度配置更细粒度的限流 |

#### logging 日志插件

| 配置项 | 类型 | 默认值 | 说明 |
|--------|------|--------|------|
| enabled | bool | true | 是否启用日志插件 |
| log_request | bool | true | 是否记录请求体内容 |
| log_response | bool | false | 是否记录响应体内容，生产环境不建议开启，可能包含敏感数据 |
| log_headers | bool | false | 是否记录请求头信息 |
| exclude_paths | array | ["/health", "/metrics"] | 不需要记录日志的请求路径 |
| max_body_size | int | 10240 | 最大记录的请求/响应体大小，单位字节，超过部分会被截断 |

#### metrics 监控插件

| 配置项 | 类型 | 默认值 | 说明 |
|--------|------|--------|------|
| enabled | bool | true | 是否启用监控插件 |
| path | string | "/metrics" | Metrics端点的路径 |
| enable_go_metrics | bool | true | 是否收集Go运行时的指标（GC、内存、协程等） |
| enable_process_metrics | bool | true | 是否收集进程级指标（CPU、内存、文件描述符等） |
| buckets | array | [0.01, 0.05, 0.1, 0.5, 1, 2.5, 5, 10, 30] | 请求延迟直方图的桶配置，单位秒 |

#### tracing 链路追踪插件

| 配置项 | 类型 | 默认值 | 说明 |
|--------|------|--------|------|
| enabled | bool | false | 是否启用链路追踪插件 |
| endpoint | string | "" | OTLP/Jaeger的收集端点地址，例如：http://jaeger:14268/api/traces |
| service_name | string | "llm-gateway" | 在链路追踪系统中显示的服务名称 |
| sampler_type | string | "probabilistic" | 采样类型：always（全采样，不推荐生产环境使用）、never（不采样）、probabilistic（按概率采样）、ratelimiting（按速率限制采样） |
| sampler_param | float | 0.1 | 采样参数：概率采样时是0-1的采样率，速率限制时是每秒采样数 |

### storage 存储配置

| 配置项 | 类型 | 默认值 | 说明 |
|--------|------|--------|------|
| type | string | "memory" | 存储类型：memory（内存存储，单实例部署适用）、redis（Redis存储，分布式部署适用） |
| prefix | string | "llm_gateway:" | 所有存储键的前缀，避免与其他系统冲突 |
| redis.addr | string | "localhost:6379" | Redis服务器地址，host:port格式 |
| redis.password | string | "" | Redis密码，没有则留空 |
| redis.db | int | 0 | Redis数据库编号 |
| redis.pool_size | int | 100 | Redis连接池大小 |

### router 路由配置

| 配置项 | 类型 | 默认值 | 说明 |
|--------|------|--------|------|
| enabled | bool | true | 是否启用路由功能 |
| default_strategy | string | "weighted_round_robin" | 默认负载均衡策略：round_robin（轮询）、weighted_round_robin（加权轮询）、least_conn（最少连接）、random（随机） |
| enable_fallback | bool | true | 是否启用故障转移，当选中的服务不可用时，自动尝试其他可用服务 |
| max_fallback_retries | int | 2 | 最大故障转移重试次数 |
| fallback_threshold | float | 0.5 | 故障转移阈值，服务的错误率超过这个值时会被暂时标记为不可用 |
| health_check_interval | int | 10 | 服务健康检查的间隔时间，单位秒 |
| rules | array | [] | 自定义路由规则，可以按用户、模型等维度灵活配置路由策略 |

## 配置热重载

LLM Gateway支持配置热重载，修改配置文件后不需要重启服务即可生效。

触发热重载的方式：
1. 发送SIGHUP信号：`kill -HUP <pid>`
2. 调用API端点：`POST /-/reload`（需要管理员权限）

注意：部分核心配置修改后需要重启服务才能生效：
- server.port
- server.enable_pprof
- server.enable_metrics
- storage.type
- plugin.enabled_plugins（插件列表修改需要重启）
