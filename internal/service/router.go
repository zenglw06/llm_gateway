package service

import (
	"context"

	"github.com/zenglw/llm_gateway/internal/llm"
	"github.com/zenglw/llm_gateway/internal/model"
	"github.com/zenglw/llm_gateway/internal/plugin"
	"github.com/zenglw/llm_gateway/pkg/errors"
)

// LLMRouterService LLM请求路由服务
type LLMRouterService struct {
	pluginManager *plugin.DefaultManager
	llmServices   []llm.Service
	quotaService  *QuotaService
	executor      *plugin.PluginExecutor
	lb            LoadBalancer
}

// NewLLMRouterService 创建LLM路由服务
func NewLLMRouterService(pluginManager *plugin.DefaultManager, llmServices []llm.Service, quotaService *QuotaService) *LLMRouterService {
	return &LLMRouterService{
		pluginManager: pluginManager,
		llmServices:   llmServices,
		quotaService:  quotaService,
		executor:      plugin.NewPluginExecutor(),
		lb:            NewRoundRobinLoadBalancer(), // 默认使用轮询负载均衡
	}
}

// NewLLMRouterServiceWithLoadBalancer 使用指定负载均衡器创建LLM路由服务
func NewLLMRouterServiceWithLoadBalancer(pluginManager *plugin.DefaultManager, llmServices []llm.Service, quotaService *QuotaService, lb LoadBalancer) *LLMRouterService {
	return &LLMRouterService{
		pluginManager: pluginManager,
		llmServices:   llmServices,
		quotaService:  quotaService,
		executor:      plugin.NewPluginExecutor(),
		lb:            lb,
	}
}

// ChatCompletion 聊天补全
func (s *LLMRouterService) ChatCompletion(ctx context.Context, req *model.ChatRequest) (*model.ChatResponse, error) {
	// 转换为统一的LLM请求
	llmReq := &model.LLMRequest{
		Model:       req.Model,
		Messages:    req.Messages,
		MaxTokens:   req.MaxTokens,
		Temperature: req.Temperature,
		TopP:        req.TopP,
		Stream:      req.Stream,
		User:        req.User,
		Context:     ctx,
	}

	// 执行请求处理插件链
	var err error
	for _, p := range s.pluginManager.GetRequestPlugins() {
		llmReq, err = s.executor.ExecuteRequestPlugin(ctx, p, llmReq)
		if err != nil {
			// 执行错误处理插件链
			for _, ep := range s.pluginManager.GetErrorPlugins() {
				err = s.executor.ExecuteErrorPlugin(ctx, ep, err)
			}
			return nil, err
		}
	}

	// 检查是否有缓存命中
	if cachedResp, ok := ctx.Value("cached_response").(*model.LLMResponse); ok {
		// 转换为ChatResponse直接返回
		chatResp := &model.ChatResponse{
			ID:      cachedResp.ID,
			Object:  cachedResp.Object,
			Created: cachedResp.Created,
			Model:   cachedResp.Model,
			Choices: make([]model.ChatChoice, len(cachedResp.Choices)),
			Usage:   cachedResp.Usage,
		}
		for i, c := range cachedResp.Choices {
			chatResp.Choices[i] = model.ChatChoice{
				Index:        c.Index,
				Message:      c.Message,
				FinishReason: c.FinishReason,
			}
		}
		return chatResp, nil
	}

	// 检查配额
	if _, err := s.quotaService.ConsumeQuota(ctx, llmReq.UserID, 1); err != nil {
		// 执行错误处理插件链
		for _, ep := range s.pluginManager.GetErrorPlugins() {
			err = s.executor.ExecuteErrorPlugin(ctx, ep, err)
		}
		return nil, err
	}

	// 查找支持该模型的服务
	service, err := s.findServiceForModel(req.Model)
	if err != nil {
		// 执行错误处理插件链
		for _, ep := range s.pluginManager.GetErrorPlugins() {
			err = s.executor.ExecuteErrorPlugin(ctx, ep, err)
		}
		return nil, err
	}

	// 调用LLM服务
	resp, err := service.ChatCompletion(ctx, req)
	if err != nil {
		// 执行错误处理插件链
		for _, ep := range s.pluginManager.GetErrorPlugins() {
			err = s.executor.ExecuteErrorPlugin(ctx, ep, err)
		}
		return nil, err
	}

	// 转换为统一的LLM响应
	llmResp := &model.LLMResponse{
		ID:      resp.ID,
		Object:  resp.Object,
		Created: resp.Created,
		Model:   resp.Model,
		Choices: make([]model.Choice, len(resp.Choices)),
		Usage:   resp.Usage,
	}
	for i, c := range resp.Choices {
		llmResp.Choices[i] = model.Choice{
			Index:        c.Index,
			Message:      c.Message,
			FinishReason: c.FinishReason,
		}
	}

	// 执行响应处理插件链
	for _, p := range s.pluginManager.GetResponsePlugins() {
		llmResp, err = s.executor.ExecuteResponsePlugin(ctx, p, llmResp)
		if err != nil {
			// 执行错误处理插件链
			for _, ep := range s.pluginManager.GetErrorPlugins() {
				err = s.executor.ExecuteErrorPlugin(ctx, ep, err)
			}
			return nil, err
		}
	}

	// 转换回ChatResponse
	chatResp := &model.ChatResponse{
		ID:      llmResp.ID,
		Object:  llmResp.Object,
		Created: llmResp.Created,
		Model:   llmResp.Model,
		Choices: make([]model.ChatChoice, len(llmResp.Choices)),
		Usage:   llmResp.Usage,
	}
	for i, c := range llmResp.Choices {
		chatResp.Choices[i] = model.ChatChoice{
			Index:        c.Index,
			Message:      c.Message,
			FinishReason: c.FinishReason,
		}
	}

	return chatResp, nil
}

// ChatCompletionStream 流式聊天补全
func (s *LLMRouterService) ChatCompletionStream(ctx context.Context, req *model.ChatRequest) (<-chan *model.StreamResponse, error) {
	// 转换为统一的LLM请求
	llmReq := &model.LLMRequest{
		Model:       req.Model,
		Messages:    req.Messages,
		MaxTokens:   req.MaxTokens,
		Temperature: req.Temperature,
		TopP:        req.TopP,
		Stream:      req.Stream,
		User:        req.User,
		Context:     ctx,
	}

	// 执行请求处理插件链
	var err error
	for _, p := range s.pluginManager.GetRequestPlugins() {
		llmReq, err = s.executor.ExecuteRequestPlugin(ctx, p, llmReq)
		if err != nil {
			// 执行错误处理插件链
			for _, ep := range s.pluginManager.GetErrorPlugins() {
				err = s.executor.ExecuteErrorPlugin(ctx, ep, err)
			}
			return nil, err
		}
	}

	// 检查配额
	if _, err := s.quotaService.ConsumeQuota(ctx, llmReq.UserID, 1); err != nil {
		// 执行错误处理插件链
		for _, ep := range s.pluginManager.GetErrorPlugins() {
			err = s.executor.ExecuteErrorPlugin(ctx, ep, err)
		}
		return nil, err
	}

	// 查找支持该模型的服务
	service, err := s.findServiceForModel(req.Model)
	if err != nil {
		// 执行错误处理插件链
		for _, ep := range s.pluginManager.GetErrorPlugins() {
			err = s.executor.ExecuteErrorPlugin(ctx, ep, err)
		}
		return nil, err
	}

	// 调用LLM服务
	stream, err := service.ChatCompletionStream(ctx, req)
	if err != nil {
		// 执行错误处理插件链
		for _, ep := range s.pluginManager.GetErrorPlugins() {
			err = s.executor.ExecuteErrorPlugin(ctx, ep, err)
		}
		return nil, err
	}

	// 创建响应通道
	respChan := make(chan *model.StreamResponse, 100)

	// 异步处理流式响应
	go func() {
		defer close(respChan)

		for resp := range stream {
			// 执行响应处理插件链
			llmResp := &model.LLMResponse{
				ID:      resp.ID,
				Object:  resp.Object,
				Created: resp.Created,
				Model:   resp.Model,
				Choices: make([]model.Choice, len(resp.Choices)),
			}
			for i, c := range resp.Choices {
				llmResp.Choices[i] = model.Choice{
					Index:        c.Index,
					Delta:        c.Delta,
					FinishReason: *c.FinishReason,
				}
			}

			var err error
			for _, p := range s.pluginManager.GetResponsePlugins() {
				llmResp, err = p.HandleResponse(ctx, llmResp)
				if err != nil {
					// 执行错误处理插件链
					for _, ep := range s.pluginManager.GetErrorPlugins() {
						err = ep.HandleError(ctx, err)
					}
					return
				}
			}

			// 转换回StreamResponse
			streamResp := &model.StreamResponse{
				ID:      llmResp.ID,
				Object:  llmResp.Object,
				Created: llmResp.Created,
				Model:   llmResp.Model,
				Choices: make([]model.StreamChoice, len(llmResp.Choices)),
			}
			for i, c := range llmResp.Choices {
				finishReason := c.FinishReason
				streamResp.Choices[i] = model.StreamChoice{
					Index:        c.Index,
					Delta:        c.Delta,
					FinishReason: &finishReason,
				}
			}

			respChan <- streamResp
		}
	}()

	return respChan, nil
}

// Completion 文本补全
func (s *LLMRouterService) Completion(ctx context.Context, req *model.CompletionRequest) (*model.CompletionResponse, error) {
	// 转换为统一的LLM请求
	llmReq := &model.LLMRequest{
		Model:       req.Model,
		Prompt:      req.Prompt,
		MaxTokens:   req.MaxTokens,
		Temperature: req.Temperature,
		TopP:        req.TopP,
		Stream:      req.Stream,
		User:        req.User,
		Context:     ctx,
	}

	// 执行请求处理插件链
	var err error
	for _, p := range s.pluginManager.GetRequestPlugins() {
		llmReq, err = p.HandleRequest(ctx, llmReq)
		if err != nil {
			// 执行错误处理插件链
			for _, ep := range s.pluginManager.GetErrorPlugins() {
				err = ep.HandleError(ctx, err)
			}
			return nil, err
		}
	}

	// 检查是否有缓存命中
	if cachedResp, ok := ctx.Value("cached_response").(*model.LLMResponse); ok {
		// 转换为CompletionResponse直接返回
		completionResp := &model.CompletionResponse{
			ID:      cachedResp.ID,
			Object:  cachedResp.Object,
			Created: cachedResp.Created,
			Model:   cachedResp.Model,
			Choices: make([]model.CompletionChoice, len(cachedResp.Choices)),
			Usage:   cachedResp.Usage,
		}
		// 文本补全的结果在之前的处理中已经保存到completionTexts变量？不对，需要看一下原有逻辑
		// 哦，原有逻辑是在调用LLM之后保存completionTexts，但是缓存的响应中没有这个，所以需要处理
		// 这里我们假设缓存的响应中Choices的Text字段已经包含了结果
		for i, c := range cachedResp.Choices {
			completionResp.Choices[i] = model.CompletionChoice{
				Index:        c.Index,
				Text:         c.Text,
				FinishReason: c.FinishReason,
			}
		}
		return completionResp, nil
	}

	// 检查配额
	if _, err := s.quotaService.ConsumeQuota(ctx, llmReq.UserID, 1); err != nil {
		// 执行错误处理插件链
		for _, ep := range s.pluginManager.GetErrorPlugins() {
			err = ep.HandleError(ctx, err)
		}
		return nil, err
	}

	// 查找支持该模型的服务
	service, err := s.findServiceForModel(req.Model)
	if err != nil {
		// 执行错误处理插件链
		for _, ep := range s.pluginManager.GetErrorPlugins() {
			err = ep.HandleError(ctx, err)
		}
		return nil, err
	}

	// 调用LLM服务
	resp, err := service.Completion(ctx, req)
	if err != nil {
		// 执行错误处理插件链
		for _, ep := range s.pluginManager.GetErrorPlugins() {
			err = ep.HandleError(ctx, err)
		}
		return nil, err
	}

	// 转换为统一的LLM响应
	llmResp := &model.LLMResponse{
		ID:      resp.ID,
		Object:  resp.Object,
		Created: resp.Created,
		Model:   resp.Model,
		Choices: make([]model.Choice, len(resp.Choices)),
		Usage:   resp.Usage,
	}
	// 保存文本补全的结果
	completionTexts := make([]string, len(resp.Choices))
	for i, c := range resp.Choices {
		llmResp.Choices[i] = model.Choice{
			Index:        c.Index,
			Text:         c.Text,
			FinishReason: c.FinishReason,
		}
		completionTexts[i] = c.Text
	}

	// 执行响应处理插件链
	for _, p := range s.pluginManager.GetResponsePlugins() {
		llmResp, err = p.HandleResponse(ctx, llmResp)
		if err != nil {
			// 执行错误处理插件链
			for _, ep := range s.pluginManager.GetErrorPlugins() {
				err = ep.HandleError(ctx, err)
			}
			return nil, err
		}
	}

	// 转换回CompletionResponse
	completionResp := &model.CompletionResponse{
		ID:      llmResp.ID,
		Object:  llmResp.Object,
		Created: llmResp.Created,
		Model:   llmResp.Model,
		Choices: make([]model.CompletionChoice, len(llmResp.Choices)),
		Usage:   llmResp.Usage,
	}
	for i, c := range llmResp.Choices {
		completionResp.Choices[i] = model.CompletionChoice{
			Index:        c.Index,
			Text:         completionTexts[i],
			FinishReason: c.FinishReason,
		}
	}

	return completionResp, nil
}

// CompletionStream 流式文本补全
func (s *LLMRouterService) CompletionStream(ctx context.Context, req *model.CompletionRequest) (<-chan *model.StreamResponse, error) {
	// 转换为统一的LLM请求
	llmReq := &model.LLMRequest{
		Model:       req.Model,
		Prompt:      req.Prompt,
		MaxTokens:   req.MaxTokens,
		Temperature: req.Temperature,
		TopP:        req.TopP,
		Stream:      req.Stream,
		User:        req.User,
		Context:     ctx,
	}

	// 执行请求处理插件链
	var err error
	for _, p := range s.pluginManager.GetRequestPlugins() {
		llmReq, err = p.HandleRequest(ctx, llmReq)
		if err != nil {
			// 执行错误处理插件链
			for _, ep := range s.pluginManager.GetErrorPlugins() {
				err = ep.HandleError(ctx, err)
			}
			return nil, err
		}
	}

	// 检查配额
	if _, err := s.quotaService.ConsumeQuota(ctx, llmReq.UserID, 1); err != nil {
		// 执行错误处理插件链
		for _, ep := range s.pluginManager.GetErrorPlugins() {
			err = ep.HandleError(ctx, err)
		}
		return nil, err
	}

	// 查找支持该模型的服务
	service, err := s.findServiceForModel(req.Model)
	if err != nil {
		// 执行错误处理插件链
		for _, ep := range s.pluginManager.GetErrorPlugins() {
			err = ep.HandleError(ctx, err)
		}
		return nil, err
	}

	// 调用LLM服务
	stream, err := service.CompletionStream(ctx, req)
	if err != nil {
		// 执行错误处理插件链
		for _, ep := range s.pluginManager.GetErrorPlugins() {
			err = ep.HandleError(ctx, err)
		}
		return nil, err
	}

	return stream, nil
}

// findServiceForModel 查找支持指定模型的服务
func (s *LLMRouterService) findServiceForModel(model string) (llm.Service, error) {
	var supportedServices []llm.Service
	for _, service := range s.llmServices {
		if service.SupportsModel(model) {
			supportedServices = append(supportedServices, service)
		}
	}
	if len(supportedServices) == 0 {
		return nil, errors.New(errors.ErrCodeModelNotSupported, "model not supported")
	}
	return s.lb.Select(supportedServices), nil
}
