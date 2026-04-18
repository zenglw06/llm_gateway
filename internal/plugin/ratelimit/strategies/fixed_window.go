package strategies

import (
	"sync"
	"time"

	"github.com/zenglw/llm_gateway/pkg/config"
)

// FixedWindowStrategy 固定窗口策略实现
type FixedWindowStrategy struct{}

// NewFixedWindowStrategy 创建固定窗口策略实例
func NewFixedWindowStrategy() *FixedWindowStrategy {
	return &FixedWindowStrategy{}
}

// Type 返回策略类型
func (s *FixedWindowStrategy) Type() string {
	return "fixed_window"
}

// NewLimiter 创建固定窗口限流器
func (s *FixedWindowStrategy) NewLimiter(rule config.LimitRule) RateLimiter {
	return &FixedWindowLimiter{
		limit:  rule.Limit,
		period: rule.Period,
		count:  0,
		start:  time.Now(),
	}
}

// FixedWindowLimiter 固定窗口限流器实现
type FixedWindowLimiter struct {
	limit  int64         // 窗口内最大请求数
	period time.Duration // 窗口大小
	count  int64         // 当前窗口计数
	start  time.Time     // 窗口开始时间
	mu     sync.Mutex    // 互斥锁
}

// Allow 检查是否允许1个请求
func (l *FixedWindowLimiter) Allow() bool {
	return l.AllowN(1)
}

// AllowN 检查是否允许n个请求
func (l *FixedWindowLimiter) AllowN(n int) bool {
	l.mu.Lock()
	defer l.mu.Unlock()

	now := time.Now()

	// 如果当前时间超出窗口，重置窗口
	if now.Sub(l.start) >= l.period {
		l.start = now
		l.count = 0
	}

	// 检查是否超过限制
	if l.count+int64(n) > l.limit {
		return false
	}

	l.count += int64(n)
	return true
}

// Reset 重置限流器
func (l *FixedWindowLimiter) Reset() {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.count = 0
	l.start = time.Now()
}
