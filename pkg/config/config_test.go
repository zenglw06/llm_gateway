package config

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

const testConfigContent = `
server:
  port: 8080
  mode: test

auth:
  jwt_secret: test_secret
  jwt_expire: 24

plugin:
  ratelimit:
    rate: 100
    burst: 200
`

func TestLoadConfig(t *testing.T) {
	// 创建临时配置文件
	tmpFile, err := os.CreateTemp("", "config_*.yaml")
	assert.NoError(t, err)
	defer os.Remove(tmpFile.Name())

	// 写入测试配置
	_, err = tmpFile.WriteString(testConfigContent)
	assert.NoError(t, err)
	err = tmpFile.Close()
	assert.NoError(t, err)

	// 加载配置
	cfg, err := Load(tmpFile.Name())
	assert.NoError(t, err)

	// 验证配置
	assert.Equal(t, 8080, cfg.Server.Port)
	assert.Equal(t, "test", cfg.Server.Mode)
	assert.Equal(t, "test_secret", cfg.Auth.JWTSecret)
	assert.Equal(t, 24, cfg.Auth.JWTExpire)
	assert.Equal(t, 100, cfg.Plugin.RateLimit.Rate)
	assert.Equal(t, 200, cfg.Plugin.RateLimit.Burst)
	assert.Equal(t, cfg, Get())
}

func TestReloadConfig(t *testing.T) {
	// 创建临时配置文件
	tmpFile, err := os.CreateTemp("", "config_*.yaml")
	assert.NoError(t, err)
	defer os.Remove(tmpFile.Name())

	// 写入初始配置
	_, err = tmpFile.WriteString(testConfigContent)
	assert.NoError(t, err)
	err = tmpFile.Close()
	assert.NoError(t, err)

	// 第一次加载
	cfg, err := Load(tmpFile.Name())
	assert.NoError(t, err)
	assert.Equal(t, 100, cfg.Plugin.RateLimit.Rate)

	// 修改配置文件
	newContent := `
server:
  port: 8081
  mode: debug

auth:
  jwt_secret: new_secret
  jwt_expire: 48

plugin:
  ratelimit:
    rate: 200
    burst: 400
`
	err = os.WriteFile(tmpFile.Name(), []byte(newContent), 0644)
	assert.NoError(t, err)

	// 重载配置
	newCfg, err := Reload()
	assert.NoError(t, err)

	// 验证新配置
	assert.Equal(t, 8081, newCfg.Server.Port)
	assert.Equal(t, "debug", newCfg.Server.Mode)
	assert.Equal(t, "new_secret", newCfg.Auth.JWTSecret)
	assert.Equal(t, 48, newCfg.Auth.JWTExpire)
	assert.Equal(t, 200, newCfg.Plugin.RateLimit.Rate)
	assert.Equal(t, 400, newCfg.Plugin.RateLimit.Burst)
	assert.Equal(t, newCfg, Get())
}
