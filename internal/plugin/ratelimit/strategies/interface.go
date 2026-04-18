package strategies

import "github.com/zenglw/llm_gateway/pkg/config"

// RateLimiter 限流器接口
type RateLimiter interface {
	Allow() bool       // 检查是否允许1个请求
	AllowN(n int) bool // 检查是否允许n个请求
	Reset()            // 重置限流器状态
}

// RateLimitStrategy 限流策略接口
type RateLimitStrategy interface {
	NewLimiter(rule config.LimitRule) RateLimiter // 创建新的限流器实例
	Type() string                                 // 返回策略类型名称
}
