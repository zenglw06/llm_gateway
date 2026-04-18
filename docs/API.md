# API 文档

LLM Gateway完全兼容OpenAI API协议，你可以直接使用任何OpenAI SDK与网关交互，只需要将base_url设置为网关的地址即可。

本文档描述网关提供的所有API接口。

## 公共信息

### 基础URL
```
http://<gateway-host>:<port>/v1
```

### 认证
所有API请求都需要在请求头中携带API Key：
```http
Authorization: Bearer <your-api-key>
```

### 错误响应
所有错误响应都遵循统一格式：
```json
{
  "error": {
    "code": "invalid_request_error",
    "message": "Invalid API key provided",
    "param": "authorization",
    "type": "authentication_error"
  }
}
```

错误码说明：
| 错误码 | HTTP状态码 | 说明 |
|--------|------------|------|
| invalid_request_error | 400 | 请求参数错误 |
| authentication_error | 401 | 认证失败，API Key无效或过期 |
| permission_denied | 403 | 权限不足，没有访问该模型或接口的权限 |
| not_found_error | 404 | 请求的资源不存在 |
| rate_limit_exceeded | 429 | 请求频率超过限制 |
| quota_exceeded | 429 | 配额不足 |
| service_unavailable | 503 | 服务不可用，LLM厂商服务故障 |
| internal_server_error | 500 | 网关内部错误 |

---

## 1. 聊天补全接口

### POST /chat/completions
创建聊天补全请求，支持非流式和流式响应。

#### 请求参数
| 参数名 | 类型 | 必填 | 说明 |
|--------|------|------|------|
| model | string | 是 | 模型名称，例如："gpt-3.5-turbo", "claude-3-sonnet-20240229" |
| messages | array | 是 | 聊天消息列表 |
| messages[].role | string | 是 | 消息角色：system、user、assistant |
| messages[].content | string | 是 | 消息内容 |
| messages[].name | string | 否 | 参与者名称 |
| temperature | float | 否 | 采样温度，0-2之间，默认1，值越大输出越随机 |
| top_p | float | 否 | 核采样参数，0-1之间，默认1 |
| n | int | 否 | 生成的候选结果数量，默认1 |
| stream | bool | 否 | 是否流式响应，默认false |
| stop | string/array | 否 | 停止生成的序列 |
| max_tokens | int | 否 | 最大生成Token数 |
| presence_penalty | float | 否 | 存在惩罚，-2到2之间，默认0 |
| frequency_penalty | float | 否 | 频率惩罚，-2到2之间，默认0 |
| user | string | 否 | 终端用户ID，用于滥用检测 |

#### 请求示例
```bash
curl http://localhost:8080/v1/chat/completions \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer sk-xxxxxxxxxxxxxxxx" \
  -d '{
    "model": "gpt-3.5-turbo",
    "messages": [
      {"role": "system", "content": "你是一个 helpful 的助手。"},
      {"role": "user", "content": "你好！"}
    ],
    "temperature": 0.7
  }'
```

#### 响应示例（非流式）
```json
{
  "id": "chatcmpl-123",
  "object": "chat.completion",
  "created": 1677652288,
  "model": "gpt-3.5-turbo-0613",
  "choices": [
    {
      "index": 0,
      "message": {
        "role": "assistant",
        "content": "你好！有什么我可以帮助你的吗？"
      },
      "finish_reason": "stop"
    }
  ],
  "usage": {
    "prompt_tokens": 20,
    "completion_tokens": 15,
    "total_tokens": 35
  }
}
```

#### 流式响应
当`stream=true`时，响应使用SSE格式，逐块返回结果：

```
data: {"id":"chatcmpl-123","object":"chat.completion.chunk","created":1677652288,"model":"gpt-3.5-turbo-0613","choices":[{"index":0,"delta":{"role":"assistant","content":"你"},"finish_reason":null}]}

data: {"id":"chatcmpl-123","object":"chat.completion.chunk","created":1677652288,"model":"gpt-3.5-turbo-0613","choices":[{"index":0,"delta":{"content":"好"},"finish_reason":null}]}

data: {"id":"chatcmpl-123","object":"chat.completion.chunk","created":1677652288,"model":"gpt-3.5-turbo-0613","choices":[{"index":0,"delta":{"content":"！"},"finish_reason":null}]}

data: [DONE]
```

---

## 2. 文本补全接口

### POST /completions
创建文本补全请求，适用于文本生成、续写等场景。

#### 请求参数
| 参数名 | 类型 | 必填 | 说明 |
|--------|------|------|------|
| model | string | 是 | 模型名称 |
| prompt | string/array | 是 | 提示文本 |
| suffix | string | 否 | 生成文本的后缀 |
| max_tokens | int | 否 | 最大生成Token数，默认16 |
| temperature | float | 否 | 采样温度，默认1 |
| top_p | float | 否 | 核采样参数，默认1 |
| n | int | 否 | 生成结果数量，默认1 |
| stream | bool | 否 | 是否流式响应，默认false |
| logprobs | int | 否 | 返回最可能的n个Token的对数概率 |
| echo | bool | 否 | 是否在响应中回显提示文本，默认false |
| stop | string/array | 否 | 停止序列 |
| presence_penalty | float | 否 | 存在惩罚，默认0 |
| frequency_penalty | float | 否 | 频率惩罚，默认0 |
| best_of | int | 否 | 服务器端生成best_of个结果，返回最好的一个 |
| user | string | 否 | 终端用户ID |

#### 请求示例
```bash
curl http://localhost:8080/v1/completions \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer sk-xxxxxxxxxxxxxxxx" \
  -d '{
    "model": "text-davinci-003",
    "prompt": "从前有座山，",
    "max_tokens": 100,
    "temperature": 0.5
  }'
```

#### 响应示例
```json
{
  "id": "cmpl-123",
  "object": "text_completion",
  "created": 1677652288,
  "model": "text-davinci-003",
  "choices": [
    {
      "text": "山里有座庙，庙里有个老和尚在给小和尚讲故事。",
      "index": 0,
      "logprobs": null,
      "finish_reason": "stop"
    }
  ],
  "usage": {
    "prompt_tokens": 7,
    "completion_tokens": 24,
    "total_tokens": 31
  }
}
```

---

## 3. 模型列表接口

### GET /models
获取所有可用的模型列表。

#### 请求示例
```bash
curl http://localhost:8080/v1/models \
  -H "Authorization: Bearer sk-xxxxxxxxxxxxxxxx"
```

#### 响应示例
```json
{
  "data": [
    {
      "id": "gpt-3.5-turbo",
      "object": "model",
      "created": 1677652288,
      "owned_by": "openai"
    },
    {
      "id": "gpt-4",
      "object": "model",
      "created": 1677652288,
      "owned_by": "openai"
    },
    {
      "id": "claude-3-sonnet-20240229",
      "object": "model",
      "created": 1677652288,
      "owned_by": "anthropic"
    },
    {
      "id": "deepseek-chat",
      "object": "model",
      "created": 1677652288,
      "owned_by": "deepseek"
    }
  ],
  "object": "list"
}
```

---

## 4. API Key 管理接口

### POST /api-keys
创建新的API Key。

#### 请求参数
| 参数名 | 类型 | 必填 | 说明 |
|--------|------|------|------|
| user_id | string | 是 | 所属用户ID |
| name | string | 否 | API Key名称 |
| permissions | array | 否 | 权限列表，例如：["chat.completions", "completions"] |
| allowed_ips | array | 否 | 允许的IP白名单 |
| expires_at | int | 否 | 过期时间，Unix时间戳，0表示永不过期 |
| quota | int | 否 | 总配额，0表示不限制 |
| quota_period | string | 否 | 配额周期："day"、"month"、"total" |

#### 响应示例
```json
{
  "id": "ak-xxxxxxxxxxxxxxxx",
  "user_id": "user123",
  "name": "测试API Key",
  "key": "sk-xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx", // 只会在创建时返回明文
  "status": "enabled",
  "permissions": ["*"],
  "allowed_ips": [],
  "expires_at": 0,
  "quota": 10000,
  "quota_used": 0,
  "quota_period": "month",
  "created_at": 1677652288,
  "updated_at": 1677652288,
  "last_used_at": null
}
```

### GET /api-keys/:id
获取API Key详情。

#### 响应示例
同上，但是不会返回`key`字段。

### GET /api-keys
获取用户的API Key列表。

#### 查询参数
| 参数名 | 类型 | 说明 |
|--------|------|------|
| user_id | string | 用户ID |
| limit | int | 分页大小，默认20 |
| offset | int | 分页偏移，默认0 |

#### 响应示例
```json
{
  "data": [
    {
      "id": "ak-xxxxxxxxxxxxxxxx",
      "user_id": "user123",
      "name": "测试API Key",
      "status": "enabled",
      "permissions": ["*"],
      "expires_at": 0,
      "quota": 10000,
      "quota_used": 1250,
      "created_at": 1677652288,
      "last_used_at": 1677652300
    }
  ],
  "total": 1,
  "limit": 20,
  "offset": 0
}
```

### PUT /api-keys/:id
更新API Key信息。

#### 请求参数
可以更新的字段：`name`、`status`、`permissions`、`allowed_ips`、`expires_at`、`quota`。

#### 响应示例
同GET接口。

### DELETE /api-keys/:id
删除API Key。

#### 响应示例
```json
{
  "status": "success"
}
```

---

## 5. 配额管理接口

### GET /quota/:user_id
获取用户配额使用情况。

#### 响应示例
```json
{
  "user_id": "user123",
  "total_quota": 100000,
  "used_quota": 25600,
  "remaining_quota": 74400,
  "quota_period": "month",
  "reset_time": 1680307200,
  "usage_details": [
    {
      "model": "gpt-3.5-turbo",
      "requests": 1200,
      "tokens": 24000
    },
    {
      "model": "gpt-4",
      "requests": 32,
      "tokens": 1600
    }
  ]
}
```

### PUT /quota/:user_id
更新用户配额。

#### 请求参数
| 参数名 | 类型 | 必填 | 说明 |
|--------|------|------|------|
| total_quota | int | 是 | 总配额 |
| quota_period | string | 是 | 配额周期：day/month/total |

#### 响应示例
同GET接口。

### POST /quota/:user_id/reset
重置用户配额。

#### 响应示例
```json
{
  "status": "success",
  "message": "Quota reset successfully"
}
```

---

## 6. 系统接口

### GET /health
健康检查接口，返回服务状态。

#### 响应示例
```json
{
  "status": "ok",
  "version": "1.0.0",
  "git_commit": "abc123def",
  "build_time": "2024-01-01T12:00:00Z",
  "uptime": 3600,
  "plugins": ["auth", "ratelimit", "logging", "metrics"],
  "llm_services": ["openai", "claude", "deepseek"]
}
```

### GET /metrics
Prometheus Metrics端点，返回监控指标。

### POST /-/reload
重载配置文件（需要管理员权限）。

#### 响应示例
```json
{
  "status": "success",
  "message": "Configuration reloaded successfully"
}
```

---

## SDK 使用示例

### Python
```python
from openai import OpenAI

client = OpenAI(
  api_key="sk-xxxxxxxxxxxxxxxx",
  base_url="http://localhost:8080/v1"
)

# 聊天补全
response = client.chat.completions.create(
  model="gpt-3.5-turbo",
  messages=[
    {"role": "user", "content": "你好！"}
  ]
)
print(response.choices[0].message.content)

# 流式响应
stream = client.chat.completions.create(
  model="gpt-3.5-turbo",
  messages=[{"role": "user", "content": "你好！"}],
  stream=True
)
for chunk in stream:
  if chunk.choices[0].delta.content is not None:
    print(chunk.choices[0].delta.content, end="")
```

### JavaScript
```javascript
import OpenAI from 'openai';

const openai = new OpenAI({
  apiKey: 'sk-xxxxxxxxxxxxxxxx',
  baseURL: 'http://localhost:8080/v1',
});

async function main() {
  const stream = await openai.chat.completions.create({
    model: 'gpt-3.5-turbo',
    messages: [{ role: 'user', content: '你好！' }],
    stream: true,
  });
  for await (const chunk of stream) {
    process.stdout.write(chunk.choices[0]?.delta?.content || '');
  }
}

main();
```

### Java
```java
import com.theokanning.openai.OpenAiService;
import com.theokanning.openai.completion.chat.ChatCompletionRequest;
import com.theokanning.openai.completion.chat.ChatMessage;
import java.util.ArrayList;
import java.util.List;

public class Example {
  public static void main(String[] args) {
    OpenAiService service = new OpenAiService("sk-xxxxxxxxxxxxxxxx", "http://localhost:8080/v1");

    List<ChatMessage> messages = new ArrayList<>();
    messages.add(new ChatMessage("user", "你好！"));

    ChatCompletionRequest request = ChatCompletionRequest.builder()
      .model("gpt-3.5-turbo")
      .messages(messages)
      .build();

    String response = service.createChatCompletion(request)
      .getChoices().get(0).getMessage().getContent();

    System.out.println(response);
  }
}
```

---

## 常见问题

### Q：OpenAI的函数调用、Tool Call等功能支持吗？
A：是的，完全支持，网关会透传所有参数给对应的LLM服务，只要该服务支持对应的功能即可。

### Q：是否支持Embedding、Fine-tuning等其他OpenAI API？
A：目前核心支持聊天和文本补全接口，其他接口可以根据需求扩展，欢迎提交PR。

### Q：如何指定使用某个特定的LLM厂商处理请求？
A：可以通过请求头`X-LLM-Service: openai`来强制指定使用某个服务，也可以通过配置路由规则来实现。

### Q：如何获取请求的实际处理服务和模型？
A：响应头会包含`X-LLM-Service`和`X-LLM-Model`字段，显示实际处理请求的服务和使用的模型。
