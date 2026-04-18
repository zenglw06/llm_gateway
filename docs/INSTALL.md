# 安装部署指南

本文档介绍LLM Gateway的多种安装和部署方式。

## 环境要求

- Go 1.22+ （源码编译时需要）
- Linux/macOS/Windows 操作系统
- 至少2核4G内存（生产环境建议4核8G以上）
- Redis（可选，分布式部署时需要）

## 方式一：二进制部署（推荐）

### 1. 下载二进制包
从 [GitHub Release](https://github.com/your-org/llm-gateway/releases) 页面下载对应平台的最新版本：

```bash
# Linux amd64
wget https://github.com/your-org/llm-gateway/releases/download/v1.0.0/llm-gateway-linux-amd64.tar.gz

# macOS arm64 (Apple Silicon)
wget https://github.com/your-org/llm-gateway/releases/download/v1.0.0/llm-gateway-darwin-arm64.tar.gz

# Windows amd64
wget https://github.com/your-org/llm-gateway/releases/download/v1.0.0/llm-gateway-windows-amd64.zip
```

### 2. 解压
```bash
# Linux/macOS
tar zxvf llm-gateway-*.tar.gz
cd llm-gateway

# Windows
unzip llm-gateway-windows-amd64.zip
cd llm-gateway
```

### 3. 配置
复制示例配置文件并修改：
```bash
cp configs/config.example.yaml configs/config.yaml
```

编辑 `configs/config.yaml`，至少需要配置：
- LLM服务的API Key
- JWT密钥（用于生成和验证API Key）

参考配置示例：
```yaml
server:
  port: 8080
  mode: release

log:
  level: info
  format: json

auth:
  jwt_secret: "your-strong-jwt-secret-key-here" # 请修改为随机字符串

llm:
  services:
    openai:
      type: openai
      api_key: "sk-xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx" # 你的OpenAI API Key
      base_url: "https://api.openai.com/v1" # 可以配置为代理地址
      timeout: 60
      max_retries: 2
      enabled: true
    claude:
      type: claude
      api_key: "sk-ant-xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx" # 你的Claude API Key
      enabled: true
```

### 4. 启动服务
```bash
# 前台启动
./bin/llm-gateway -c configs/config.yaml

# 后台启动
nohup ./bin/llm-gateway -c configs/config.yaml > llm-gateway.log 2>&1 &
```

### 5. 验证服务
```bash
curl http://localhost:8080/health
# 预期输出: {"status":"ok"}
```

## 方式二：Docker 部署

### 1. 拉取镜像
```bash
docker pull llmgateway/llm-gateway:latest
```

### 2. 准备配置文件
```bash
mkdir -p /opt/llm-gateway/configs
wget https://raw.githubusercontent.com/your-org/llm-gateway/main/configs/config.example.yaml -O /opt/llm-gateway/configs/config.yaml
# 编辑配置文件，填写你的API Key等信息
```

### 3. 运行容器
```bash
docker run -d \
  --name llm-gateway \
  -p 8080:8080 \
  -v /opt/llm-gateway/configs:/app/configs \
  llmgateway/llm-gateway:latest
```

### 使用环境变量配置（不需要配置文件）
```bash
docker run -d \
  --name llm-gateway \
  -p 8080:8080 \
  -e AUTH_JWT_SECRET="your-strong-jwt-secret-key" \
  -e LLM_SERVICES_OPENAI_API_KEY="sk-xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx" \
  -e LLM_SERVICES_OPENAI_ENABLED="true" \
  llmgateway/llm-gateway:latest
```

### 4. 查看日志
```bash
docker logs -f llm-gateway
```

## 方式三：Docker Compose 部署

### 1. 下载docker-compose.yml
```bash
wget https://raw.githubusercontent.com/your-org/llm-gateway/main/deploy/docker-compose.yml
```

### 2. 配置环境变量
创建 `.env` 文件：
```env
AUTH_JWT_SECRET=your-strong-jwt-secret-key
LLM_SERVICES_OPENAI_API_KEY=sk-xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx
LLM_SERVICES_OPENAI_ENABLED=true
```

### 3. 启动服务
```bash
docker-compose up -d
```

### 包含Redis的部署（用于分布式部署）
如果需要使用Redis存储，使用`docker-compose-with-redis.yml`：
```bash
wget https://raw.githubusercontent.com/your-org/llm-gateway/main/deploy/docker-compose-with-redis.yml
docker-compose -f docker-compose-with-redis.yml up -d
```

## 方式四：Kubernetes 部署（Helm）

### 1. 添加Helm仓库
```bash
helm repo add llm-gateway https://helm.llm-gateway.io
helm repo update
```

### 2. 配置values.yaml
```yaml
replicaCount: 3

image:
  repository: llmgateway/llm-gateway
  tag: latest

service:
  type: LoadBalancer
  port: 80

config:
  auth:
    jwtSecret: "your-strong-jwt-secret-key"
  llm:
    services:
      openai:
        type: openai
        apiKey: "sk-xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx"
        enabled: true

redis:
  enabled: true
  storage:
    type: redis

ingress:
  enabled: true
  hosts:
    - host: llm-gateway.example.com
      paths:
        - path: /
```

### 3. 安装
```bash
helm install llm-gateway llm-gateway/llm-gateway -f values.yaml
```

## 方式五：源码编译部署

### 1. 克隆代码
```bash
git clone https://github.com/your-org/llm-gateway.git
cd llm-gateway
```

### 2. 编译
```bash
# 编译当前平台版本
make build

# 编译所有平台版本
make build-all
```

编译后的二进制文件在 `bin/` 目录下。

### 3. 配置和运行
参考二进制部署方式的步骤3-5。

## 生产环境部署建议

### 1. 高可用部署
- 至少部署2个实例，避免单点故障
- 使用负载均衡（Nginx、云厂商LB等）分发流量
- 使用Redis作为共享存储，支持多实例共享限流、配额等数据

### 2. 配置优化
```yaml
server:
  mode: release # 生产环境必须使用release模式
  read_timeout: 120 # 根据实际场景调整超时时间
  write_timeout: 120
  max_header_bytes: 1048576

log:
  level: info # 生产环境不要使用debug模式，会影响性能
  format: json # 便于日志收集和分析
  output_path: stdout # 容器化部署输出到stdout即可
```

### 3. 可观测性配置
- 开启Metrics插件，暴露`/metrics`端点，接入Prometheus监控
- 配置Grafana仪表盘，监控QPS、延迟、错误率、Token用量等指标
- 开启链路追踪插件，接入Jaeger/Zipkin，便于问题排查
- 日志接入ELK/ Loki等日志系统，方便检索和分析

### 4. 安全配置
- API服务不要直接暴露在公网，建议放在内网或通过WAF接入
- 配置API Key的IP白名单限制
- 定期轮换API Key和JWT密钥
- 开启请求日志，便于安全审计
- 配置合理的限流策略，防止恶意攻击和滥用

### 5. 性能优化
- 根据服务器配置调整GOMAXPROCS，一般设置为CPU核数
- 配置合理的连接池大小：`llm.services.*.max_idle_conns`、`max_idle_conns_per_host`
- 开启缓存插件，降低重复请求的LLM调用成本
- 如果部署在K8s中，配置合理的资源请求和限制：
```yaml
resources:
  requests:
    cpu: "2"
    memory: "4Gi"
  limits:
    cpu: "4"
    memory: "8Gi"
```

## 常见部署问题排查

### 1. 服务启动失败
- 检查配置文件格式是否正确
- 检查端口是否被占用：`lsof -i :8080`
- 查看日志输出：`./llm-gateway -c configs/config.yaml` 前台启动查看错误信息

### 2. LLM服务调用失败
- 检查API Key是否正确
- 检查网络是否能访问对应LLM服务的API地址
- 可以配置代理：`llm.services.*.proxy = "http://your-proxy:port"`
- 查看日志中的错误信息，确认是网关问题还是LLM服务问题

### 3. 性能问题
- 检查服务器的CPU、内存、网络带宽是否足够
- 检查是否开启了debug日志，生产环境建议关闭
- 调整连接池配置，增加最大连接数
- 对于高并发场景，建议配置Redis存储，避免内存存储的性能瓶颈

### 4. 流式响应问题
- 确保Nginx/Ingress等反向代理配置了正确的流式响应参数：
```nginx
proxy_http_version 1.1;
proxy_set_header Connection "";
proxy_buffering off;
proxy_cache off;
chunked_transfer_encoding on;
```
- 检查是否配置了过短的超时时间，流式请求可能需要较长的超时时间

## 开机自启配置（Systemd）

### 1. 创建Systemd服务文件
```ini
# /etc/systemd/system/llm-gateway.service
[Unit]
Description=LLM Gateway Service
After=network.target

[Service]
Type=simple
User=root
WorkingDirectory=/opt/llm-gateway
ExecStart=/opt/llm-gateway/bin/llm-gateway -c /opt/llm-gateway/configs/config.yaml
Restart=always
RestartSec=5
LimitNOFILE=65535

[Install]
WantedBy=multi-user.target
```

### 2. 启动并启用服务
```bash
systemctl daemon-reload
systemctl start llm-gateway
systemctl enable llm-gateway
```

### 3. 查看服务状态
```bash
systemctl status llm-gateway
journalctl -u llm-gateway -f # 查看实时日志
```
