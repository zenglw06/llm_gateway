package plugin

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestManagerRegister(t *testing.T) {
	manager := NewManager()
	plugin := NewTestPlugin("test_plugin")

	// 注册插件
	err := manager.Register(plugin)
	assert.NoError(t, err)

	// 重复注册应该报错
	err = manager.Register(plugin)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "already registered")

	// 检查插件是否存在
	foundPlugin, ok := manager.GetPlugin("test_plugin")
	assert.True(t, ok)
	assert.Equal(t, plugin.Name(), foundPlugin.Name())
}

func TestManagerUnregister(t *testing.T) {
	manager := NewManager()
	plugin := NewTestPlugin("test_plugin")
	err := manager.Register(plugin)
	assert.NoError(t, err)

	// 检查插件存在
	_, ok := manager.GetPlugin("test_plugin")
	assert.True(t, ok)

	// 卸载插件
	err = manager.Unregister("test_plugin")
	assert.NoError(t, err)

	// 检查插件不存在
	_, ok = manager.GetPlugin("test_plugin")
	assert.False(t, ok)

	// 卸载不存在的插件应该报错
	err = manager.Unregister("not_exists")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

func TestManagerPluginClassification(t *testing.T) {
	manager := NewManager()
	plugin := NewTestPlugin("test_plugin") // 实现了RequestPlugin, ResponsePlugin, ErrorPlugin
	err := manager.Register(plugin)
	assert.NoError(t, err)

	// 检查分类是否正确
	requestPlugins := manager.GetRequestPlugins()
	assert.Len(t, requestPlugins, 1)
	assert.Equal(t, "test_plugin", requestPlugins[0].Name())

	responsePlugins := manager.GetResponsePlugins()
	assert.Len(t, responsePlugins, 1)
	assert.Equal(t, "test_plugin", responsePlugins[0].Name())

	errorPlugins := manager.GetErrorPlugins()
	assert.Len(t, errorPlugins, 1)
	assert.Equal(t, "test_plugin", errorPlugins[0].Name())

	// 卸载后分类应该为空
	err = manager.Unregister("test_plugin")
	assert.NoError(t, err)

	assert.Len(t, manager.GetRequestPlugins(), 0)
	assert.Len(t, manager.GetResponsePlugins(), 0)
	assert.Len(t, manager.GetErrorPlugins(), 0)
}

func TestManagerReload(t *testing.T) {
	manager := NewManager()
	plugin := NewTestPlugin("test_plugin")
	err := manager.Register(plugin)
	assert.NoError(t, err)

	// 重载配置
	config := make(map[string]interface{})
	config["test_plugin"] = map[string]interface{}{"key": "value"}
	err = manager.Reload(config)
	assert.NoError(t, err)
}

func TestManagerInitAndClose(t *testing.T) {
	manager := NewManager()
	plugin := NewTestPlugin("test_plugin")
	err := manager.Register(plugin)
	assert.NoError(t, err)

	// 初始化所有插件
	config := make(map[string]interface{})
	config["test_plugin"] = map[string]interface{}{}
	err = manager.InitAll(config)
	assert.NoError(t, err)

	// 关闭所有插件
	err = manager.CloseAll()
	assert.NoError(t, err)
}
