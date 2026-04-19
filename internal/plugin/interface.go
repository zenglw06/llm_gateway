package plugin

import (
	"context"

	"github.com/zenglw/llm_gateway/internal/model"
)

// Plugin 基础插件接口
type Plugin interface {
	// Name 返回插件名称
	Name() string
	// Init 初始化插件
	Init(config map[string]interface{}) error
	// Close 关闭插件
	Close() error
}

// RequestPlugin 请求处理插件接口
type RequestPlugin interface {
	Plugin
	// HandleRequest 处理请求，返回修改后的请求或错误
	HandleRequest(ctx context.Context, req *model.LLMRequest) (*model.LLMRequest, error)
}

// ResponsePlugin 响应处理插件接口
type ResponsePlugin interface {
	Plugin
	// HandleResponse 处理响应，返回修改后的响应或错误
	HandleResponse(ctx context.Context, resp *model.LLMResponse) (*model.LLMResponse, error)
}

// ErrorPlugin 错误处理插件接口
type ErrorPlugin interface {
	Plugin
	// HandleError 处理错误，返回修改后的错误或nil
	HandleError(ctx context.Context, err error) error
}

// Manager 插件管理器接口
type Manager interface {
	// Register 注册插件
	Register(plugin Plugin) error
	// Unregister 卸载插件
	Unregister(name string) error
	// GetPlugin 获取插件
	GetPlugin(name string) (Plugin, bool)
	// GetRequestPlugins 获取所有请求处理插件
	GetRequestPlugins() []RequestPlugin
	// GetResponsePlugins 获取所有响应处理插件
	GetResponsePlugins() []ResponsePlugin
	// GetErrorPlugins 获取所有错误处理插件
	GetErrorPlugins() []ErrorPlugin
	// InitAll 初始化所有插件
	InitAll(config map[string]interface{}) error
	// CloseAll 关闭所有插件
	CloseAll() error
	// Reload 重新加载插件配置
	Reload(config map[string]interface{}) error
}
