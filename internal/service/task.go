package service

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/zenglw/llm_gateway/internal/model"
	"github.com/zenglw/llm_gateway/pkg/errors"
	"github.com/zenglw/llm_gateway/pkg/utils"
)

// TaskService 异步任务服务
type TaskService struct {
	llmRouter   *LLMRouterService
	batchService *BatchService
	callbackService *CallbackService
	tasks       sync.Map // 任务存储，key: taskID, value: *model.Task
	workerCount int      // 工作协程数
	taskChan    chan *model.Task // 任务队列
	ctx         context.Context
	cancel      context.CancelFunc
}

// NewTaskService 创建异步任务服务
func NewTaskService(llmRouter *LLMRouterService, batchService *BatchService, workerCount int) *TaskService {
	if workerCount <= 0 {
		workerCount = 10 // 默认10个工作协程
	}
	ctx, cancel := context.WithCancel(context.Background())
	ts := &TaskService{
		llmRouter:   llmRouter,
		batchService: batchService,
		callbackService: NewCallbackService(),
		workerCount: workerCount,
		taskChan:    make(chan *model.Task, 1000), // 队列长度1000
		ctx:         ctx,
		cancel:      cancel,
	}
	// 启动工作协程
	for i := 0; i < workerCount; i++ {
		go ts.worker()
	}
	return ts
}

// Start 启动任务服务
func (s *TaskService) Start() {
	// 已经在New中启动了工作协程
}

// Stop 停止任务服务
func (s *TaskService) Stop() {
	s.cancel()
	close(s.taskChan)
}

// CreateTask 创建异步任务
func (s *TaskService) CreateTask(ctx context.Context, req *model.CreateAsyncTaskRequest) (*model.Task, error) {
	taskID := fmt.Sprintf("task_%s", utils.RandomString(16))
	now := time.Now().Unix()
	task := &model.Task{
		ID:         taskID,
		Object:     "async_task",
		CreatedAt:  now,
		UpdatedAt:  now,
		Status:     model.TaskStatusPending,
		Type:       req.Type,
		Request:    req.Request,
		WebhookURL: req.WebhookURL,
		Metadata:   req.Metadata,
	}
	// 存储任务
	s.tasks.Store(taskID, task)
	// 加入队列
	s.taskChan <- task
	return task, nil
}

// GetTask 获取任务信息
func (s *TaskService) GetTask(ctx context.Context, taskID string) (*model.Task, error) {
	task, ok := s.tasks.Load(taskID)
	if !ok {
		return nil, errors.New(errors.ErrCodeNotFound, "task not found")
	}
	return task.(*model.Task), nil
}

// CancelTask 取消任务
func (s *TaskService) CancelTask(ctx context.Context, taskID string) error {
	task, ok := s.tasks.Load(taskID)
	if !ok {
		return errors.New(errors.ErrCodeNotFound, "task not found")
	}
	t := task.(*model.Task)
	if t.Status == model.TaskStatusPending || t.Status == model.TaskStatusRunning {
		t.Status = model.TaskStatusCanceled
		t.UpdatedAt = time.Now().Unix()
		s.tasks.Store(taskID, t)
	}
	return nil
}

// ListTasks 获取任务列表
func (s *TaskService) ListTasks(ctx context.Context, req *model.ListTasksRequest) (*model.ListTasksResponse, error) {
	var tasks []model.Task
	var total int

	s.tasks.Range(func(key, value interface{}) bool {
		task := value.(*model.Task)
		// 过滤条件
		if req.UserID != "" && task.UserID != req.UserID {
			return true
		}
		if req.Status != "" && task.Status != req.Status {
			return true
		}
		total++
		tasks = append(tasks, *task)
		return true
	})

	// 分页
	start := req.Offset
	end := req.Offset + req.Limit
	if start >= len(tasks) {
		tasks = []model.Task{}
	} else if end > len(tasks) {
		tasks = tasks[start:]
	} else {
		tasks = tasks[start:end]
	}

	return &model.ListTasksResponse{
		Data:  tasks,
		Total: total,
		Limit: req.Limit,
		Offset: req.Offset,
	}, nil
}

// worker 工作协程，处理任务
func (s *TaskService) worker() {
	for {
		select {
		case <-s.ctx.Done():
			return
		case task, ok := <-s.taskChan:
			if !ok {
				return
			}
			s.processTask(task)
		}
	}
}

// processTask 处理单个任务
func (s *TaskService) processTask(task *model.Task) {
	// 更新状态为运行中
	task.Status = model.TaskStatusRunning
	task.UpdatedAt = time.Now().Unix()
	s.tasks.Store(task.ID, task)

	var err error
	var response interface{}

	// 根据任务类型处理
	switch task.Type {
	case "chat":
		// 聊天补全
		req, ok := task.Request.(map[string]interface{})
		if !ok {
			err = errors.New(errors.ErrCodeInvalidParams, "invalid chat request")
			break
		}
		// 转换为ChatRequest
		chatReq := &model.ChatRequest{}
		if err := utils.MapToStruct(req, chatReq); err != nil {
			err = errors.Wrap(errors.ErrCodeInvalidParams, "invalid chat request format", err)
			break
		}
		// 处理请求
		response, err = s.llmRouter.ChatCompletion(context.Background(), chatReq)

	case "completion":
		// 文本补全
		req, ok := task.Request.(map[string]interface{})
		if !ok {
			err = errors.New(errors.ErrCodeInvalidParams, "invalid completion request")
			break
		}
		// 转换为CompletionRequest
		completionReq := &model.CompletionRequest{}
		if err := utils.MapToStruct(req, completionReq); err != nil {
			err = errors.Wrap(errors.ErrCodeInvalidParams, "invalid completion request format", err)
			break
		}
		// 处理请求
		response, err = s.llmRouter.Completion(context.Background(), completionReq)

	case "batch":
		// 批量处理
		req, ok := task.Request.(map[string]interface{})
		if !ok {
			err = errors.New(errors.ErrCodeInvalidParams, "invalid batch request")
			break
		}
		// 转换为BatchRequest
		batchReq := &model.BatchRequest{}
		if err := utils.MapToStruct(req, batchReq); err != nil {
			err = errors.Wrap(errors.ErrCodeInvalidParams, "invalid batch request format", err)
			break
		}
		// 处理请求
		response, err = s.batchService.ProcessBatch(context.Background(), batchReq)

	default:
		err = errors.New(errors.ErrCodeInvalidParams, "unsupported task type")
	}

	// 更新任务状态
	if err != nil {
		task.Status = model.TaskStatusFailed
		task.Error = err.Error()
	} else {
		task.Status = model.TaskStatusSuccess
		task.Response = response
	}
	task.UpdatedAt = time.Now().Unix()
	s.tasks.Store(task.ID, task)

	// 回调通知
	if task.WebhookURL != "" {
		go s.callbackService.SendCallback(task)
	}
}
