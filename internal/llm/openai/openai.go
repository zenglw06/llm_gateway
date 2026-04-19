package openai

import (
	"bytes"
	"context"
	"net/http"
	"strings"
	"time"

	"github.com/zenglw/llm_gateway/internal/llm"
	"github.com/zenglw/llm_gateway/internal/model"
)

// Service OpenAI服务实现
type Service struct {
	apiKey     string
	baseURL    string
	timeout    time.Duration
	maxRetries int
	client     *http.Client
}

// Config OpenAI服务配置
type Config struct {
	APIKey     string `mapstructure:"api_key"`
	BaseURL    string `mapstructure:"base_url"`
	Timeout    int    `mapstructure:"timeout"`
	MaxRetries int    `mapstructure:"max_retries"`
}

// NewService 创建OpenAI服务实例
func NewService(cfg *Config) *Service {
	if cfg.BaseURL == "" {
		cfg.BaseURL = "https://api.openai.com/v1"
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
	return "openai"
}

// SupportsModel 是否支持指定模型
func (s *Service) SupportsModel(modelName string) bool {
	// OpenAI支持的模型前缀
	supportedPrefixes := []string{
		"gpt-3.5-turbo", "gpt-4", "gpt-4o", "text-davinci",
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

// Completion 文本补全接口
func (s *Service) Completion(ctx context.Context, req *model.CompletionRequest) (*model.CompletionResponse, error) {
	// Mock返回hello world
	return &model.CompletionResponse{
		ID:      "mock-id",
		Object:  "text_completion",
		Created: time.Now().Unix(),
		Model:   req.Model,
		Choices: []model.CompletionChoice{
			{
				Index:        0,
				Text:         "hello world",
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

// CompletionStream 流式文本补全接口
func (s *Service) CompletionStream(ctx context.Context, req *model.CompletionRequest) (<-chan *model.StreamResponse, error) {
	streamChan := make(chan *model.StreamResponse, 100)

	go func() {
		defer close(streamChan)

		// Mock流式返回hello world
		streamChan <- &model.StreamResponse{
			ID:      "mock-id",
			Object:  "text_completion.chunk",
			Created: time.Now().Unix(),
			Model:   req.Model,
			Choices: []model.StreamChoice{
				{
					Index: 0,
					Delta: model.ChatMessage{
						Content: "hello",
					},
					FinishReason: nil,
				},
			},
		}

		streamChan <- &model.StreamResponse{
			ID:      "mock-id",
			Object:  "text_completion.chunk",
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
			Object:  "text_completion.chunk",
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

// Register 注册OpenAI服务工厂
func Register() {
	llm.Register("openai", func(cfg *llm.Config) (llm.Service, error) {
		return NewService(&Config{
			APIKey:     cfg.APIKey,
			BaseURL:    cfg.BaseURL,
			Timeout:    cfg.Timeout,
			MaxRetries: cfg.MaxRetries,
		}), nil
	})
}
