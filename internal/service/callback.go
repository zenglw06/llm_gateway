package service

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/zenglw/llm_gateway/internal/model"
	"github.com/zenglw/llm_gateway/pkg/logger"
)

// CallbackService 回调通知服务
type CallbackService struct {
	client *http.Client
}

// NewCallbackService 创建回调服务
func NewCallbackService() *CallbackService {
	return &CallbackService{
		client: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

// CallbackRequest 回调请求结构
type CallbackRequest struct {
	Event     string      `json:"event"` // 事件类型：task.success/task.failed/task.canceled
	TaskID    string      `json:"task_id"`
	Status    string      `json:"status"`
	Task      *model.Task `json:"task"`
	Timestamp int64       `json:"timestamp"`
}

// SendCallback 发送回调通知
func (s *CallbackService) SendCallback(task *model.Task) error {
	// 构造回调请求
	callbackReq := &CallbackRequest{
		TaskID:    task.ID,
		Status:    string(task.Status),
		Task:      task,
		Timestamp: time.Now().Unix(),
	}

	// 根据状态设置事件类型
	switch task.Status {
	case model.TaskStatusSuccess:
		callbackReq.Event = "task.success"
	case model.TaskStatusFailed:
		callbackReq.Event = "task.failed"
	case model.TaskStatusCanceled:
		callbackReq.Event = "task.canceled"
	default:
		callbackReq.Event = "task.updated"
	}

	// 序列化请求体
	body, err := json.Marshal(callbackReq)
	if err != nil {
		logger.Errorf("Failed to marshal callback request: %v", err)
		return err
	}

	// 发送请求，最多重试3次
	var resp *http.Response
	for i := 0; i < 3; i++ {
		req, err := http.NewRequestWithContext(context.Background(), "POST", task.WebhookURL, bytes.NewBuffer(body))
		if err != nil {
			logger.Errorf("Failed to create callback request: %v", err)
			time.Sleep(time.Duration(i+1) * time.Second)
			continue
		}
		req.Header.Set("Content-Type", "application/json")

		resp, err = s.client.Do(req)
		if err == nil && resp.StatusCode >= 200 && resp.StatusCode < 300 {
			// 成功
			logger.Infof("Callback sent successfully to %s for task %s", task.WebhookURL, task.ID)
			resp.Body.Close()
			return nil
		}
		if resp != nil {
			resp.Body.Close()
		}
		logger.Warnf("Callback attempt %d failed for task %s: %v", i+1, task.ID, err)
		time.Sleep(time.Duration(i+1) * time.Second)
	}

	err = fmt.Errorf("all callback attempts failed for task %s", task.ID)
	logger.Error(err.Error())
	return err
}
