package llm

import (
    "fmt"
    "sync"
)

// Config LLM服务配置
type Config struct {
    APIKey     string `mapstructure:"api_key"`
    BaseURL    string `mapstructure:"base_url"`
    Timeout    int    `mapstructure:"timeout"`
    MaxRetries int    `mapstructure:"max_retries"`
}

// ServiceFactory 服务工厂函数类型
type ServiceFactory func(cfg *Config) (Service, error)

var (
    factories = make(map[string]ServiceFactory)
    mu        sync.RWMutex
)

// Register 注册LLM服务工厂
func Register(serviceType string, factory ServiceFactory) {
    mu.Lock()
    defer mu.Unlock()
    factories[serviceType] = factory
}

// CreateService 创建LLM服务实例
func CreateService(serviceType string, cfg *Config) (Service, error) {
    mu.RLock()
    factory, ok := factories[serviceType]
    mu.RUnlock()

    if !ok {
        return nil, fmt.Errorf("unsupported LLM service type: %s", serviceType)
    }

    return factory(cfg)
}
