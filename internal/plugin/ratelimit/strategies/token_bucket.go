package strategies

import (
	"time"

	"golang.org/x/time/rate"

	"github.com/zenglw/llm_gateway/pkg/config"
)

// TokenBucketStrategy 令牌桶策略实现
type TokenBucketStrategy struct{}

// NewTokenBucketStrategy 创建令牌桶策略实例
func NewTokenBucketStrategy() *TokenBucketStrategy {
	return &TokenBucketStrategy{}
}

// Type 返回策略类型
func (s *TokenBucketStrategy) Type() string {
	return "token_bucket"
}

// NewLimiter 创建令牌桶限流器
func (s *TokenBucketStrategy) NewLimiter(rule config.LimitRule) RateLimiter {
	var r rate.Limit
	if rule.Period == time.Second {
		r = rate.Limit(rule.Limit)
	} else {
		// 转换为每秒速率
		r = rate.Limit(float64(rule.Limit) / rule.Period.Seconds())
	}

	burst := rule.Burst
	if burst <= 0 {
		burst = int(rule.Limit)
	}

	return &TokenBucketLimiter{
		limiter: rate.NewLimiter(r, burst),
	}
}

// TokenBucketLimiter 令牌桶限流器实现
type TokenBucketLimiter struct {
	limiter *rate.Limiter
}

// Allow 检查是否允许1个请求
func (l *TokenBucketLimiter) Allow() bool {
	return l.limiter.Allow()
}

// AllowN 检查是否允许n个请求
func (l *TokenBucketLimiter) AllowN(n int) bool {
	return l.limiter.AllowN(time.Now(), n)
}

// Reset 重置限流器（令牌桶不支持重置，这里是空实现）
func (l *TokenBucketLimiter) Reset() {
	// 令牌桶是基于时间的，不需要手动重置
}
