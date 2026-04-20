package apiserver

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/zenglw/llm_gateway/internal/model"
	"github.com/zenglw/llm_gateway/internal/service"
	"github.com/zenglw/llm_gateway/pkg/errors"
)

// Handler API处理器
type Handler struct {
	llmRouter     *service.LLMRouterService
	apiKeyService *service.APIKeyService
	quotaService  *service.QuotaService
	batchService  *service.BatchService
	taskService   *service.TaskService
}

// NewHandler 创建API处理器
func NewHandler(llmRouter *service.LLMRouterService, apiKeyService *service.APIKeyService, quotaService *service.QuotaService) *Handler {
	batchService := service.NewBatchService(llmRouter)
	taskService := service.NewTaskService(llmRouter, batchService, 10)
	return &Handler{
		llmRouter:     llmRouter,
		apiKeyService: apiKeyService,
		quotaService:  quotaService,
		batchService:  batchService,
		taskService:   taskService,
	}
}

// 处理请求头，添加上下文参数
func addHeadersToContext(c *gin.Context) context.Context {
	ctx := c.Request.Context()
	// 缓存刷新参数
	cacheRefresh := strings.ToLower(c.GetHeader("X-Cache-Refresh")) == "true"
	ctx = context.WithValue(ctx, "X-Cache-Refresh", cacheRefresh)
	return ctx
}

// ChatCompletion 聊天补全接口
func (h *Handler) ChatCompletion(c *gin.Context) {
	var req model.ChatRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.Error(errors.Wrap(errors.ErrCodeInvalidParams, "invalid request body", err))
		return
	}

	// 添加请求头参数到上下文
	ctx := addHeadersToContext(c)

	// 处理流式请求
	if req.Stream {
		stream, err := h.llmRouter.ChatCompletionStream(ctx, &req)
		if err != nil {
			c.Error(err)
			return
		}

		// 设置流式响应头
		c.Header("Content-Type", "text/event-stream")
		c.Header("Cache-Control", "no-cache")
		c.Header("Connection", "keep-alive")
		c.Header("Transfer-Encoding", "chunked")

		// 流式输出
		writer := c.Writer
		for resp := range stream {
			data, err := json.Marshal(resp)
			if err != nil {
				continue
			}
			_, _ = writer.WriteString("data: ")
			_, _ = writer.Write(data)
			_, _ = writer.WriteString("\n\n")
			writer.Flush()
		}

		// 结束标识
		_, _ = writer.WriteString("data: [DONE]\n\n")
		return
	}

	// 普通请求
	resp, err := h.llmRouter.ChatCompletion(ctx, &req)
	if err != nil {
		c.Error(err)
		return
	}

	// 添加缓存命中响应头
	if _, ok := ctx.Value("cached_response").(*model.LLMResponse); ok {
		c.Header("X-Cache-Hit", "true")
	}

	c.JSON(http.StatusOK, resp)
}

// Completion 文本补全接口
func (h *Handler) Completion(c *gin.Context) {
	var req model.CompletionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.Error(errors.Wrap(errors.ErrCodeInvalidParams, "invalid request body", err))
		return
	}

	// 添加请求头参数到上下文
	ctx := addHeadersToContext(c)

	// 处理流式请求
	if req.Stream {
		stream, err := h.llmRouter.CompletionStream(ctx, &req)
		if err != nil {
			c.Error(err)
			return
		}

		// 设置流式响应头
		c.Header("Content-Type", "text/event-stream")
		c.Header("Cache-Control", "no-cache")
		c.Header("Connection", "keep-alive")
		c.Header("Transfer-Encoding", "chunked")

		// 流式输出
		writer := c.Writer
		for resp := range stream {
			data, err := json.Marshal(resp)
			if err != nil {
				continue
			}
			_, _ = writer.WriteString("data: ")
			_, _ = writer.Write(data)
			_, _ = writer.WriteString("\n\n")
			writer.Flush()
		}

		// 结束标识
		_, _ = writer.WriteString("data: [DONE]\n\n")
		return
	}

	// 普通请求
	resp, err := h.llmRouter.Completion(ctx, &req)
	if err != nil {
		c.Error(err)
		return
	}

	// 添加缓存命中响应头
	if _, ok := ctx.Value("cached_response").(*model.LLMResponse); ok {
		c.Header("X-Cache-Hit", "true")
	}

	c.JSON(http.StatusOK, resp)
}

// CreateAPIKey 创建API Key
func (h *Handler) CreateAPIKey(c *gin.Context) {
	var req model.CreateAPIKeyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.Error(errors.Wrap(errors.ErrCodeInvalidParams, "invalid request body", err))
		return
	}

	apiKey, err := h.apiKeyService.CreateAPIKey(c.Request.Context(), &req)
	if err != nil {
		c.Error(err)
		return
	}

	c.JSON(http.StatusCreated, apiKey)
}

// GetAPIKey 获取API Key
func (h *Handler) GetAPIKey(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.Error(errors.New(errors.ErrCodeInvalidParams, "missing api key id"))
		return
	}

	apiKey, err := h.apiKeyService.GetAPIKey(c.Request.Context(), id)
	if err != nil {
		c.Error(err)
		return
	}

	c.JSON(http.StatusOK, apiKey)
}

// ListAPIKeys 获取用户API Key列表
func (h *Handler) ListAPIKeys(c *gin.Context) {
	userID := c.Query("user_id")
	if userID == "" {
		c.Error(errors.New(errors.ErrCodeInvalidParams, "missing user_id"))
		return
	}

	apiKeys, err := h.apiKeyService.ListAPIKeys(c.Request.Context(), userID)
	if err != nil {
		c.Error(err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": apiKeys,
	})
}

// UpdateAPIKey 更新API Key
func (h *Handler) UpdateAPIKey(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.Error(errors.New(errors.ErrCodeInvalidParams, "missing api key id"))
		return
	}

	var req model.UpdateAPIKeyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.Error(errors.Wrap(errors.ErrCodeInvalidParams, "invalid request body", err))
		return
	}

	apiKey, err := h.apiKeyService.UpdateAPIKey(c.Request.Context(), id, &req)
	if err != nil {
		c.Error(err)
		return
	}

	c.JSON(http.StatusOK, apiKey)
}

// DeleteAPIKey 删除API Key
func (h *Handler) DeleteAPIKey(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.Error(errors.New(errors.ErrCodeInvalidParams, "missing api key id"))
		return
	}

	if err := h.apiKeyService.DeleteAPIKey(c.Request.Context(), id); err != nil {
		c.Error(err)
		return
	}

	c.Status(http.StatusNoContent)
}

// GetUserQuota 获取用户配额
func (h *Handler) GetUserQuota(c *gin.Context) {
	userID := c.Param("user_id")
	if userID == "" {
		c.Error(errors.New(errors.ErrCodeInvalidParams, "missing user_id"))
		return
	}

	quota, err := h.quotaService.GetUserQuota(c.Request.Context(), userID)
	if err != nil {
		c.Error(err)
		return
	}

	c.JSON(http.StatusOK, quota)
}

// UpdateUserQuota 更新用户配额
func (h *Handler) UpdateUserQuota(c *gin.Context) {
	userID := c.Param("user_id")
	if userID == "" {
		c.Error(errors.New(errors.ErrCodeInvalidParams, "missing user_id"))
		return
	}

	var req model.UpdateQuotaRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.Error(errors.Wrap(errors.ErrCodeInvalidParams, "invalid request body", err))
		return
	}

	quota, err := h.quotaService.UpdateUserQuota(c.Request.Context(), userID, &req)
	if err != nil {
		c.Error(err)
		return
	}

	c.JSON(http.StatusOK, quota)
}

// ResetUserQuota 重置用户配额
func (h *Handler) ResetUserQuota(c *gin.Context) {
	userID := c.Param("user_id")
	if userID == "" {
		c.Error(errors.New(errors.ErrCodeInvalidParams, "missing user_id"))
		return
	}

	if err := h.quotaService.ResetUserQuota(c.Request.Context(), userID); err != nil {
		c.Error(err)
		return
	}

	c.Status(http.StatusNoContent)
}

// BatchCompletion 批量补全接口
func (h *Handler) BatchCompletion(c *gin.Context) {
	var req model.BatchRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.Error(errors.Wrap(errors.ErrCodeInvalidParams, "invalid request body", err))
		return
	}

	// 添加请求头参数到上下文
	ctx := addHeadersToContext(c)

	// 处理批量请求
	resp, err := h.batchService.ProcessBatch(ctx, &req)
	if err != nil {
		c.Error(err)
		return
	}

	c.JSON(http.StatusOK, resp)
}

// CreateAsyncTask 创建异步任务
func (h *Handler) CreateAsyncTask(c *gin.Context) {
	var req model.CreateAsyncTaskRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.Error(errors.Wrap(errors.ErrCodeInvalidParams, "invalid request body", err))
		return
	}

	task, err := h.taskService.CreateTask(c.Request.Context(), &req)
	if err != nil {
		c.Error(err)
		return
	}

	c.JSON(http.StatusAccepted, model.AsyncTaskResponse{
		TaskID:    task.ID,
		Object:    "async_task",
		CreatedAt: time.Unix(task.CreatedAt, 0),
		Status:    string(task.Status),
		Message:   "Task created successfully, you can query status via /v1/tasks/{task_id}",
	})
}

// GetTask 查询任务状态
func (h *Handler) GetTask(c *gin.Context) {
	taskID := c.Param("task_id")
	if taskID == "" {
		c.Error(errors.New(errors.ErrCodeInvalidParams, "missing task id"))
		return
	}

	task, err := h.taskService.GetTask(c.Request.Context(), taskID)
	if err != nil {
		c.Error(err)
		return
	}

	c.JSON(http.StatusOK, model.GetTaskResponse{
		Task: *task,
	})
}

// ListTasks 查询任务列表
func (h *Handler) ListTasks(c *gin.Context) {
	var req model.ListTasksRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		c.Error(errors.Wrap(errors.ErrCodeInvalidParams, "invalid query params", err))
		return
	}

	resp, err := h.taskService.ListTasks(c.Request.Context(), &req)
	if err != nil {
		c.Error(err)
		return
	}

	c.JSON(http.StatusOK, resp)
}

// CancelTask 取消任务
func (h *Handler) CancelTask(c *gin.Context) {
	taskID := c.Param("task_id")
	if taskID == "" {
		c.Error(errors.New(errors.ErrCodeInvalidParams, "missing task id"))
		return
	}

	if err := h.taskService.CancelTask(c.Request.Context(), taskID); err != nil {
		c.Error(err)
		return
	}

	c.Status(http.StatusNoContent)
}

// HealthCheck 健康检查
func (h *Handler) HealthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status": "ok",
	})
}
