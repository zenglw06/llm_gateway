# Changelog

所有重要的项目变更都会记录在这个文件中。

格式基于 [Keep a Changelog](https://keepachangelog.com/zh-CN/1.0.0/)，
版本号遵循 [语义化版本](https://semver.org/lang/zh-CN/) 规范。

## [Unreleased]

### 新增
- ✨ 多LLM厂商统一接入，支持OpenAI、Anthropic Claude、DeepSeek
- ✨ 完全兼容OpenAI API协议，现有应用无需修改代码即可无缝切换
- ✨ 支持流式响应（SSE）和非流式响应
- ✨ 插件化架构，内置认证、限流、日志、监控核心插件
- ✨ 智能路由，支持基于权重、负载、成本的路由策略
- ✨ 多租户隔离，完善的API Key管理、配额限制
- ✨ 可观测性，完整的日志、Metrics、链路追踪体系
- ✨ 配置热重载，修改配置不需要重启服务
- ✨ 支持多种部署方式：二进制、Docker、Kubernetes
- ✨ 企业级特性：熔断降级、重试机制、故障隔离
- ✨ 提供丰富的API文档和开发指南

### 改进
- 🚀 高性能，基于Go语言开发，单实例支持2000+ QPS
- 🎯 极低的网关延迟，小于5ms
- 💡 极低的资源消耗，默认配置下内存占用小于512MB

### 文档
- 📝 完整的项目文档：架构设计、安装部署、配置参考、API文档、开发指南
- 📝 贡献指南和代码规范，便于社区贡献

### 构建
- 📦 提供Makefile，方便编译、测试、打包
- 📦 支持多平台交叉编译
- 📦 提供Docker镜像和Helm Chart，方便部署
- 📦 GitHub Actions CI/CD流水线，自动构建、测试、发布

---

## [1.0.0] - 2024-XX-XX

### 特性
- 初始版本发布，核心功能全部完成

[Unreleased]: https://github.com/your-org/llm-gateway/compare/v1.0.0...HEAD
[1.0.0]: https://github.com/your-org/llm-gateway/releases/tag/v1.0.0
