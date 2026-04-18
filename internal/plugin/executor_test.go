package plugin

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/zenglw/llm_gateway/internal/model"
)

func TestExecutorTimeout(t *testing.T) {
	// 创建超时时间为100ms的执行器
	executor := NewPluginExecutorWithConfig(100*time.Millisecond, 0.5, 10, 30*time.Second)
	plugin := NewTestPlugin("timeout_test")
	plugin.SleepTime = 200 * time.Millisecond // 插件执行需要200ms，超过超时时间

	req := &model.LLMRequest{UserID: "test"}
	result, err := executor.ExecuteRequestPlugin(context.Background(), plugin, req)

	assert.NoError(t, err)
	assert.Equal(t, "test", result.UserID) // 超时情况下应该返回原始请求，不修改
	assert.Equal(t, 1, plugin.CallCount()) // 插件确实被调用了一次
}

func TestExecutorPanic(t *testing.T) {
	executor := NewPluginExecutor()
	plugin := NewTestPlugin("panic_test")
	plugin.ShouldPanic = true

	req := &model.LLMRequest{UserID: "test"}
	result, err := executor.ExecuteRequestPlugin(context.Background(), plugin, req)

	assert.NoError(t, err)
	assert.Equal(t, "test", result.UserID) // panic情况下应该返回原始请求
	assert.Equal(t, 1, plugin.CallCount()) // 插件确实被调用了一次
}

func TestExecutorCircuitBreaker(t *testing.T) {
	// 创建窗口大小为2，失败率50%的执行器
	executor := NewPluginExecutorWithConfig(1*time.Second, 0.5, 2, 1*time.Second)
	plugin := NewTestPlugin("circuit_test")
	plugin.ShouldErr = true

	req := &model.LLMRequest{UserID: "test"}

	// 第一次调用，失败 -> 失败率50%，触发熔断
	_, err := executor.ExecuteRequestPlugin(context.Background(), plugin, req)
	assert.NoError(t, err)
	assert.Equal(t, 1, plugin.CallCount())

	// 第二次调用，应该被熔断，插件不会被调用
	_, err = executor.ExecuteRequestPlugin(context.Background(), plugin, req)
	assert.NoError(t, err)
	assert.Equal(t, 1, plugin.CallCount()) // 调用次数还是1，说明熔断生效

	// 第三次调用，还是被熔断
	_, err = executor.ExecuteRequestPlugin(context.Background(), plugin, req)
	assert.NoError(t, err)
	assert.Equal(t, 1, plugin.CallCount())

	// 等待1秒，熔断恢复
	time.Sleep(1100 * time.Millisecond)

	// 第四次调用，半开状态，允许调用
	plugin.ShouldErr = false
	result, err := executor.ExecuteRequestPlugin(context.Background(), plugin, req)
	assert.NoError(t, err)
	assert.Equal(t, "modified_test", result.UserID) // 调用成功，修改了请求
	assert.Equal(t, 2, plugin.CallCount())

	// 后续调用应该正常
	_, err = executor.ExecuteRequestPlugin(context.Background(), plugin, req)
	assert.NoError(t, err)
	assert.Equal(t, 3, plugin.CallCount())
}

func TestExecutorSuccess(t *testing.T) {
	executor := NewPluginExecutor()
	plugin := NewTestPlugin("success_test")

	req := &model.LLMRequest{UserID: "test"}
	result, err := executor.ExecuteRequestPlugin(context.Background(), plugin, req)

	assert.NoError(t, err)
	assert.Equal(t, "modified_test", result.UserID) // 正常修改了请求
	assert.Equal(t, 1, plugin.CallCount())
}

func TestExecutorResponsePlugin(t *testing.T) {
	executor := NewPluginExecutor()
	plugin := NewTestPlugin("response_test")

	resp := &model.LLMResponse{ID: "test_resp"}
	result, err := executor.ExecuteResponsePlugin(context.Background(), plugin, resp)

	assert.NoError(t, err)
	assert.Equal(t, "modified_test_resp", result.ID)
	assert.Equal(t, 1, plugin.CallCount())
}

func TestExecutorErrorPlugin(t *testing.T) {
	executor := NewPluginExecutor()
	plugin := NewTestPlugin("error_test")

	testErr := context.Canceled
	result := executor.ExecuteErrorPlugin(context.Background(), plugin, testErr)

	assert.Equal(t, testErr, result)
	assert.Equal(t, 1, plugin.CallCount())
}
