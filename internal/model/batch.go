package model

// BatchRequest 批量请求
type BatchRequest struct {
	Model       string          `json:"model,omitempty"`       // 公共模型，子请求可覆盖
	MaxTokens   int             `json:"max_tokens,omitempty"`  // 公共max_tokens，子请求可覆盖
	Temperature float64         `json:"temperature,omitempty"` // 公共temperature，子请求可覆盖
	TopP        float64         `json:"top_p,omitempty"`       // 公共top_p，子请求可覆盖
	User        string          `json:"user,omitempty"`        // 公共user，子请求可覆盖
	Requests    []BatchSubRequest `json:"requests" binding:"required,min=1"` // 子请求列表
}

// BatchSubRequest 批量子请求
type BatchSubRequest struct {
	CustomID    string        `json:"custom_id,omitempty"`    // 自定义ID，用于匹配响应
	Model       string        `json:"model,omitempty"`        // 模型，优先使用
	Messages    []ChatMessage `json:"messages,omitempty"`     // 聊天消息（聊天补全用）
	Prompt      string        `json:"prompt,omitempty"`       // 提示词（文本补全用）
	MaxTokens   int           `json:"max_tokens,omitempty"`   // 最大token数
	Temperature float64       `json:"temperature,omitempty"`  // 温度值
	TopP        float64       `json:"top_p,omitempty"`        // top_p
	User        string        `json:"user,omitempty"`         // 用户标识
}

// BatchResponse 批量响应
type BatchResponse struct {
	ID         string             `json:"id"`
	Object     string             `json:"object"`
	Created    int64              `json:"created"`
	Model      string             `json:"model,omitempty"`
	Responses  []BatchSubResponse `json:"responses"`
}

// BatchSubResponse 批量子响应
type BatchSubResponse struct {
	CustomID   string                `json:"custom_id,omitempty"` // 对应请求的custom_id
	Status     string                `json:"status"`               // 状态：success/failed
	Error      *ErrorResponse        `json:"error,omitempty"`      // 错误信息（失败时）
	Response   *BatchSubResponseData `json:"response,omitempty"`   // 响应数据（成功时）
}

// BatchSubResponseData 批量子响应数据
type BatchSubResponseData struct {
	ID      string      `json:"id"`
	Object  string      `json:"object"`
	Created int64       `json:"created"`
	Model   string      `json:"model"`
	Choices interface{} `json:"choices"`
	Usage   Usage       `json:"usage"`
}

// ErrorResponse 错误响应
type ErrorResponse struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}
