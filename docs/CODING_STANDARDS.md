# 代码规范

本文档定义LLM Gateway项目的Go代码编写规范，所有贡献者必须遵循。

## 1. 基础规范

### 1.1 代码格式
- 使用`gofmt`格式化代码，不需要手动调整缩进（Go标准是4空格）
- 每行长度不超过120字符，超过时合理换行
- 运算符两侧加空格，例如：`a + b * c`
- 逗号后面加空格，例如：`fmt.Println(a, b, c)`
- 代码块之间合理空行，提高可读性

### 1.2 命名规范
#### 1.2.1 包命名
- 包名使用小写，不包含下划线和大写字母
- 简短有意义，尽量和目录名保持一致
- 避免和标准库重名，例如不要命名为`fmt`、`os`等
- 包名尽量使用名词，例如`model`、`service`、`storage`

#### 1.2.2 文件名
- 文件名使用小写，单词之间用下划线分隔
- 测试文件以`_test.go`结尾
- 平台相关文件包含系统名，例如`file_windows.go`、`file_linux.go`

#### 1.2.3 函数/方法命名
- 使用驼峰命名法，公共函数首字母大写，私有函数首字母小写
- 函数名要能够清晰表达函数的功能，见名知意
- 动词开头，例如`GetUser`、`CreateAPIKey`、`CalculateQuota`
- 返回bool类型的函数可以用`Is`、`Has`、`Can`等前缀，例如`IsEnabled`、`HasPermission`
- 避免使用模糊的动词，例如`do`、`handle`、`process`，除非上下文非常清晰

#### 1.2.4 变量/常量命名
- 局部变量尽量简短，例如循环变量用`i`、`j`，参数用`req`、`resp`等
- 全局变量/常量使用驼峰命名，公共常量首字母大写
- 常量尽量使用枚举类型，避免魔术数字
- 避免使用拼音，除非是通用的专有名词

#### 1.2.5 结构体/接口命名
- 结构体使用驼峰命名，首字母根据访问性决定大小写
- 接口名通常以`er`结尾，例如`Service`、`Storage`、`Handler`、`Logger`
- 避免过于宽泛的命名，例如`Manager`、`Util`

### 1.3 注释规范
- 所有公共的结构体、接口、函数、常量必须有注释
- 注释使用完整的句子，首字母大写，结尾加句号
- 注释要简洁明了，说明功能、参数、返回值、注意事项等
- 对于复杂的算法和业务逻辑，需要添加注释说明设计思路
- 代码修改时同步更新对应的注释，避免注释和代码不一致
- 不要添加无用的注释，例如`i++ // i加1`这种不言自明的代码不需要注释

#### 示例
```go
// APIKey 表示API密钥信息
type APIKey struct {
    ID        string    `json:"id"`         // API Key唯一标识
    UserID    string    `json:"user_id"`    // 所属用户ID
    Key       string    `json:"key"`        // 密钥明文，创建时返回
    Status    string    `json:"status"`     // 状态：enabled/disabled
    CreatedAt time.Time `json:"created_at"` // 创建时间
}

// CreateAPIKey 创建新的API Key
// 参数:
//   ctx: 上下文，包含trace_id等信息
//   req: 创建API Key的请求参数
// 返回值:
//   *APIKey: 创建成功的API Key信息
//   error: 创建失败时返回的错误信息
func (s *APIKeyService) CreateAPIKey(ctx context.Context, req *CreateAPIKeyRequest) (*APIKey, error) {
    // 实现逻辑
}
```

## 2. 错误处理规范

### 2.1 错误定义
- 使用`pkg/errors`包定义统一的错误类型
- 错误码统一，避免重复定义
- 错误信息要清晰，包含足够的上下文信息，便于排查问题
- 不要返回裸的`error`，应该使用`errors.Wrap`包裹，添加上下文信息

#### 示例
```go
import "github.com/zenglw/llm-gateway/pkg/errors"

var (
    ErrAPIKeyNotFound = errors.NewNotFoundError("api_key_not_found", "API Key不存在")
    ErrAPIKeyExpired  = errors.NewInvalidRequestError("api_key_expired", "API Key已过期")
)

// 错误包裹示例
apikey, err := s.store.GetAPIKeyByKey(ctx, key)
if err != nil {
    return nil, errors.Wrap(ErrAPIKeyNotFound, fmt.Sprintf("get api key %s failed", key), err)
}
```

### 2.2 错误处理原则
- 不要忽略错误，除非你明确知道后果
- 尽早处理错误，避免错误传递到上层后丢失上下文
- 对于可恢复的错误，可以重试或者降级处理
- 对于不可恢复的错误，直接返回，不要继续执行
- 不要使用`panic`处理正常业务逻辑的错误，`panic`仅用于严重的不可恢复的错误，例如初始化失败

#### 反例
```go
// 错误：忽略错误
apikey, _ := s.store.GetAPIKeyByKey(ctx, key)

// 错误：过多的错误嵌套
if err != nil {
    if err == io.EOF {
        // 处理
    } else {
        return err
    }
}
```

#### 正例
```go
apikey, err := s.store.GetAPIKeyByKey(ctx, key)
if err != nil {
    if errors.Is(err, ErrNotFound) {
        return nil, ErrAPIKeyNotFound
    }
    return nil, errors.Wrap(err, "get api key failed")
}
```

## 3. 函数/方法规范

### 3.1 函数参数
- 参数数量尽量不超过5个，太多时考虑使用结构体传递
- 上下文`context.Context`作为第一个参数
- 参数名要有意义，避免使用单个字母除非是通用的缩写
- 尽量传递值而不是指针，除非结构体很大或者需要修改内容
- 避免使用布尔类型的参数，导致调用者不明白含义，考虑使用枚举类型

#### 反例
```go
// 布尔参数不清晰
func ProcessRequest(req *Request, isDebug bool) error
```

#### 正例
```go
type ProcessMode int
const (
    ProcessModeNormal ProcessMode = iota
    ProcessModeDebug
)

func ProcessRequest(req *Request, mode ProcessMode) error
```

### 3.2 返回值
- 通常返回`(结果, error)`的形式，`error`作为最后一个返回值
- 如果结果为空或者不需要，直接返回`error`即可
- 不要返回`(nil, nil)`的情况，容易导致空指针问题
- 多个返回值时，如果返回的结果是结构体，尽量返回指针而不是值，减少拷贝

### 3.3 函数长度
- 函数尽量保持简短，建议不超过100行
- 复杂的逻辑拆分成多个小函数，每个函数只做一件事
- 避免出现"面条函数"，一层一层嵌套很深

### 3.4 方法接收器
- 接收器名尽量简短，通常用1-2个字母，例如结构体是`APIKeyService`，接收器可以用`s`
- 如果方法需要修改结构体的内容，必须使用指针接收器
- 如果结构体很大，即使不需要修改，也建议使用指针接收器，减少拷贝
- 同一个结构体的方法，接收器类型要保持一致，不要一会值一会指针

## 4. 并发安全规范

### 4.1 并发访问
- 明确哪些结构体是并发安全的，哪些不是
- 并发访问共享资源必须加锁，使用`sync.Mutex`或者读写锁`sync.RWMutex`
- 避免死锁，注意加锁顺序，不要在持有锁的时候调用可能阻塞的函数
- 锁的粒度尽量小，减少锁的持有时间，提高并发性能

### 4.2 Channel 使用
- 明确Channel的方向，只读`<-chan`或者只写`chan<-`，提高代码安全性
- 发送方负责关闭Channel，不要在接收方关闭
- 使用`select`处理多个Channel操作，避免阻塞
- 避免使用无缓冲的Channel，除非明确知道需要同步语义
- 一定要处理Channel关闭的情况，避免panic

### 4.3 Goroutine 使用
- 不要随意启动Goroutine，必须有明确的退出机制
- 使用`context.Context`来控制Goroutine的生命周期
- 长时间运行的Goroutine必须添加监控，防止泄漏
- 避免Goroutine泄漏，确保所有Goroutine都能正常退出
- 不要在循环中无限制启动Goroutine，使用协程池控制并发数

#### 示例
```go
func (s *Service) Start(ctx context.Context) error {
    go func() {
        ticker := time.NewTicker(time.Minute)
        defer ticker.Stop()

        for {
            select {
            case <-ctx.Done():
                logger.Info("service stopped")
                return
            case <-ticker.C:
                s.doBackgroundTask()
            }
        }
    }()
    return nil
}
```

## 5. 性能优化规范

### 5.1 内存管理
- 尽量减少内存分配，复用对象，使用`sync.Pool`缓存常用的对象
- 避免在循环中进行大量内存分配
- 对于大的切片，提前预估容量，避免多次扩容
- 注意内存逃逸，尽量在栈上分配内存，减少堆分配

#### 示例
```go
// 不好：切片会多次扩容
var result []string
for i := 0; i < 1000; i++ {
    result = append(result, strconv.Itoa(i))
}

// 好：提前分配容量
result := make([]string, 0, 1000)
for i := 0; i < 1000; i++ {
    result = append(result, strconv.Itoa(i))
}
```

### 5.2 字符串处理
- 大量字符串拼接使用`strings.Builder`，不要用`+=`
- 避免频繁的字符串和字节数组转换
- 使用正则表达式时，预编译正则，不要每次使用都编译

### 5.3 循环优化
- 避免在循环中做重复的计算，例如获取长度、计算hash等
- 避免在循环中调用高开销的函数
- 对于大循环，考虑使用并发加速，但要注意并发安全

### 5.4 I/O 优化
- 批量读写，减少系统调用次数
- 使用缓冲I/O，减少磁盘/网络IO次数
- 异步处理非关键路径的操作，不阻塞主流程

## 6. 安全规范

### 6.1 输入验证
- 所有外部输入都必须验证合法性，包括HTTP请求参数、配置项等
- 验证参数的类型、长度、格式、取值范围
- 防止注入攻击，例如SQL注入、命令注入、XSS等
- 对于JSON等格式的输入，限制最大大小，防止内存溢出

### 6.2 敏感信息处理
- 不要在日志中打印敏感信息，例如API Key、密码、用户隐私数据等
- 敏感信息在内存中使用后尽快清零，避免被dump
- 配置文件中的敏感信息不要明文存储，使用环境变量或者密钥管理服务
- 传输敏感信息使用HTTPS加密

### 6.3 权限控制
- 所有API接口都需要做权限验证
- 最小权限原则，每个API Key只授予必要的权限
- 管理员接口需要特殊的认证，不要暴露在公网

### 6.4 依赖安全
- 定期检查依赖库的安全漏洞，及时升级
- 不要使用来源不明的依赖库
- 对于间接依赖也要关注安全问题

## 7. 最佳实践

### 7.1 空值处理
- 不要返回 nil 切片，返回空切片`[]T{}`，避免调用方判断 nil 时出错
- 字符串空值判断用`str == ""`，不要用`len(str) == 0`
- 对于指针类型，使用前必须判断是否为 nil，避免空指针panic

### 7.2 枚举值
- 使用自定义类型定义枚举，不要直接用整数或字符串
- 定义所有可能的枚举常量，避免魔术数字

#### 示例
```go
type APIKeyStatus string
const (
    APIKeyStatusEnabled  APIKeyStatus = "enabled"
    APIKeyStatusDisabled APIKeyStatus = "disabled"
)
```

### 7.3 配置处理
- 配置项要有合理的默认值，避免必须配置的情况
- 配置加载后要做合法性校验
- 不要在代码中硬编码配置，所有可配置的参数都放到配置文件中

### 7.4 日志规范
- 使用`pkg/logger`包打印日志，不要直接使用`fmt`打印
- 日志级别合理使用：
  - Debug：调试信息，开发阶段使用，生产环境关闭
  - Info：正常运行的关键信息，例如服务启动、请求完成等
  - Warn：警告信息，不影响正常运行，但需要关注
  - Error：错误信息，需要处理的问题
  - Fatal：致命错误，服务无法继续运行，打印后退出
- 日志包含足够的上下文信息，例如`trace_id`、`user_id`、`model`等，便于排查问题
- 不要打印敏感信息

### 7.5 代码复用
- 避免重复代码，相同的逻辑抽取成公共函数
- 不要过度封装，保持简单，避免过度设计
- 优先使用标准库，不要重复造轮子，除非标准库不能满足需求

## 8. 工具使用

### 8.1 代码检查
所有代码必须通过`golangci-lint`检查，我们使用的主要检查规则包括：
- `errcheck`：检查错误是否被处理
- `unused`：检查未使用的变量、函数、常量
- `gosimple`：简化代码
- `govet`：静态分析，检查常见错误
- `gofmt`：检查代码格式
- `ineffassign`：检查无效的赋值
- `misspell`：检查拼写错误
- `staticcheck`：静态代码分析

### 8.2 CI 检查
提交代码后CI会自动运行以下检查，全部通过才能合并：
- 代码格式检查
- 静态代码分析
- 单元测试
- 集成测试
- 代码覆盖率检查

## 总结
遵循这些规范可以保证项目代码的一致性、可读性和可维护性。如果你有更好的建议或者发现规范有不合理的地方，欢迎提出Issue讨论改进。
