package main

import (
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/zenglw/llm_gateway/internal/apiserver"
	"github.com/zenglw/llm_gateway/internal/llm"
	"github.com/zenglw/llm_gateway/internal/llm/claude"
	"github.com/zenglw/llm_gateway/internal/llm/deepseek"
	"github.com/zenglw/llm_gateway/internal/llm/openai"
	"github.com/zenglw/llm_gateway/internal/plugin"
	"github.com/zenglw/llm_gateway/internal/plugin/auth"
	"github.com/zenglw/llm_gateway/internal/plugin/cache"
	"github.com/zenglw/llm_gateway/internal/plugin/logging"
	"github.com/zenglw/llm_gateway/internal/plugin/metrics"
	"github.com/zenglw/llm_gateway/internal/plugin/ratelimit"
	"github.com/zenglw/llm_gateway/internal/service"
	"github.com/zenglw/llm_gateway/internal/storage"
	"github.com/zenglw/llm_gateway/pkg/config"
	"github.com/zenglw/llm_gateway/pkg/logger"
)

var (
	configFile = flag.String("config", "configs/config.yaml", "path to config file")
	help       = flag.Bool("help", false, "show help")
)

func main() {
	flag.Parse()

	if *help {
		flag.Usage()
		return
	}

	// 加载配置
	cfg, err := config.Load(*configFile)
	if err != nil {
		fmt.Printf("Failed to load config: %v\n", err)
		os.Exit(1)
	}

	// 初始化日志
	if err := logger.Init(&cfg.Log); err != nil {
		fmt.Printf("Failed to init logger: %v\n", err)
		os.Exit(1)
	}
	defer logger.Sync()

	logger.Infof("Config loaded successfully")

	// 注册LLM服务工厂
	openai.Register()
	claude.Register()
	deepseek.Register()

	// 初始化存储
	store, err := storage.NewStore(cfg.Storage)
	if err != nil {
		logger.Fatalf("Failed to init storage: %v", err)
	}
	logger.Infof("Storage initialized, type: %s", cfg.Storage.Type)

	// 初始化插件管理器
	pluginManager := plugin.NewManager()

	// 注册插件
	authPlugin := auth.NewPlugin(store)
	pluginManager.Register(authPlugin)

	ratelimitPlugin := ratelimit.NewPlugin()
	pluginManager.Register(ratelimitPlugin)

	loggingPlugin := logging.NewPlugin()
	pluginManager.Register(loggingPlugin)

	metricsPlugin := metrics.NewPlugin()
	pluginManager.Register(metricsPlugin)

	cachePlugin := cache.NewPlugin(cfg.Storage)
	pluginManager.Register(cachePlugin)

	// 初始化插件
	pluginConfig := make(map[string]interface{})
	pluginConfig["auth"] = map[string]interface{}{
		"jwt_secret": cfg.Auth.JWTSecret,
		"jwt_expire": cfg.Auth.JWTExpire,
	}
	pluginConfig["ratelimit"] = map[string]interface{}{
		"rate":  cfg.Plugin.RateLimit.Rate,
		"burst": cfg.Plugin.RateLimit.Burst,
	}
	pluginConfig["logging"] = map[string]interface{}{
		"log_request":  cfg.Plugin.Logging.LogRequest,
		"log_response": cfg.Plugin.Logging.LogResponse,
	}
	pluginConfig["metrics"] = map[string]interface{}{
		"enabled": cfg.Plugin.Metrics.Enabled,
		"path":    cfg.Plugin.Metrics.Path,
	}
	pluginConfig["cache"] = map[string]interface{}{
		"enabled":    cfg.Plugin.Cache.Enabled,
		"ttl":        cfg.Plugin.Cache.TTL,
		"max_size":   cfg.Plugin.Cache.MaxSize,
		"type":       cfg.Plugin.Cache.Type,
		"prefix":     cfg.Plugin.Cache.Prefix,
		"model_skip": cfg.Plugin.Cache.ModelSkip,
	}

	if err := pluginManager.InitAll(pluginConfig); err != nil {
		logger.Fatalf("Failed to init plugins: %v", err)
	}
	logger.Infof("Plugins initialized")

	// 初始化LLM服务
	var llmServices []llm.Service
	for name, svcCfg := range cfg.LLM.Services {
		if !svcCfg.Enabled {
			continue
		}
		svc, err := llm.CreateService(svcCfg.Type, &llm.Config{
			APIKey:     svcCfg.APIKey,
			BaseURL:    svcCfg.BaseURL,
			Timeout:    svcCfg.Timeout,
			MaxRetries: svcCfg.MaxRetries,
		})
		if err != nil {
			logger.Errorf("Failed to create LLM service %s: %v", name, err)
			continue
		}
		llmServices = append(llmServices, svc)
		logger.Infof("LLM service %s initialized", name)
	}

	// 初始化业务服务
	apiKeyService := service.NewAPIKeyService(store)
	quotaService := service.NewQuotaService(store)
	llmRouter := service.NewLLMRouterService(pluginManager, llmServices, quotaService)
	logger.Infof("Business services initialized")

	// 初始化API服务
	handler := apiserver.NewHandler(llmRouter, apiKeyService, quotaService)
	router := apiserver.NewRouter(handler, cfg.Server.Mode)
	logger.Infof("API server initialized")

	// 启动服务
	addr := fmt.Sprintf(":%d", cfg.Server.Port)
	logger.Infof("Starting LLM gateway server on %s", addr)

	// 信号处理
	quit := make(chan os.Signal, 1)
	reload := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	signal.Notify(reload, syscall.SIGHUP) // SIGHUP信号触发重载

	// 启动服务
	go func() {
		if err := router.Run(addr); err != nil {
			logger.Fatalf("Failed to start server: %v", err)
		}
	}()

	// 处理信号
	for {
		select {
		case <-quit:
			logger.Infof("Shutting down server...")
			// 关闭插件
			if err := pluginManager.CloseAll(); err != nil {
				logger.Errorf("Failed to close plugins: %v", err)
			}
			// 关闭存储
			if err := store.Close(); err != nil {
				logger.Errorf("Failed to close storage: %v", err)
			}
			logger.Infof("Server exited successfully")
			return
		case <-reload:
			logger.Infof("Received reload signal, reloading configuration...")
			// 重新加载配置
			newCfg, err := config.Reload()
			if err != nil {
				logger.Errorf("Failed to reload config: %v", err)
				continue
			}
			// 重新加载插件配置
			pluginConfig := make(map[string]interface{})
			pluginConfig["auth"] = map[string]interface{}{
				"jwt_secret": newCfg.Auth.JWTSecret,
				"jwt_expire": newCfg.Auth.JWTExpire,
			}
			pluginConfig["ratelimit"] = map[string]interface{}{
				"rate":  newCfg.Plugin.RateLimit.Rate,
				"burst": newCfg.Plugin.RateLimit.Burst,
			}
			pluginConfig["logging"] = map[string]interface{}{
				"log_request":  newCfg.Plugin.Logging.LogRequest,
				"log_response": newCfg.Plugin.Logging.LogResponse,
			}
			pluginConfig["metrics"] = map[string]interface{}{
				"enabled": newCfg.Plugin.Metrics.Enabled,
				"path":    newCfg.Plugin.Metrics.Path,
			}
			if err := pluginManager.Reload(pluginConfig); err != nil {
				logger.Errorf("Failed to reload plugins: %v", err)
				continue
			}
			logger.Infof("Configuration reloaded successfully")
		}
	}
}
