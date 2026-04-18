package llm

import (
    "context"

    "github.com/zenglw/llm_gateway/internal/model"
)

// Service LLM服务接口
type Service interface {
    // Name 返回服务名称
    Name() string
    // SupportsModel 是否支持指定模型
    SupportsModel(modelName string) bool
    // ChatCompletion 聊天补全接口
    ChatCompletion(ctx context.Context, req *model.ChatRequest) (*model.ChatResponse, error)
    // ChatCompletionStream 流式聊天补全接口
    ChatCompletionStream(ctx context.Context, req *model.ChatRequest) (<-chan *model.StreamResponse, error)
    // Completion 文本补全接口
    Completion(ctx context.Context, req *model.CompletionRequest) (*model.CompletionResponse, error)
    // CompletionStream 流式文本补全接口
    CompletionStream(ctx context.Context, req *model.CompletionRequest) (<-chan *model.StreamResponse, error)
}

