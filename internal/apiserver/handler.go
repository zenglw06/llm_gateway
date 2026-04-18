package apiserver

import (
	"encoding/json"
	"net/http"

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
}

// NewHandler 创建API处理器
func NewHandler(llmRouter *service.LLMRouterService, apiKeyService *service.APIKeyService, quotaService *service.QuotaService) *Handler {
	return &Handler{
		llmRouter:     llmRouter,
		apiKeyService: apiKeyService,
		quotaService:  quotaService,
	}
}

// ChatCompletion 聊天补全接口
func (h *Handler) ChatCompletion(c *gin.Context) {
	var req model.ChatRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.Error(errors.Wrap(errors.ErrCodeInvalidParams, "invalid request body", err))
		return
	}

	// 处理流式请求
	if req.Stream {
		stream, err := h.llmRouter.ChatCompletionStream(c.Request.Context(), &req)
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
	resp, err := h.llmRouter.ChatCompletion(c.Request.Context(), &req)
	if err != nil {
		c.Error(err)
		return
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

	// 处理流式请求
	if req.Stream {
		stream, err := h.llmRouter.CompletionStream(c.Request.Context(), &req)
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
	resp, err := h.llmRouter.Completion(c.Request.Context(), &req)
	if err != nil {
		c.Error(err)
		return
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

// HealthCheck 健康检查
func (h *Handler) HealthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status": "ok",
	})
}
