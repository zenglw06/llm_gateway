package model

import "time"

// TaskStatus 任务状态
type TaskStatus string

const (
	TaskStatusPending  TaskStatus = "pending"  // 等待处理
	TaskStatusRunning  TaskStatus = "running"  // 处理中
	TaskStatusSuccess  TaskStatus = "success"  // 处理成功
	TaskStatusFailed   TaskStatus = "failed"   // 处理失败
	TaskStatusCanceled TaskStatus = "canceled" // 已取消
)

// Task 异步任务
type Task struct {
	ID         string                 `json:"id"`
	Object     string                 `json:"object"`
	CreatedAt  int64                  `json:"created_at"`
	UpdatedAt  int64                  `json:"updated_at"`
	Status     TaskStatus             `json:"status"`
	Type       string                 `json:"type"` // 任务类型：chat/completion/batch
	Request    interface{}            `json:"request,omitempty"`
	Response   interface{}            `json:"response,omitempty"`
	Error      string                 `json:"error,omitempty"`
	WebhookURL string                 `json:"webhook_url,omitempty"` // 回调URL
	Metadata   map[string]interface{} `json:"metadata,omitempty"`
	UserID     string                 `json:"-"` // 所属用户ID
}

// CreateAsyncTaskRequest 创建异步任务请求
type CreateAsyncTaskRequest struct {
	Type       string                 `json:"type" binding:"required,oneof=chat completion batch"`
	Request    interface{}            `json:"request" binding:"required"`
	WebhookURL string                 `json:"webhook_url,omitempty"`
	Metadata   map[string]interface{} `json:"metadata,omitempty"`
}

// AsyncTaskResponse 异步任务响应
type AsyncTaskResponse struct {
	TaskID    string    `json:"task_id"`
	Object    string    `json:"object"`
	CreatedAt time.Time `json:"created_at"`
	Status    string    `json:"status"`
	Message   string    `json:"message,omitempty"`
}

// GetTaskResponse 获取任务响应
type GetTaskResponse struct {
	Task
}

// ListTasksRequest 任务列表请求
type ListTasksRequest struct {
	UserID string     `form:"user_id"`
	Status TaskStatus `form:"status"`
	Limit  int        `form:"limit,default=20"`
	Offset int        `form:"offset,default=0"`
}

// ListTasksResponse 任务列表响应
type ListTasksResponse struct {
	Data   []Task `json:"data"`
	Total  int    `json:"total"`
	Limit  int    `json:"limit"`
	Offset int    `json:"offset"`
}
