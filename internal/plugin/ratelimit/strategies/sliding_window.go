package strategies

import (
	"sync"
	"time"

	"github.com/zenglw/llm_gateway/pkg/config"
)

// SlidingWindowStrategy 滑动窗口策略实现
type SlidingWindowStrategy struct {
	buckets int // 子窗口数量，默认10个
}

// NewSlidingWindowStrategy 创建滑动窗口策略实例
func NewSlidingWindowStrategy() *SlidingWindowStrategy {
	return &SlidingWindowStrategy{
		buckets: 10, // 默认将窗口分为10个子窗口
	}
}

// Type 返回策略类型
func (s *SlidingWindowStrategy) Type() string {
	return "sliding_window"
}

// NewLimiter 创建滑动窗口限流器
func (s *SlidingWindowStrategy) NewLimiter(rule config.LimitRule) RateLimiter {
	bucketDuration := rule.Period / time.Duration(s.buckets)
	return &SlidingWindowLimiter{
		limit:          rule.Limit,
		period:         rule.Period,
		bucketDuration: bucketDuration,
		buckets:        make([]int64, s.buckets),
		lastUpdate:     time.Now(),
		currentBucket:  0,
	}
}

// SlidingWindowLimiter 滑动窗口限流器实现
type SlidingWindowLimiter struct {
	limit          int64         // 窗口内最大请求数
	period         time.Duration // 总窗口大小
	bucketDuration time.Duration // 子窗口大小
	buckets        []int64       // 子窗口计数数组
	lastUpdate     time.Time     // 最后更新时间
	currentBucket  int           // 当前子窗口索引
	mu             sync.Mutex    // 互斥锁
}

// Allow 检查是否允许1个请求
func (l *SlidingWindowLimiter) Allow() bool {
	return l.AllowN(1)
}

// AllowN 检查是否允许n个请求
func (l *SlidingWindowLimiter) AllowN(n int) bool {
	l.mu.Lock()
	defer l.mu.Unlock()

	now := time.Now()
	l.updateBuckets(now)

	// 计算总请求数
	var total int64
	for _, count := range l.buckets {
		total += count
	}

	// 检查是否超过限制
	if total+int64(n) > l.limit {
		return false
	}

	// 增加当前窗口计数
	l.buckets[l.currentBucket] += int64(n)
	return true
}

// updateBuckets 更新子窗口
func (l *SlidingWindowLimiter) updateBuckets(now time.Time) {
	// 计算时间差
	diff := now.Sub(l.lastUpdate)
	if diff <= 0 {
		return
	}

	// 计算需要滑动的窗口数
	slots := int(diff / l.bucketDuration)
	if slots <= 0 {
		return
	}

	// 滑动窗口，重置过期的子窗口
	for i := 0; i < slots; i++ {
		l.currentBucket = (l.currentBucket + 1) % len(l.buckets)
		l.buckets[l.currentBucket] = 0
	}

	// 更新最后更新时间
	l.lastUpdate = l.lastUpdate.Add(time.Duration(slots) * l.bucketDuration)
}

// Reset 重置限流器
func (l *SlidingWindowLimiter) Reset() {
	l.mu.Lock()
	defer l.mu.Unlock()
	for i := range l.buckets {
		l.buckets[i] = 0
	}
	l.lastUpdate = time.Now()
	l.currentBucket = 0
}
