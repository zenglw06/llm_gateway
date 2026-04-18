package ratelimit

import (
	"context"
	"sort"
	"sync"
	"time"

	"github.com/zenglw/llm_gateway/internal/model"
	"github.com/zenglw/llm_gateway/internal/plugin/ratelimit/strategies"
	"github.com/zenglw/llm_gateway/pkg/config"
	"github.com/zenglw/llm_gateway/pkg/errors"
	"github.com/zenglw/llm_gateway/pkg/logger"
)

// LimitKey 限流key，用于唯一标识一个限流器
type LimitKey struct {
	UserID string
	Model  string
}

// String 返回key的字符串表示
func (k LimitKey) String() string {
	return k.UserID + ":" + k.Model
}

// Plugin 限流插件
type Plugin struct {
	config     config.RateLimitPluginConfig
	strategies map[string]strategies.RateLimitStrategy // 策略注册表
	limiters   sync.Map                     // key: LimitKey.String() -> *limiterEntry
	mu         sync.RWMutex                 // 配置更新锁
	lastClean  time.Time                    // 上次清理时间
}

// limiterEntry 限流器条目，包含限流器和对应的规则
type limiterEntry struct {
	limiter strategies.RateLimiter
	rule    config.LimitRule
	lastUse time.Time
}

// NewPlugin 创建限流插件
func NewPlugin() *Plugin {
	p := &Plugin{
		strategies: make(map[string]strategies.RateLimitStrategy),
		lastClean:  time.Now(),
	}

	// 注册默认策略
	p.registerStrategy(strategies.NewTokenBucketStrategy())
	p.registerStrategy(strategies.NewFixedWindowStrategy())
	p.registerStrategy(strategies.NewSlidingWindowStrategy())

	return p
}

// registerStrategy 注册限流策略
func (p *Plugin) registerStrategy(strategy strategies.RateLimitStrategy) {
	p.strategies[strategy.Type()] = strategy
}

// Name 返回插件名称
func (p *Plugin) Name() string {
	return "ratelimit"
}

// Init 初始化插件
func (p *Plugin) Init(cfg map[string]interface{}) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	// 解析配置
	var pluginConfig config.RateLimitPluginConfig
	// 这里应该用mapstructure解析，但为了兼容旧配置，我们手动处理
	if rate, ok := cfg["rate"].(int); ok {
		pluginConfig.Rate = rate
	}
	if burst, ok := cfg["burst"].(int); ok {
		pluginConfig.Burst = burst
	}
	if enabled, ok := cfg["enabled"].(bool); ok {
		pluginConfig.Enabled = enabled
	}

	// 兼容旧配置，转换为默认规则
	if pluginConfig.Rate > 0 {
		pluginConfig.Default = config.LimitRule{
			Strategy:  "token_bucket",
			Limit:     int64(pluginConfig.Rate),
			LimitType: "request",
			Period:    time.Second,
			Burst:     pluginConfig.Burst,
			Priority:  0,
		}
	} else {
		// 默认配置
		pluginConfig.Default = config.LimitRule{
			Strategy:  "token_bucket",
			Limit:     100,
			LimitType: "request",
			Period:    time.Second,
			Burst:     200,
			Priority:  0,
		}
	}

	p.config = pluginConfig

	logger.Infof("Rate limit plugin initialized, default strategy: %s, limit: %d/%s",
		p.config.Default.Strategy, p.config.Default.Limit, p.config.Default.Period)
	return nil
}

// Close 关闭插件
func (p *Plugin) Close() error {
	// 清理所有限流器
	p.limiters.Range(func(key, value interface{}) bool {
		p.limiters.Delete(key)
		return true
	})
	return nil
}

// HandleRequest 处理请求限流
func (p *Plugin) HandleRequest(ctx context.Context, req *model.LLMRequest) (*model.LLMRequest, error) {
	if !p.config.Enabled {
		return req, nil
	}

	// 构造限流key
	key := LimitKey{
		UserID: req.UserID,
		Model:  req.Model,
	}
	if key.UserID == "" {
		key.UserID = req.APIKey
	}
	if key.UserID == "" {
		key.UserID = "default"
	}
	if key.Model == "" {
		key.Model = "*"
	}

	// 匹配限流规则
	rule := p.matchRule(key.UserID, key.Model)
	if rule.Limit <= 0 {
		// 限流为0，直接拒绝
		return nil, errors.New(errors.ErrCodeTooManyRequests, "too many requests")
	}

	// 获取或创建限流器
	limiter := p.getOrCreateLimiter(key, rule)

	// 检查是否允许请求
	// TODO: 支持根据LimitType计算不同的n值（如token数、流量大小）
	if !limiter.Allow() {
		return nil, errors.New(errors.ErrCodeTooManyRequests, "too many requests")
	}

	// 定期清理过期的限流器
	p.cleanup()

	return req, nil
}

// matchRule 匹配最适合的限流规则
func (p *Plugin) matchRule(userID, model string) config.LimitRule {
	p.mu.RLock()
	defer p.mu.RUnlock()

	var matchedRules []config.LimitRule

	// 遍历所有规则，找到匹配的
	for _, rule := range p.config.Rules {
		userMatch := rule.UserID == "*" || rule.UserID == userID
		modelMatch := rule.Model == "*" || rule.Model == model
		if userMatch && modelMatch {
			matchedRules = append(matchedRules, rule)
		}
	}

	if len(matchedRules) == 0 {
		return p.config.Default
	}

	// 按优先级排序，优先级高的在前
	sort.Slice(matchedRules, func(i, j int) bool {
		return matchedRules[i].Priority > matchedRules[j].Priority
	})

	return matchedRules[0]
}

// getOrCreateLimiter 获取或创建限流器
func (p *Plugin) getOrCreateLimiter(key LimitKey, rule config.LimitRule) strategies.RateLimiter {
	// 先尝试获取现有条目
	if entry, ok := p.limiters.Load(key.String()); ok {
		le := entry.(*limiterEntry)
		// 如果规则没有变化，更新最后使用时间并返回
		if le.rule.ID == rule.ID && le.rule.Strategy == rule.Strategy &&
			le.rule.Limit == rule.Limit && le.rule.Period == rule.Period {
			le.lastUse = time.Now()
			return le.limiter
		}
		// 规则有变化，删除旧的限流器
		p.limiters.Delete(key.String())
	}

	// 创建新的限流器
	strategy, ok := p.strategies[rule.Strategy]
	if !ok {
		// 策略不存在，使用默认策略
		strategy = p.strategies["token_bucket"]
		logger.Warnf("Unknown rate limit strategy: %s, using token_bucket instead", rule.Strategy)
	}

	limiter := strategy.NewLimiter(rule)

	// 保存到缓存
	p.limiters.Store(key.String(), &limiterEntry{
		limiter: limiter,
		rule:    rule,
		lastUse: time.Now(),
	})

	return limiter
}

// cleanup 清理长时间未使用的限流器
func (p *Plugin) cleanup() {
	if time.Since(p.lastClean) < time.Hour {
		return
	}

	p.mu.Lock()
	defer p.mu.Unlock()

	// 再次检查，避免多个goroutine同时清理
	if time.Since(p.lastClean) < time.Hour {
		return
	}

	// 清理超过1小时未使用的限流器
	now := time.Now()
	p.limiters.Range(func(key, value interface{}) bool {
		entry := value.(*limiterEntry)
		if now.Sub(entry.lastUse) > time.Hour {
			p.limiters.Delete(key)
		}
		return true
	})

	p.lastClean = now
	logger.Debugf("Cleaned up rate limiters, remaining: %d", p.countLimiters())
}

// countLimiters 统计当前限流器数量
func (p *Plugin) countLimiters() int {
	count := 0
	p.limiters.Range(func(_, _ interface{}) bool {
		count++
		return true
	})
	return count
}

// UpdateConfig 更新配置
func (p *Plugin) UpdateConfig(newConfig config.RateLimitPluginConfig) {
	p.mu.Lock()
	defer p.mu.Unlock()

	p.config = newConfig

	// 配置更新后，所有限流器需要重新创建
	// 这里不直接删除，而是在下次访问时自动重建，避免影响正在处理的请求
	logger.Info("Rate limit config updated, all limiters will be recreated on next access")
}
