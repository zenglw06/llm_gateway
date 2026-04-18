package plugin

import (
	"context"
	"time"

	"github.com/zenglw/llm_gateway/internal/model"
)

// TestPlugin 测试用插件
type TestPlugin struct {
	name        string
	SleepTime   time.Duration // 执行耗时，模拟慢请求
	ShouldErr   bool          // 是否返回错误
	ShouldPanic bool          // 是否panic
	callCount   int           // 调用次数
}

func NewTestPlugin(name string) *TestPlugin {
	return &TestPlugin{name: name}
}

func (p *TestPlugin) Name() string {
	return p.name
}

func (p *TestPlugin) Init(config map[string]interface{}) error {
	return nil
}

func (p *TestPlugin) Close() error {
	return nil
}

func (p *TestPlugin) HandleRequest(ctx context.Context, req *model.LLMRequest) (*model.LLMRequest, error) {
	p.callCount++

	if p.ShouldPanic {
		panic("test panic")
	}

	if p.SleepTime > 0 {
		time.Sleep(p.SleepTime)
	}

	if p.ShouldErr {
		return req, context.DeadlineExceeded
	}

	req.UserID = "modified_" + req.UserID
	return req, nil
}

func (p *TestPlugin) HandleResponse(ctx context.Context, resp *model.LLMResponse) (*model.LLMResponse, error) {
	p.callCount++

	if p.ShouldPanic {
		panic("test panic")
	}

	if p.SleepTime > 0 {
		time.Sleep(p.SleepTime)
	}

	if p.ShouldErr {
		return resp, context.DeadlineExceeded
	}

	resp.ID = "modified_" + resp.ID
	return resp, nil
}

func (p *TestPlugin) HandleError(ctx context.Context, err error) error {
	p.callCount++

	if p.ShouldPanic {
		panic("test panic")
	}

	if p.SleepTime > 0 {
		time.Sleep(p.SleepTime)
	}

	if p.ShouldErr {
		return context.DeadlineExceeded
	}

	return err
}

func (p *TestPlugin) CallCount() int {
	return p.callCount
}

func (p *TestPlugin) Reset() {
	p.callCount = 0
	p.SleepTime = 0
	p.ShouldErr = false
	p.ShouldPanic = false
}
