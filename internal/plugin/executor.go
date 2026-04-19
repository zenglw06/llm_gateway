package plugin

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/zenglw/llm_gateway/internal/model"
	"github.com/zenglw/llm_gateway/pkg/logger"
)

// 熔断状态
const (
	circuitClosed   = iota // 关闭状态，正常调用
	circuitOpen            // 打开状态，禁止调用
	circuitHalfOpen        // 半开状态，尝试调用
)

// 默认配置
const (
	defaultTimeout      = 5 * time.Second  // 默认超时时间
	defaultFailureRate  = 0.5              // 默认失败率阈值
	defaultWindowSize   = 10               // 默认统计窗口大小
	defaultRecoveryTime = 30 * time.Second // 默认恢复时间
)

// PluginExecutor 插件执行器，支持超时和熔断
type PluginExecutor struct {
	timeout      time.Duration
	failureRate  float64
	windowSize   int
	recoveryTime time.Duration

	mu           sync.RWMutex
	circuitState map[string]int       // 每个插件的熔断状态
	failureCount map[string]int       // 每个插件的失败计数
	lastFailure  map[string]time.Time // 每个插件的最后失败时间
}

// NewPluginExecutor 创建默认插件执行器
func NewPluginExecutor() *PluginExecutor {
	return &PluginExecutor{
		timeout:      defaultTimeout,
		failureRate:  defaultFailureRate,
		windowSize:   defaultWindowSize,
		recoveryTime: defaultRecoveryTime,
		circuitState: make(map[string]int),
		failureCount: make(map[string]int),
		lastFailure:  make(map[string]time.Time),
	}
}

// NewPluginExecutorWithConfig 创建带配置的插件执行器
func NewPluginExecutorWithConfig(timeout time.Duration, failureRate float64, windowSize int, recoveryTime time.Duration) *PluginExecutor {
	return &PluginExecutor{
		timeout:      timeout,
		failureRate:  failureRate,
		windowSize:   windowSize,
		recoveryTime: recoveryTime,
		circuitState: make(map[string]int),
		failureCount: make(map[string]int),
		lastFailure:  make(map[string]time.Time),
	}
}

// 检查是否允许调用插件
func (e *PluginExecutor) allowCall(pluginName string) bool {
	e.mu.RLock()
	defer e.mu.RUnlock()

	state, ok := e.circuitState[pluginName]
	if !ok {
		return true // 没有状态记录，允许调用
	}

	switch state {
	case circuitClosed:
		return true
	case circuitOpen:
		// 检查是否到了恢复时间
		lastFail, ok := e.lastFailure[pluginName]
		if !ok {
			return false
		}
		if time.Since(lastFail) > e.recoveryTime {
			// 进入半开状态
			e.mu.RUnlock()
			e.mu.Lock()
			e.circuitState[pluginName] = circuitHalfOpen
			e.mu.Unlock()
			e.mu.RLock()
			return true
		}
		return false
	case circuitHalfOpen:
		return true // 半开状态允许尝试调用
	default:
		return true
	}
}

// 记录调用结果
func (e *PluginExecutor) recordResult(pluginName string, success bool) {
	e.mu.Lock()
	defer e.mu.Unlock()

	state, ok := e.circuitState[pluginName]
	if !ok {
		state = circuitClosed
	}

	if success {
		// 调用成功
		if state == circuitHalfOpen {
			// 半开状态下成功，恢复到关闭状态
			e.circuitState[pluginName] = circuitClosed
			e.failureCount[pluginName] = 0
			logger.Infof("Plugin %s circuit recovered to closed state", pluginName)
		} else if state == circuitClosed {
			// 关闭状态下，失败计数减1，最小为0
			if count, ok := e.failureCount[pluginName]; ok && count > 0 {
				e.failureCount[pluginName] = count - 1
			}
		}
	} else {
		// 调用失败
		e.lastFailure[pluginName] = time.Now()
		count := e.failureCount[pluginName] + 1
		e.failureCount[pluginName] = count

		if state == circuitHalfOpen {
			// 半开状态下失败，回到打开状态
			e.circuitState[pluginName] = circuitOpen
			logger.Warnf("Plugin %s circuit open again after half-open test failure", pluginName)
		} else if state == circuitClosed {
			// 检查失败率是否超过阈值
			failureRate := float64(count) / float64(e.windowSize)
			if failureRate >= e.failureRate {
				e.circuitState[pluginName] = circuitOpen
				logger.Warnf("Plugin %s circuit open, failure rate %.2f exceeds threshold %.2f",
					pluginName, failureRate, e.failureRate)
			}
		}
	}
}

// ExecuteRequestPlugin 执行请求处理插件，带超时和熔断
func (e *PluginExecutor) ExecuteRequestPlugin(ctx context.Context, p RequestPlugin, req *model.LLMRequest) (*model.LLMRequest, error) {
	pluginName := p.Name()

	// 检查熔断状态
	if !e.allowCall(pluginName) {
		logger.Warnf("Plugin %s is circuit open, skip execution", pluginName)
		return req, nil // 熔断状态下跳过插件，不返回错误，不影响主流程
	}

	// 创建超时上下文
	timeoutCtx, cancel := context.WithTimeout(ctx, e.timeout)
	defer cancel()

	// 异步执行插件
	resultChan := make(chan struct {
		req *model.LLMRequest
		err error
	}, 1)

	go func() {
		defer func() {
			if r := recover(); r != nil {
				logger.Errorf("Plugin %s panicked: %v", pluginName, r)
				resultChan <- struct {
					req *model.LLMRequest
					err error
				}{req, fmt.Errorf("plugin panicked: %v", r)}
			}
		}()
		req, err := p.HandleRequest(timeoutCtx, req)
		resultChan <- struct {
			req *model.LLMRequest
			err error
		}{req, err}
	}()

	// 等待结果或超时
	select {
	case result := <-resultChan:
		if result.err != nil {
			logger.Errorf("Plugin %s execution failed: %v", pluginName, result.err)
			e.recordResult(pluginName, false)
			return req, nil // 插件执行失败，返回原始请求，不影响主流程
		}
		e.recordResult(pluginName, true)
		return result.req, nil
	case <-timeoutCtx.Done():
		logger.Errorf("Plugin %s execution timed out after %v", pluginName, e.timeout)
		e.recordResult(pluginName, false)
		return req, nil // 超时，返回原始请求，不影响主流程
	}
}

// ExecuteResponsePlugin 执行响应处理插件，带超时和熔断
func (e *PluginExecutor) ExecuteResponsePlugin(ctx context.Context, p ResponsePlugin, resp *model.LLMResponse) (*model.LLMResponse, error) {
	pluginName := p.Name()

	// 检查熔断状态
	if !e.allowCall(pluginName) {
		logger.Warnf("Plugin %s is circuit open, skip execution", pluginName)
		return resp, nil
	}

	// 创建超时上下文
	timeoutCtx, cancel := context.WithTimeout(ctx, e.timeout)
	defer cancel()

	// 异步执行插件
	resultChan := make(chan struct {
		resp *model.LLMResponse
		err  error
	}, 1)

	go func() {
		defer func() {
			if r := recover(); r != nil {
				logger.Errorf("Plugin %s panicked: %v", pluginName, r)
				resultChan <- struct {
					resp *model.LLMResponse
					err  error
				}{resp, fmt.Errorf("plugin panicked: %v", r)}
			}
		}()
		resp, err := p.HandleResponse(timeoutCtx, resp)
		resultChan <- struct {
			resp *model.LLMResponse
			err  error
		}{resp, err}
	}()

	// 等待结果或超时
	select {
	case result := <-resultChan:
		if result.err != nil {
			logger.Errorf("Plugin %s execution failed: %v", pluginName, result.err)
			e.recordResult(pluginName, false)
			return resp, nil
		}
		e.recordResult(pluginName, true)
		return result.resp, nil
	case <-timeoutCtx.Done():
		logger.Errorf("Plugin %s execution timed out after %v", pluginName, e.timeout)
		e.recordResult(pluginName, false)
		return resp, nil
	}
}

// ExecuteErrorPlugin 执行错误处理插件，带超时和熔断
func (e *PluginExecutor) ExecuteErrorPlugin(ctx context.Context, p ErrorPlugin, err error) error {
	pluginName := p.Name()

	// 检查熔断状态
	if !e.allowCall(pluginName) {
		logger.Warnf("Plugin %s is circuit open, skip execution", pluginName)
		return err
	}

	// 创建超时上下文
	timeoutCtx, cancel := context.WithTimeout(ctx, e.timeout)
	defer cancel()

	// 异步执行插件
	resultChan := make(chan error, 1)

	go func() {
		defer func() {
			if r := recover(); r != nil {
				logger.Errorf("Plugin %s panicked: %v", pluginName, r)
				resultChan <- fmt.Errorf("plugin panicked: %v", r)
			}
		}()
		resultChan <- p.HandleError(timeoutCtx, err)
	}()

	// 等待结果或超时
	select {
	case resultErr := <-resultChan:
		if resultErr != nil && resultErr != err {
			logger.Errorf("Plugin %s execution failed: %v", pluginName, resultErr)
			e.recordResult(pluginName, false)
			return err
		}
		e.recordResult(pluginName, true)
		return resultErr
	case <-timeoutCtx.Done():
		logger.Errorf("Plugin %s execution timed out after %v", pluginName, e.timeout)
		e.recordResult(pluginName, false)
		return err
	}
}
