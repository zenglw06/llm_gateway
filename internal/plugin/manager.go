package plugin

import (
    "fmt"
    "sync"

    "github.com/zenglw/llm_gateway/pkg/logger"
)

// DefaultManager 默认插件管理器实现
type DefaultManager struct {
    plugins         map[string]Plugin
    requestPlugins  []RequestPlugin
    responsePlugins []ResponsePlugin
    errorPlugins    []ErrorPlugin
    mu              sync.RWMutex
}

// NewManager 创建新的插件管理器
func NewManager() *DefaultManager {
    return &DefaultManager{
        plugins: make(map[string]Plugin),
    }
}

// Register 注册插件
func (m *DefaultManager) Register(plugin Plugin) error {
    m.mu.Lock()
    defer m.mu.Unlock()

    name := plugin.Name()
    if _, ok := m.plugins[name]; ok {
        return fmt.Errorf("plugin %s already registered", name)
    }

    m.plugins[name] = plugin

    // 按类型分类
    if rp, ok := plugin.(RequestPlugin); ok {
        m.requestPlugins = append(m.requestPlugins, rp)
    }
    if rp, ok := plugin.(ResponsePlugin); ok {
        m.responsePlugins = append(m.responsePlugins, rp)
    }
    if ep, ok := plugin.(ErrorPlugin); ok {
        m.errorPlugins = append(m.errorPlugins, ep)
    }

    logger.Infof("Plugin %s registered", name)
    return nil
}

// GetPlugin 获取插件
func (m *DefaultManager) GetPlugin(name string) (Plugin, bool) {
    m.mu.RLock()
    defer m.mu.RUnlock()
    plugin, ok := m.plugins[name]
    return plugin, ok
}

// GetRequestPlugins 获取所有请求处理插件
func (m *DefaultManager) GetRequestPlugins() []RequestPlugin {
    m.mu.RLock()
    defer m.mu.RUnlock()
    return append([]RequestPlugin(nil), m.requestPlugins...)
}

// GetResponsePlugins 获取所有响应处理插件
func (m *DefaultManager) GetResponsePlugins() []ResponsePlugin {
    m.mu.RLock()
    defer m.mu.RUnlock()
    return append([]ResponsePlugin(nil), m.responsePlugins...)
}

// GetErrorPlugins 获取所有错误处理插件
func (m *DefaultManager) GetErrorPlugins() []ErrorPlugin {
    m.mu.RLock()
    defer m.mu.RUnlock()
    return append([]ErrorPlugin(nil), m.errorPlugins...)
}

// InitAll 初始化所有插件
func (m *DefaultManager) InitAll(config map[string]interface{}) error {
    m.mu.RLock()
    defer m.mu.RUnlock()

    for name, plugin := range m.plugins {
        pluginConfig := config[name]
        cfg, ok := pluginConfig.(map[string]interface{})
        if !ok {
            cfg = make(map[string]interface{})
        }

        if err := plugin.Init(cfg); err != nil {
            return fmt.Errorf("failed to init plugin %s: %w", name, err)
        }
        logger.Infof("Plugin %s initialized", name)
    }

    return nil
}

// Unregister 卸载插件
func (m *DefaultManager) Unregister(name string) error {
    m.mu.Lock()
    defer m.mu.Unlock()

    plugin, ok := m.plugins[name]
    if !ok {
        return fmt.Errorf("plugin %s not found", name)
    }

    // 先关闭插件
    if err := plugin.Close(); err != nil {
        logger.Warnf("Failed to close plugin %s when unregistering: %v", name, err)
    }

    // 从map中删除
    delete(m.plugins, name)

    // 清理请求插件列表
    newRequestPlugins := make([]RequestPlugin, 0, len(m.requestPlugins))
    for _, p := range m.requestPlugins {
        if p.Name() != name {
            newRequestPlugins = append(newRequestPlugins, p)
        }
    }
    m.requestPlugins = newRequestPlugins

    // 清理响应插件列表
    newResponsePlugins := make([]ResponsePlugin, 0, len(m.responsePlugins))
    for _, p := range m.responsePlugins {
        if p.Name() != name {
            newResponsePlugins = append(newResponsePlugins, p)
        }
    }
    m.responsePlugins = newResponsePlugins

    // 清理错误插件列表
    newErrorPlugins := make([]ErrorPlugin, 0, len(m.errorPlugins))
    for _, p := range m.errorPlugins {
        if p.Name() != name {
            newErrorPlugins = append(newErrorPlugins, p)
        }
    }
    m.errorPlugins = newErrorPlugins

    logger.Infof("Plugin %s unregistered successfully", name)
    return nil
}

// Reload 重新加载插件配置
func (m *DefaultManager) Reload(config map[string]interface{}) error {
    m.mu.RLock()
    defer m.mu.RUnlock()

    var errs []error
    for name, plugin := range m.plugins {
        pluginConfig := config[name]
        cfg, ok := pluginConfig.(map[string]interface{})
        if !ok {
            cfg = make(map[string]interface{})
        }

        if err := plugin.Init(cfg); err != nil {
            errs = append(errs, fmt.Errorf("failed to reload plugin %s: %w", name, err))
        } else {
            logger.Infof("Plugin %s reloaded successfully", name)
        }
    }

    if len(errs) > 0 {
        return fmt.Errorf("failed to reload some plugins: %v", errs)
    }
    return nil
}

// CloseAll 关闭所有插件
func (m *DefaultManager) CloseAll() error {
    m.mu.RLock()
    defer m.mu.RUnlock()

    var errs []error
    for name, plugin := range m.plugins {
        if err := plugin.Close(); err != nil {
            errs = append(errs, fmt.Errorf("failed to close plugin %s: %w", name, err))
        } else {
            logger.Infof("Plugin %s closed", name)
        }
    }

    if len(errs) > 0 {
        return fmt.Errorf("failed to close some plugins: %v", errs)
    }
    return nil
}
