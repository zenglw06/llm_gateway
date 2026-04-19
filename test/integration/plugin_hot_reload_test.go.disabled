package integration

import (
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/zenglw/llm_gateway/internal/plugin"
	"github.com/zenglw/llm_gateway/pkg/config"
)

// 测试插件热重载
func TestPluginHotReload(t *testing.T) {
	// 创建临时配置文件
	tmpConfig := `
plugin:
  ratelimit:
    rate: 100
    burst: 200
  auth:
    enabled: true
`
	tmpFile, err := os.CreateTemp("", "config_*.yaml")
	assert.NoError(t, err)
	defer os.Remove(tmpFile.Name())

	_, err = tmpFile.WriteString(tmpConfig)
	assert.NoError(t, err)
	err = tmpFile.Close()
	assert.NoError(t, err)

	// 加载配置
	cfg, err := config.Load(tmpFile.Name())
	assert.NoError(t, err)

	// 初始化插件管理器
	pluginManager := plugin.NewManager()
	authPlugin := plugin.NewTestPlugin("auth")
	rateLimitPlugin := plugin.NewTestPlugin("ratelimit")
	err = pluginManager.Register(authPlugin)
	assert.NoError(t, err)
	err = pluginManager.Register(rateLimitPlugin)
	assert.NoError(t, err)

	// 初始化插件
	pluginConfig := make(map[string]interface{})
	pluginConfig["auth"] = map[string]interface{}{
		"jwt_secret": cfg.Auth.JWTSecret,
		"jwt_expire": cfg.Auth.JWTExpire,
	}
	pluginConfig["ratelimit"] = map[string]interface{}{
		"rate":  cfg.Plugin.RateLimit.Rate,
		"burst": cfg.Plugin.RateLimit.Burst,
	}
	err = pluginManager.InitAll(pluginConfig)
	assert.NoError(t, err)

	// 修改配置文件
	newConfig := `
plugin:
  ratelimit:
    rate: 500
    burst: 1000
  auth:
    enabled: true
`
	err = os.WriteFile(tmpFile.Name(), []byte(newConfig), 0644)
	assert.NoError(t, err)

	// 重载配置
	newCfg, err := config.Reload()
	assert.NoError(t, err)
	assert.Equal(t, 500, newCfg.Plugin.RateLimit.Rate)
	assert.Equal(t, 1000, newCfg.Plugin.RateLimit.Burst)

	// 重载插件
	newPluginConfig := make(map[string]interface{})
	newPluginConfig["ratelimit"] = map[string]interface{}{
		"rate":  newCfg.Plugin.RateLimit.Rate,
		"burst": newCfg.Plugin.RateLimit.Burst,
	}
	err = pluginManager.Reload(newPluginConfig)
	assert.NoError(t, err)
}
