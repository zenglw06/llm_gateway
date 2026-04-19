package claude

import (
	"bytes"
	"context"
	"net/http"
	"strings"
	"time"

	"github.com/zenglw/llm_gateway/internal/llm"
	"github.com/zenglw/llm_gateway/internal/model"
)

// Service Claude服务实现
type Service struct {
	apiKey     string
	baseURL    string
	timeout    time.Duration
	maxRetries int
	client     *http.Client
}

// Config Claude服务配置
type Config struct {
	APIKey     string `mapstructure:"api_key"`
	BaseURL    string `mapstructure:"base_url"`
	Timeout    int    `mapstructure:"timeout"`
	MaxRetries int    `mapstructure:"max_retries"`
}

// NewService 创建Claude服务实例
func NewService(cfg *Config) *Service {
	if cfg.BaseURL == "" {
		cfg.BaseURL = "https://api.anthropic.com/v1"
	}
	if cfg.Timeout <= 0 {
		cfg.Timeout = 60 // 默认60秒超时
	}
	if cfg.MaxRetries < 0 {
		cfg.MaxRetries = 2 // 默认重试2次
	}

	return &Service{
		apiKey:     cfg.APIKey,
		baseURL:    cfg.BaseURL,
		timeout:    time.Duration(cfg.Timeout) * time.Second,
		maxRetries: cfg.MaxRetries,
		client: &http.Client{
			Timeout: time.Duration(cfg.Timeout) * time.Second,
		},
	}
}

// Name 返回服务名称
func (s *Service) Name() string {
	return "claude"
}

// SupportsModel 是否支持指定模型
func (s *Service) SupportsModel(modelName string) bool {
	// Claude支持的模型前缀
	supportedPrefixes := []string{
		"claude-3", "claude-2", "claude-instant",
	}
	for _, prefix := range supportedPrefixes {
		if strings.HasPrefix(modelName, prefix) {
			return true
		}
	}
	return false
}

// ChatCompletion 聊天补全接口
func (s *Service) ChatCompletion(ctx context.Context, req *model.ChatRequest) (*model.ChatResponse, error) {
	// Mock返回hello world
	return &model.ChatResponse{
		ID:      "mock-id",
		Object:  "chat.completion",
		Created: time.Now().Unix(),
		Model:   req.Model,
		Choices: []model.ChatChoice{
			{
				Index: 0,
				Message: model.ChatMessage{
					Role:    "assistant",
					Content: "hello world",
				},
				FinishReason: "stop",
			},
		},
		Usage: model.Usage{
			PromptTokens:     10,
			CompletionTokens: 2,
			TotalTokens:      12,
		},
	}, nil
}

// ChatCompletionStream 流式聊天补全接口
func (s *Service) ChatCompletionStream(ctx context.Context, req *model.ChatRequest) (<-chan *model.StreamResponse, error) {
	streamChan := make(chan *model.StreamResponse, 100)

	go func() {
		defer close(streamChan)

		// Mock流式返回hello world
		streamChan <- &model.StreamResponse{
			ID:      "mock-id",
			Object:  "chat.completion.chunk",
			Created: time.Now().Unix(),
			Model:   req.Model,
			Choices: []model.StreamChoice{
				{
					Index: 0,
					Delta: model.ChatMessage{
						Role:    "assistant",
						Content: "hello",
					},
					FinishReason: nil,
				},
			},
		}

		streamChan <- &model.StreamResponse{
			ID:      "mock-id",
			Object:  "chat.completion.chunk",
			Created: time.Now().Unix(),
			Model:   req.Model,
			Choices: []model.StreamChoice{
				{
					Index: 0,
					Delta: model.ChatMessage{
						Content: " world",
					},
					FinishReason: nil,
				},
			},
		}

		finishReason := "stop"
		streamChan <- &model.StreamResponse{
			ID:      "mock-id",
			Object:  "chat.completion.chunk",
			Created: time.Now().Unix(),
			Model:   req.Model,
			Choices: []model.StreamChoice{
				{
					Index:        0,
					Delta:        model.ChatMessage{},
					FinishReason: &finishReason,
				},
			},
		}
	}()

	return streamChan, nil
}

// Completion 文本补全接口（Claude没有专门的文本补全接口，复用聊天接口）
func (s *Service) Completion(ctx context.Context, req *model.CompletionRequest) (*model.CompletionResponse, error) {
	// 转换为聊天请求
	chatReq := &model.ChatRequest{
		Model: req.Model,
		Messages: []model.ChatMessage{
			{
				Role:    "user",
				Content: req.Prompt,
			},
		},
		MaxTokens:   req.MaxTokens,
		Temperature: req.Temperature,
		TopP:        req.TopP,
		Stream:      req.Stream,
		User:        req.User,
	}

	chatResp, err := s.ChatCompletion(ctx, chatReq)
	if err != nil {
		return nil, err
	}

	// 转换为文本补全响应
	completionResp := &model.CompletionResponse{
		ID:      chatResp.ID,
		Object:  "text_completion",
		Created: chatResp.Created,
		Model:   chatResp.Model,
		Choices: []model.CompletionChoice{
			{
				Index:        0,
				Text:         chatResp.Choices[0].Message.Content,
				FinishReason: chatResp.Choices[0].FinishReason,
			},
		},
		Usage: chatResp.Usage,
	}

	return completionResp, nil
}

// CompletionStream 流式文本补全接口
func (s *Service) CompletionStream(ctx context.Context, req *model.CompletionRequest) (<-chan *model.StreamResponse, error) {
	// 转换为聊天请求
	chatReq := &model.ChatRequest{
		Model: req.Model,
		Messages: []model.ChatMessage{
			{
				Role:    "user",
				Content: req.Prompt,
			},
		},
		MaxTokens:   req.MaxTokens,
		Temperature: req.Temperature,
		TopP:        req.TopP,
		Stream:      req.Stream,
		User:        req.User,
	}

	return s.ChatCompletionStream(ctx, chatReq)
}

// scanLines 自定义行扫描函数，处理CRLF换行
func scanLines(data []byte, atEOF bool) (advance int, token []byte, err error) {
	if atEOF && len(data) == 0 {
		return 0, nil, nil
	}
	if i := bytes.IndexByte(data, '\n'); i >= 0 {
		// 有完整的行
		return i + 1, bytes.TrimSuffix(data[:i], []byte("\r")), nil
	}
	// 如果在EOF，返回剩余的数据
	if atEOF {
		return len(data), bytes.TrimSuffix(data, []byte("\r")), nil
	}
	// 需要更多数据
	return 0, nil, nil
}

// Register 注册Claude服务工厂
func Register() {
	llm.Register("claude", func(cfg *llm.Config) (llm.Service, error) {
		return NewService(&Config{
			APIKey:     cfg.APIKey,
			BaseURL:    cfg.BaseURL,
			Timeout:    cfg.Timeout,
			MaxRetries: cfg.MaxRetries,
		}), nil
	})
}
