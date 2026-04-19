package logging

import (
	"context"
	"encoding/json"
	"time"

	"github.com/zenglw/llm_gateway/internal/model"
	"github.com/zenglw/llm_gateway/pkg/logger"
)

// Plugin 日志插件
type Plugin struct {
	logRequest  bool
	logResponse bool
}

// Config 日志插件配置
type Config struct {
	LogRequest  bool `mapstructure:"log_request"`
	LogResponse bool `mapstructure:"log_response"`
}

// NewPlugin 创建日志插件
func NewPlugin() *Plugin {
	return &Plugin{}
}

// Name 返回插件名称
func (p *Plugin) Name() string {
	return "logging"
}

// Init 初始化插件
func (p *Plugin) Init(config map[string]interface{}) error {
	logRequest, ok := config["log_request"].(bool)
	if ok {
		p.logRequest = logRequest
	} else {
		p.logRequest = true
	}

	logResponse, ok := config["log_response"].(bool)
	if ok {
		p.logResponse = logResponse
	} else {
		p.logResponse = false
	}

	logger.Infof("Logging plugin initialized, log_request: %v, log_response: %v", p.logRequest, p.logResponse)
	return nil
}

// Close 关闭插件
func (p *Plugin) Close() error {
	return nil
}

// HandleRequest 记录请求日志
func (p *Plugin) HandleRequest(ctx context.Context, req *model.LLMRequest) (*model.LLMRequest, error) {
	if !p.logRequest {
		return req, nil
	}

	reqJSON, err := json.Marshal(req)
	if err != nil {
		logger.Warnf("Failed to marshal request: %v", err)
		return req, nil
	}

	logger.Infof("Request received: user_id=%s, model=%s, stream=%v, request=%s",
		req.UserID, req.Model, req.Stream, string(reqJSON))

	// 在上下文中记录请求开始时间
	ctx = context.WithValue(ctx, "request_start_time", time.Now())
	req.Context = ctx

	return req, nil
}

// HandleResponse 记录响应日志
func (p *Plugin) HandleResponse(ctx context.Context, resp *model.LLMResponse) (*model.LLMResponse, error) {
	if !p.logResponse {
		return resp, nil
	}

	respJSON, err := json.Marshal(resp)
	if err != nil {
		logger.Warnf("Failed to marshal response: %v", err)
		return resp, nil
	}

	// 计算耗时
	var duration time.Duration
	if startTime, ok := ctx.Value("request_start_time").(time.Time); ok {
		duration = time.Since(startTime)
	}

	logger.Infof("Response sent: model=%s, total_tokens=%d, duration=%v, response=%s",
		resp.Model, resp.Usage.TotalTokens, duration, string(respJSON))

	return resp, nil
}

// HandleError 记录错误日志
func (p *Plugin) HandleError(ctx context.Context, err error) error {
	// 计算耗时
	var duration time.Duration
	if startTime, ok := ctx.Value("request_start_time").(time.Time); ok {
		duration = time.Since(startTime)
	}

	logger.Errorf("Request failed: duration=%v, error=%v", duration, err)

	return err
}
