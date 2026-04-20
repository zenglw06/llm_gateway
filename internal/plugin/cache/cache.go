package cache

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/zenglw/llm_gateway/internal/model"
	"github.com/zenglw/llm_gateway/pkg/config"
	"github.com/zenglw/llm_gateway/pkg/logger"
)

// Cache 缓存接口
type Cache interface {
	Get(ctx context.Context, key string) ([]byte, bool)
	Set(ctx context.Context, key string, value []byte, ttl time.Duration)
	Delete(ctx context.Context, key string)
}

// MemoryCache 内存缓存实现
type MemoryCache struct {
	cache map[string][]byte
	ttl   map[string]time.Time
	mu    sync.RWMutex
}

func NewMemoryCache() *MemoryCache {
	mc := &MemoryCache{
		cache: make(map[string][]byte),
		ttl:   make(map[string]time.Time),
	}
	// 启动清理过期缓存的协程
	go mc.cleanup()
	return mc
}

func (mc *MemoryCache) Get(ctx context.Context, key string) ([]byte, bool) {
	mc.mu.RLock()
	defer mc.mu.RUnlock()

	value, ok := mc.cache[key]
	if !ok {
		return nil, false
	}

	// 检查是否过期
	if time.Now().After(mc.ttl[key]) {
		return nil, false
	}

	return value, true
}

func (mc *MemoryCache) Set(ctx context.Context, key string, value []byte, ttl time.Duration) {
	mc.mu.Lock()
	defer mc.mu.Unlock()

	mc.cache[key] = value
	mc.ttl[key] = time.Now().Add(ttl)
}

func (mc *MemoryCache) Delete(ctx context.Context, key string) {
	mc.mu.Lock()
	defer mc.mu.Unlock()

	delete(mc.cache, key)
	delete(mc.ttl, key)
}

// cleanup 定期清理过期缓存
func (mc *MemoryCache) cleanup() {
	ticker := time.NewTicker(1 * time.Hour)
	defer ticker.Stop()

	for range ticker.C {
		mc.mu.Lock()
		now := time.Now()
		for key, expireTime := range mc.ttl {
			if now.After(expireTime) {
				delete(mc.cache, key)
				delete(mc.ttl, key)
			}
		}
		mc.mu.Unlock()
		logger.Debugf("Cleaned up expired cache entries")
	}
}

// RedisCache Redis缓存实现
type RedisCache struct {
	client *redis.Client
	prefix string
}

func NewRedisCache(cfg config.RedisConfig, prefix string) *RedisCache {
	client := redis.NewClient(&redis.Options{
		Addr:     cfg.Addr,
		Password: cfg.Password,
		DB:       cfg.DB,
	})
	return &RedisCache{
		client: client,
		prefix: prefix,
	}
}

func (rc *RedisCache) getKey(key string) string {
	return fmt.Sprintf("%s:%s", rc.prefix, key)
}

func (rc *RedisCache) Get(ctx context.Context, key string) ([]byte, bool) {
	value, err := rc.client.Get(ctx, rc.getKey(key)).Bytes()
	if err == redis.Nil {
		return nil, false
	} else if err != nil {
		logger.Errorf("Redis cache get failed: %v", err)
		return nil, false
	}
	return value, true
}

func (rc *RedisCache) Set(ctx context.Context, key string, value []byte, ttl time.Duration) {
	err := rc.client.Set(ctx, rc.getKey(key), value, ttl).Err()
	if err != nil {
		logger.Errorf("Redis cache set failed: %v", err)
	}
}

func (rc *RedisCache) Delete(ctx context.Context, key string) {
	err := rc.client.Del(ctx, rc.getKey(key)).Err()
	if err != nil {
		logger.Errorf("Redis cache delete failed: %v", err)
	}
}

// Plugin 缓存插件
type Plugin struct {
	config        config.CachePluginConfig
	cache         Cache
	enabled       bool
	modelSkip     map[string]bool // 不需要缓存的模型
	storageConfig config.StorageConfig
}

// NewPlugin 创建缓存插件
func NewPlugin(storageConfig config.StorageConfig) *Plugin {
	p := &Plugin{
		modelSkip: make(map[string]bool),
	}
	// 存储配置需要在Init时传入，这里先暂存
	p.storageConfig = storageConfig
	return p
}

// Name 返回插件名称
func (p *Plugin) Name() string {
	return "cache"
}

// Init 初始化插件
func (p *Plugin) Init(cfg map[string]interface{}) error {
	// 解析配置
	var pluginConfig config.CachePluginConfig
	if enabled, ok := cfg["enabled"].(bool); ok {
		pluginConfig.Enabled = enabled
	}
	if ttl, ok := cfg["ttl"].(int); ok {
		pluginConfig.TTL = ttl
	}
	if maxSize, ok := cfg["max_size"].(int); ok {
		pluginConfig.MaxSize = maxSize
	}
	if cacheType, ok := cfg["type"].(string); ok {
		pluginConfig.Type = cacheType
	}
	if prefix, ok := cfg["prefix"].(string); ok {
		pluginConfig.Prefix = prefix
	}
	if modelSkip, ok := cfg["model_skip"].([]string); ok {
		for _, model := range modelSkip {
			p.modelSkip[model] = true
		}
	}

	// 配置默认值
	if pluginConfig.TTL <= 0 {
		pluginConfig.TTL = 3600 // 默认1小时
	}
	if pluginConfig.MaxSize <= 0 {
		pluginConfig.MaxSize = 10000 // 默认1万条
	}
	if pluginConfig.Type == "" {
		pluginConfig.Type = "memory" // 默认内存缓存
	}
	if pluginConfig.Prefix == "" {
		pluginConfig.Prefix = "llm_gateway:cache"
	}

	p.config = pluginConfig
	p.enabled = pluginConfig.Enabled

	if !p.enabled {
		logger.Infof("Cache plugin is disabled")
		return nil
	}

	// 初始化缓存
	switch pluginConfig.Type {
	case "redis":
		p.cache = NewRedisCache(p.storageConfig.Redis, pluginConfig.Prefix)
	case "memory":
		fallthrough
	default:
		p.cache = NewMemoryCache()
	}

	logger.Infof("Cache plugin initialized, type: %s, ttl: %ds", pluginConfig.Type, pluginConfig.TTL)
	return nil
}

// Close 关闭插件
func (p *Plugin) Close() error {
	return nil
}

// generateCacheKey 生成缓存Key
func (p *Plugin) generateCacheKey(req interface{}) string {
	// 序列化请求参数
	data, err := json.Marshal(req)
	if err != nil {
		logger.Errorf("Failed to marshal request for cache key: %v", err)
		return ""
	}

	// 计算SHA256哈希
	hash := sha256.Sum256(data)
	return hex.EncodeToString(hash[:])
}

// HandleRequest 处理请求
func (p *Plugin) HandleRequest(ctx context.Context, req *model.LLMRequest) (*model.LLMRequest, error) {
	if !p.enabled || req.Stream || p.modelSkip[req.Model] {
		return req, nil
	}

	// 检查是否强制刷新缓存
	if refresh, ok := ctx.Value("X-Cache-Refresh").(bool); ok && refresh {
		return req, nil
	}

	// 生成缓存Key
	cacheKey := p.generateCacheKey(req)
	if cacheKey == "" {
		return req, nil
	}

	// 尝试从缓存获取
	cachedData, ok := p.cache.Get(ctx, cacheKey)
	if !ok {
		// 缓存未命中，将缓存Key存入上下文，供响应处理使用
		ctx = context.WithValue(ctx, "cache_key", cacheKey)
		return req, nil
	}

	// 缓存命中，解析缓存数据
	var cachedResp model.LLMResponse
	if err := json.Unmarshal(cachedData, &cachedResp); err != nil {
		logger.Errorf("Failed to unmarshal cached response: %v", err)
		return req, nil
	}

	// 将缓存响应存入上下文，直接返回，不需要调用LLM
	ctx = context.WithValue(ctx, "cached_response", &cachedResp)
	logger.Debugf("Cache hit for model: %s", req.Model)
	return req, nil
}

// HandleResponse 处理响应
func (p *Plugin) HandleResponse(ctx context.Context, resp *model.LLMResponse) (*model.LLMResponse, error) {
	if !p.enabled {
		return resp, nil
	}

	// 检查是否有缓存Key需要存储
	cacheKey, ok := ctx.Value("cache_key").(string)
	if !ok || cacheKey == "" {
		return resp, nil
	}

	// 序列化响应
	data, err := json.Marshal(resp)
	if err != nil {
		logger.Errorf("Failed to marshal response for cache: %v", err)
		return resp, nil
	}

	// 存储到缓存
	ttl := time.Duration(p.config.TTL) * time.Second
	p.cache.Set(ctx, cacheKey, data, ttl)
	logger.Debugf("Cached response for key: %s, ttl: %v", cacheKey, ttl)
	return resp, nil
}
