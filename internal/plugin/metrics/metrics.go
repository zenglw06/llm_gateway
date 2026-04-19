package metrics

import (
	"context"
	"strconv"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"

	"github.com/zenglw/llm_gateway/internal/model"
	"github.com/zenglw/llm_gateway/pkg/logger"
)

// Plugin Metrics插件
type Plugin struct {
	requestCount   *prometheus.CounterVec
	requestLatency *prometheus.HistogramVec
	tokenUsage     *prometheus.CounterVec
	errorCount     *prometheus.CounterVec
}

// Config Metrics插件配置
type Config struct {
	Enabled bool   `mapstructure:"enabled"`
	Path    string `mapstructure:"path"`
}

// NewPlugin 创建Metrics插件
func NewPlugin() *Plugin {
	return &Plugin{}
}

// Name 返回插件名称
func (p *Plugin) Name() string {
	return "metrics"
}

// Init 初始化插件
func (p *Plugin) Init(config map[string]interface{}) error {
	// 初始化指标
	p.requestCount = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "llm_gateway_requests_total",
		Help: "Total number of requests",
	}, []string{"model", "user_id", "status"})

	p.requestLatency = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "llm_gateway_request_duration_seconds",
		Help:    "Request duration in seconds",
		Buckets: []float64{0.1, 0.25, 0.5, 1, 2.5, 5, 10, 30},
	}, []string{"model", "user_id"})

	p.tokenUsage = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "llm_gateway_tokens_total",
		Help: "Total number of tokens used",
	}, []string{"model", "user_id", "type"})

	p.errorCount = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "llm_gateway_errors_total",
		Help: "Total number of errors",
	}, []string{"model", "user_id", "error_code"})

	logger.Info("Metrics plugin initialized")
	return nil
}

// Close 关闭插件
func (p *Plugin) Close() error {
	return nil
}

// HandleRequest 记录请求指标
func (p *Plugin) HandleRequest(ctx context.Context, req *model.LLMRequest) (*model.LLMRequest, error) {
	// 在上下文中记录请求开始时间
	ctx = context.WithValue(ctx, "request_start_time", time.Now())
	req.Context = ctx

	return req, nil
}

// HandleResponse 记录响应指标
func (p *Plugin) HandleResponse(ctx context.Context, resp *model.LLMResponse) (*model.LLMResponse, error) {
	// 从上下文中获取用户ID和模型
	var userID, modelName string
	if req, ok := ctx.Value("request").(*model.LLMRequest); ok {
		userID = req.UserID
		modelName = req.Model
	} else {
		modelName = resp.Model
	}

	// 记录请求成功计数
	p.requestCount.WithLabelValues(modelName, userID, "200").Inc()

	// 记录请求延迟
	if startTime, ok := ctx.Value("request_start_time").(time.Time); ok {
		duration := time.Since(startTime).Seconds()
		p.requestLatency.WithLabelValues(modelName, userID).Observe(duration)
	}

	// 记录Token使用量
	if resp.Usage.PromptTokens > 0 {
		p.tokenUsage.WithLabelValues(modelName, userID, "prompt").Add(float64(resp.Usage.PromptTokens))
	}
	if resp.Usage.CompletionTokens > 0 {
		p.tokenUsage.WithLabelValues(modelName, userID, "completion").Add(float64(resp.Usage.CompletionTokens))
	}
	if resp.Usage.TotalTokens > 0 {
		p.tokenUsage.WithLabelValues(modelName, userID, "total").Add(float64(resp.Usage.TotalTokens))
	}

	return resp, nil
}

// HandleError 记录错误指标
func (p *Plugin) HandleError(ctx context.Context, err error) error {
	// 从上下文中获取用户ID和模型
	var userID, modelName string
	if req, ok := ctx.Value("request").(*model.LLMRequest); ok {
		userID = req.UserID
		modelName = req.Model
	}

	// 获取错误码
	errorCode := "500"
	if e, ok := err.(interface{ Code() int }); ok {
		errorCode = strconv.Itoa(e.Code())
	}

	// 记录错误计数
	p.errorCount.WithLabelValues(modelName, userID, errorCode).Inc()

	// 记录请求失败计数
	p.requestCount.WithLabelValues(modelName, userID, errorCode).Inc()

	return err
}
