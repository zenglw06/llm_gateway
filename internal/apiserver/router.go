package apiserver

import (
	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// Router API路由
type Router struct {
	handler *Handler
	engine  *gin.Engine
}

// NewRouter 创建API路由
func NewRouter(handler *Handler, mode string) *Router {
	if mode == "" {
		mode = gin.ReleaseMode
	}
	gin.SetMode(mode)

	engine := gin.New()

	// 注册中间件
	engine.Use(PanicRecoveryMiddleware())
	engine.Use(LoggerMiddleware())
	engine.Use(CORSMiddleware())
	engine.Use(ErrorHandlerMiddleware())
	engine.Use(ValidateContentTypeMiddleware())

	router := &Router{
		handler: handler,
		engine:  engine,
	}

	router.registerRoutes()
	return router
}

// registerRoutes 注册路由
func (r *Router) registerRoutes() {
	// 健康检查
	r.engine.GET("/health", r.handler.HealthCheck)

	// Metrics接口
	r.engine.GET("/metrics", gin.WrapH(promhttp.Handler()))

	// OpenAI兼容API v1组
	v1 := r.engine.Group("/v1")
	v1.Use(AuthMiddleware())
	{
		// 聊天补全
		v1.POST("/chat/completions", r.handler.ChatCompletion)

		// 文本补全
		v1.POST("/completions", r.handler.Completion)

		// 批量补全
		v1.POST("/batch/completions", r.handler.BatchCompletion)

		// API Key管理
		apiKeys := v1.Group("/api-keys")
		{
			apiKeys.POST("", r.handler.CreateAPIKey)
			apiKeys.GET("/:id", r.handler.GetAPIKey)
			apiKeys.GET("", r.handler.ListAPIKeys)
			apiKeys.PUT("/:id", r.handler.UpdateAPIKey)
			apiKeys.DELETE("/:id", r.handler.DeleteAPIKey)
		}

		// 配额管理
		quota := v1.Group("/quota")
		{
			quota.GET("/:user_id", r.handler.GetUserQuota)
			quota.PUT("/:user_id", r.handler.UpdateUserQuota)
			quota.POST("/:user_id/reset", r.handler.ResetUserQuota)
		}

		// 异步任务管理
		tasks := v1.Group("/tasks")
		{
			tasks.POST("", r.handler.CreateAsyncTask)
			tasks.GET("", r.handler.ListTasks)
			tasks.GET("/:task_id", r.handler.GetTask)
			tasks.DELETE("/:task_id", r.handler.CancelTask)
		}
	}
}

// GetEngine 获取gin引擎
func (r *Router) GetEngine() *gin.Engine {
	return r.engine
}

// Run 启动服务
func (r *Router) Run(addr string) error {
	return r.engine.Run(addr)
}
