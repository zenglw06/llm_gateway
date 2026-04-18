# 贡献指南

首先，感谢您有兴趣为LLM Gateway项目做贡献！我们欢迎任何形式的贡献，包括但不限于：

- 报告Bug
- 提出功能建议
- 提交代码改进
- 完善文档
- 分享使用案例和最佳实践

## 行为准则

在参与项目之前，请确保您已阅读并遵守我们的[行为准则](CODE_OF_CONDUCT.md)。我们致力于为所有贡献者提供一个友好、包容的环境。

## 如何贡献

### 1. 报告Bug
如果您发现了Bug，请通过GitHub Issue报告，包含以下信息：
- 系统环境（操作系统、Go版本、LLM Gateway版本）
- 详细的复现步骤
- 预期行为和实际行为
- 相关的日志和配置（请隐去敏感信息）

### 2. 提出功能建议
我们欢迎任何新功能的建议，在提出建议前请先确认：
- 该功能是否已经在最新版本中实现
- 是否有类似的功能请求已经存在
- 该功能是否符合项目的定位和发展方向

### 3. 提交代码贡献

#### 准备开发环境
1. Fork本仓库到您的GitHub账号
2. Clone您Fork的仓库到本地：
```bash
git clone https://github.com/your-username/llm-gateway.git
cd llm-gateway
```
3. 安装依赖：
```bash
go mod download
```
4. 安装开发工具：
```bash
# 安装golangci-lint (代码检查工具)
go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest

# 安装pre-commit (可选，提交前自动检查)
pip install pre-commit
pre-commit install
```

#### 开发流程
1. 从main分支创建新的功能分支：
```bash
git checkout -b feature/your-feature-name main
```

2. 进行代码开发，请遵循：
- [代码规范](docs/CODING_STANDARDS.md)
- 所有公共方法、结构体必须有注释
- 新增功能必须包含对应的单元测试
- 确保所有测试通过：`make test`
- 确保代码检查通过：`make lint`

3. 提交代码，提交信息格式遵循[约定式提交](https://www.conventionalcommits.org/zh-hans/v1.0.0/)：
```
<类型>[可选 作用域]: <描述>

[可选 正文]

[可选 页脚]
```
类型包括：
- feat: 新功能
- fix: 修复Bug
- docs: 文档更新
- style: 代码格式调整（不影响代码运行）
- refactor: 重构（既不是新增功能，也不是修改Bug的代码变动）
- perf: 性能优化
- test: 测试相关
- chore: 构建/工具/依赖等相关的变动

示例：
```
feat(auth): 支持API Key IP白名单验证

- 新增IP白名单配置项
- 实现IP校验逻辑
- 增加相关单元测试

Closes #123
```

4. 推送分支到您的Fork仓库：
```bash
git push origin feature/your-feature-name
```

5. 在GitHub上创建Pull Request：
- 目标分支选择本仓库的main分支
- PR标题遵循提交信息格式
- PR描述包含：
  - 变更的内容和目的
  - 相关的Issue（如果有的话）
  - 测试情况
  - 任何需要注意的变更

#### PR审查流程
1. 项目维护者会在3个工作日内对您的PR进行初步审查
2. 可能会要求您做一些修改，请根据反馈进行调整
3. 通过所有检查和审查后，PR会被合并
4. 您会出现在贡献者列表中

### 4. 贡献文档
文档贡献和代码贡献同样重要！我们欢迎：
- 修复文档中的错误和不准确的地方
- 完善使用指南和教程
- 增加示例和最佳实践
- 翻译为其他语言

文档位于`docs/`目录下，使用Markdown格式。

## 开发规范

### 代码规范
请参考[代码规范文档](docs/CODING_STANDARDS.md)，所有代码必须通过golangci-lint检查。

### 测试规范
- 新增功能必须包含对应的单元测试，覆盖率不低于80%
- 修复Bug必须包含回归测试，防止问题复现
- 集成测试覆盖核心流程，确保功能正常工作
- 性能测试用于验证性能优化效果

### 版本发布
我们采用[语义化版本号](https://semver.org/lang/zh-CN/)：
- 主版本号：不兼容的API变更
- 次版本号：向下兼容的功能新增
- 修订号：向下兼容的问题修正

## 社区交流
- GitHub Issues：用于报告Bug和提出功能请求
- GitHub Discussions：用于讨论设计方案、使用问题等
- 微信群/ Discord：用于实时交流（加入方式请见README）

## 许可证
您提交的所有贡献都将受[MIT许可证](LICENSE)保护。
