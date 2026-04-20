package config

import (
	"strings"
	"time"

	"github.com/spf13/viper"
	"github.com/zenglw/llm_gateway/pkg/logger"
)

// Config 全局配置
type Config struct {
	Server  ServerConfig  `mapstructure:"server"`
	Log     logger.Config `mapstructure:"log"`
	LLM     LLMConfig     `mapstructure:"llm"`
	Plugin  PluginConfig  `mapstructure:"plugin"`
	Storage StorageConfig `mapstructure:"storage"`
	Auth    AuthConfig    `mapstructure:"auth"`
}

// ServerConfig 服务配置
type ServerConfig struct {
	Port         int    `mapstructure:"port"`
	Mode         string `mapstructure:"mode"`
	ReadTimeout  int    `mapstructure:"read_timeout"`
	WriteTimeout int    `mapstructure:"write_timeout"`
}

// LLMConfig LLM服务配置
type LLMConfig struct {
	Services map[string]LLMServiceConfig `mapstructure:"services"`
}

// LLMServiceConfig 单个LLM服务配置
type LLMServiceConfig struct {
	Type       string `mapstructure:"type"`
	APIKey     string `mapstructure:"api_key"`
	BaseURL    string `mapstructure:"base_url"`
	Timeout    int    `mapstructure:"timeout"`
	MaxRetries int    `mapstructure:"max_retries"`
	Enabled    bool   `mapstructure:"enabled"`
}

// PluginConfig 插件配置
type PluginConfig struct {
	EnabledPlugins []string              `mapstructure:"enabled"`
	Auth           AuthPluginConfig      `mapstructure:"auth"`
	RateLimit      RateLimitPluginConfig `mapstructure:"ratelimit"`
	Metrics        MetricsPluginConfig   `mapstructure:"metrics"`
	Logging        LoggingPluginConfig   `mapstructure:"logging"`
	Cache          CachePluginConfig     `mapstructure:"cache"`
}

// AuthPluginConfig 鉴权插件配置
type AuthPluginConfig struct {
	Enabled bool `mapstructure:"enabled"`
}

// LimitRule 限流规则
type LimitRule struct {
	ID        string        `json:"id" mapstructure:"id"`
	UserID    string        `json:"user_id" mapstructure:"user_id"`       // 用户ID，*表示匹配所有用户
	Model     string        `json:"model" mapstructure:"model"`           // 模型名称，*表示匹配所有模型
	Strategy  string        `json:"strategy" mapstructure:"strategy"`     // 策略类型：token_bucket, fixed_window, sliding_window
	Limit     int64         `json:"limit" mapstructure:"limit"`           // 限流阈值
	LimitType string        `json:"limit_type" mapstructure:"limit_type"` // 阈值类型：request, token, bandwidth
	Period    time.Duration `json:"period" mapstructure:"period"`         // 时间窗口
	Burst     int           `json:"burst" mapstructure:"burst"`           // 突发请求数（仅令牌桶策略）
	Priority  int           `json:"priority" mapstructure:"priority"`     // 优先级，数值越大优先级越高
}

// RateLimitPluginConfig 限流插件配置
type RateLimitPluginConfig struct {
	Enabled bool        `mapstructure:"enabled"`
	Rate    int         `mapstructure:"rate"`    // 兼容旧配置，默认令牌桶速率
	Burst   int         `mapstructure:"burst"`   // 兼容旧配置，默认令牌桶突发
	Default LimitRule   `mapstructure:"default"` // 默认规则
	Rules   []LimitRule `mapstructure:"rules"`   // 自定义规则列表
}

// MetricsPluginConfig Metrics插件配置
type MetricsPluginConfig struct {
	Enabled bool   `mapstructure:"enabled"`
	Path    string `mapstructure:"path"`
}

// LoggingPluginConfig 日志插件配置
type LoggingPluginConfig struct {
	Enabled     bool `mapstructure:"enabled"`
	LogRequest  bool `mapstructure:"log_request"`
	LogResponse bool `mapstructure:"log_response"`
}

// CachePluginConfig 缓存插件配置
type CachePluginConfig struct {
	Enabled   bool     `mapstructure:"enabled"`
	TTL       int      `mapstructure:"ttl"`       // 缓存过期时间，单位秒，默认3600
	MaxSize   int      `mapstructure:"max_size"`  // 最大缓存条目数，默认10000
	Type      string   `mapstructure:"type"`      // 缓存类型：memory/redis，默认memory
	Prefix    string   `mapstructure:"prefix"`    // 缓存Key前缀，默认llm_gateway:cache
	ModelSkip []string `mapstructure:"model_skip"` // 不需要缓存的模型列表
}

// StorageConfig 存储配置
type StorageConfig struct {
	Type  string      `mapstructure:"type"`
	Redis RedisConfig `mapstructure:"redis"`
}

// RedisConfig Redis配置
type RedisConfig struct {
	Addr     string `mapstructure:"addr"`
	Password string `mapstructure:"password"`
	DB       int    `mapstructure:"db"`
}

// AuthConfig 认证配置
type AuthConfig struct {
	JWTSecret string `mapstructure:"jwt_secret"`
	JWTExpire int    `mapstructure:"jwt_expire"`
}

var (
	globalConfig *Config
	configPath   string
)

// Load 加载配置文件
func Load(path string) (*Config, error) {
	configPath = path
	v := viper.New()

	// 设置配置文件信息
	v.SetConfigFile(configPath)
	v.SetConfigType("yaml")

	// 环境变量支持
	v.SetEnvPrefix("LLM_GATEWAY")
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	v.AutomaticEnv()

	// 读取配置
	if err := v.ReadInConfig(); err != nil {
		return nil, err
	}

	// 解析到结构体
	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, err
	}

	globalConfig = &cfg
	return &cfg, nil
}

// Reload 重新加载配置文件
func Reload() (*Config, error) {
	if configPath == "" {
		return nil, nil
	}
	return Load(configPath)
}

// Get 获取全局配置
func Get() *Config {
	if globalConfig == nil {
		return &Config{}
	}
	return globalConfig
}
