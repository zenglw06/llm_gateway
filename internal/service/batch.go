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

// BatchService 批量处理服务
type BatchService struct {
	llmRouter *LLMRouterService
}

// NewBatchService 创建批量处理服务
func NewBatchService(llmRouter *LLMRouterService) *BatchService {
	return &BatchService{
		llmRouter: llmRouter,
	}
}

// ProcessBatch 处理批量请求
func (s *BatchService) ProcessBatch(ctx context.Context, req *model.BatchRequest) (*model.BatchResponse, error) {
	resp := &model.BatchResponse{
		ID:      utils.RandomString(16),
		Object:  "batch",
		Created: time.Now().Unix(),
		Model:   req.Model,
	}

	// 使用WaitGroup并发处理子请求
	var wg sync.WaitGroup
	respChan := make(chan model.BatchSubResponse, len(req.Requests))

	for _, subReq := range req.Requests {
		wg.Add(1)
		go func(sr model.BatchSubRequest) {
			defer wg.Done()

			// 合并公共参数和子请求参数
			mergedReq := mergeRequestParams(req, sr)

			// 处理子请求
			subResp := processSingleRequest(ctx, s.llmRouter, mergedReq)
			subResp.CustomID = sr.CustomID

			respChan <- subResp
		}(subReq)
	}

	// 等待所有请求完成
	go func() {
		wg.Wait()
		close(respChan)
	}()

	// 收集响应
	for subResp := range respChan {
		resp.Responses = append(resp.Responses, subResp)
	}

	return resp, nil
}

// mergeRequestParams 合并公共参数和子请求参数
func mergeRequestParams(public *model.BatchRequest, sub model.BatchSubRequest) interface{} {
	// 判断是聊天补全还是文本补全
	if len(sub.Messages) > 0 {
		// 聊天补全请求
		req := &model.ChatRequest{
			Model:       public.Model,
			Messages:    sub.Messages,
			MaxTokens:   public.MaxTokens,
			Temperature: public.Temperature,
			TopP:        public.TopP,
			User:        public.User,
		}
		// 子请求参数覆盖公共参数
		if sub.Model != "" {
			req.Model = sub.Model
		}
		if sub.MaxTokens > 0 {
			req.MaxTokens = sub.MaxTokens
		}
		if sub.Temperature != 0 {
			req.Temperature = sub.Temperature
		}
		if sub.TopP != 0 {
			req.TopP = sub.TopP
		}
		if sub.User != "" {
			req.User = sub.User
		}
		return req
	} else if sub.Prompt != "" {
		// 文本补全请求
		req := &model.CompletionRequest{
			Model:       public.Model,
			Prompt:      sub.Prompt,
			MaxTokens:   public.MaxTokens,
			Temperature: public.Temperature,
			TopP:        public.TopP,
			User:        public.User,
		}
		// 子请求参数覆盖公共参数
		if sub.Model != "" {
			req.Model = sub.Model
		}
		if sub.MaxTokens > 0 {
			req.MaxTokens = sub.MaxTokens
		}
		if sub.Temperature != 0 {
			req.Temperature = sub.Temperature
		}
		if sub.TopP != 0 {
			req.TopP = sub.TopP
		}
		if sub.User != "" {
			req.User = sub.User
		}
		return req
	}
	// 两种类型都没有，返回nil
	return nil
}

// processSingleRequest 处理单个请求
func processSingleRequest(ctx context.Context, router *LLMRouterService, req interface{}) model.BatchSubResponse {
	var subResp model.BatchSubResponse

	if req == nil {
		subResp.Status = "failed"
		subResp.Error = &model.ErrorResponse{
			Code:    "invalid_params",
			Message: "request must have either messages or prompt",
		}
		return subResp
	}

	switch r := req.(type) {
	case *model.ChatRequest:
		// 处理聊天补全请求
		resp, err := router.ChatCompletion(ctx, r)
		if err != nil {
			subResp.Status = "failed"
			if e, ok := err.(*errors.Error); ok {
				subResp.Error = &model.ErrorResponse{
					Code:    fmt.Sprintf("%d", e.Code),
					Message: e.Message,
				}
			} else {
				subResp.Error = &model.ErrorResponse{
					Code:    "internal_error",
					Message: err.Error(),
				}
			}
			return subResp
		}
		// 构造响应
		subResp.Status = "success"
		subResp.Response = &model.BatchSubResponseData{
			ID:      resp.ID,
			Object:  resp.Object,
			Created: resp.Created,
			Model:   resp.Model,
			Choices: resp.Choices,
			Usage:   resp.Usage,
		}

	case *model.CompletionRequest:
		// 处理文本补全请求
		resp, err := router.Completion(ctx, r)
		if err != nil {
			subResp.Status = "failed"
			if e, ok := err.(*errors.Error); ok {
				subResp.Error = &model.ErrorResponse{
					Code:    fmt.Sprintf("%d", e.Code),
					Message: e.Message,
				}
			} else {
				subResp.Error = &model.ErrorResponse{
					Code:    "internal_error",
					Message: err.Error(),
				}
			}
			return subResp
		}
		// 构造响应
		subResp.Status = "success"
		subResp.Response = &model.BatchSubResponseData{
			ID:      resp.ID,
			Object:  resp.Object,
			Created: resp.Created,
			Model:   resp.Model,
			Choices: resp.Choices,
			Usage:   resp.Usage,
		}

	default:
		subResp.Status = "failed"
		subResp.Error = &model.ErrorResponse{
			Code:    string(errors.ErrCodeInvalidParams),
			Message: "unsupported request type",
		}
	}

	return subResp
}
