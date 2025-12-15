# Machine Server - 统一仿真微服务平台

Machine Server 是一个面向多后端（QEMU、Renode）的统一仿真微服务平台，提供一致的 RESTful API 接口用于嵌入式处理器仿真、程序执行和调试、代码覆盖率分析以及系统级协同仿真。

## 特性

### 核心功能

- **多后端支持**: 统一接口支持 QEMU 和 Renode 仿真器
- **能力发现**: 动态查询支持的处理器、外设和总线类型
- **会话管理**: 完整的会话生命周期管理（创建、查询、删除）
- **板卡配置**: 灵活的 JSON/YAML 硬件配置系统
- **仿真控制**: PowerOn/Off、Reset 等基本控制功能
- **程序管理**: 支持 ELF、Binary、Intel HEX 格式

### 调试功能

- **GDB 集成**: 标准 GDB 协议支持
- **远程调试**: 通过 GDB Remote Serial Protocol 进行调试

### 高级功能

- **实时流**: WebSocket 实时推送控制台输出、日志和状态变更
- **快照/恢复**: 仿真状态保存和恢复
- **覆盖率分析**: 支持基本块、函数和分支覆盖率，输出 LCOV 和 HTML 报告
- **作业队列**: 基于 Redis 的异步任务处理系统
- **系统级仿真**: 支持多节点协同仿真

### 安全与运维

- **身份认证**: API Key 和 JWT Token 认证
- **审计日志**: 完整的操作审计记录
- **资源配额**: CPU、内存、磁盘配额管理
- **监控指标**: Prometheus 集成，提供 API 请求量、响应时间等指标

## 快速开始

### 前置要求

- Go 1.21+
- Redis (可选，用于作业队列)
- QEMU 和/或 Renode (根据需要)

### 安装

```bash
# 克隆仓库
git clone https://github.com/forfire912/machineServer.git
cd machineServer

# 下载依赖
go mod download

# 构建
make build

# 运行
make run
```

### 使用 Docker

```bash
# 构建镜像
docker build -t machineserver:latest .

# 运行
docker-compose up -d
```

### 使用 Kubernetes

```bash
# 应用配置
kubectl apply -f deployments/kubernetes/

# 检查状态
kubectl get pods
kubectl get services
```

## API 文档

### 基础端点

#### 健康检查
```bash
GET /health
```

#### 获取后端能力
```bash
GET /api/v1/capabilities
```

### 会话管理

#### 创建会话
```bash
POST /api/v1/sessions
Content-Type: application/json

{
  "name": "my-session",
  "backend": "qemu",
  "board_config": {
    "processor": {
      "model": "cortex-m3",
      "frequency": 72000000
    },
    "memory": {
      "flash": {
        "base": 134217728,
        "size": 131072
      },
      "ram": {
        "base": 536870912,
        "size": 20480
      }
    }
  }
}
```

#### 列出会话
```bash
GET /api/v1/sessions?page=1&page_size=10
```

#### 获取会话详情
```bash
GET /api/v1/sessions/{id}
```

#### 删除会话
```bash
DELETE /api/v1/sessions/{id}
```

### 仿真控制

#### 上电
```bash
POST /api/v1/sessions/{id}/poweron
```

#### 断电
```bash
POST /api/v1/sessions/{id}/poweroff
```

#### 重置
```bash
POST /api/v1/sessions/{id}/reset
```

### 程序管理

#### 上传程序
```bash
POST /api/v1/programs
Content-Type: multipart/form-data

file: <binary file>
name: "my-program"
format: "elf"
```

#### 加载程序到会话
```bash
POST /api/v1/sessions/{id}/program
Content-Type: application/json

{
  "program_id": "program-uuid"
}
```

### 快照管理

#### 创建快照
```bash
POST /api/v1/sessions/{id}/snapshots
Content-Type: application/json

{
  "name": "checkpoint-1",
  "description": "Before critical operation"
}
```

#### 恢复快照
```bash
POST /api/v1/sessions/{id}/restore
Content-Type: application/json

{
  "snapshot_id": "snapshot-uuid"
}
```

### WebSocket 流

#### 连接控制台输出流
```bash
WS /api/v1/sessions/{id}/stream/console
```

## 配置

配置文件位于 `configs/config.yaml`：

```yaml
server:
  host: "0.0.0.0"
  port: 8080
  mode: "release"

auth:
  enabled: true
  jwt_secret: "your-secret-key"
  api_keys:
    - "your-api-key"

backends:
  qemu:
    enabled: true
    binary: "qemu-system-arm"
  renode:
    enabled: true
    binary: "renode"

resources:
  max_sessions: 100
  max_memory_mb: 4096
  session_timeout: 3600

monitoring:
  enabled: true
  prometheus_port: 9090
```

## 认证

### API Key 认证

```bash
curl -H "Authorization: ApiKey your-api-key" \
  http://localhost:8080/api/v1/capabilities
```

### JWT Token 认证

```bash
curl -H "Authorization: Bearer your-jwt-token" \
  http://localhost:8080/api/v1/sessions
```

## 监控

Prometheus 指标端点：

```bash
GET /metrics
```

主要指标：
- `http_requests_total` - HTTP 请求总数
- `http_request_duration_seconds` - HTTP 请求持续时间
- `simulation_active_sessions` - 活跃会话数
- `simulation_programs_uploaded_total` - 上传程序总数
- `simulation_jobs_queued` - 队列中的作业数

## 项目结构

```
machineServer/
├── cmd/
│   └── server/          # 主程序入口
├── internal/
│   ├── adapter/         # 后端适配器（QEMU、Renode）
│   ├── api/             # HTTP API 处理器和路由
│   ├── config/          # 配置管理
│   ├── model/           # 数据模型
│   └── service/         # 业务逻辑层
├── pkg/
│   ├── coverage/        # 覆盖率分析
│   ├── gdb/             # GDB 服务器
│   └── queue/           # 作业队列
├── configs/             # 配置文件
├── deployments/         # 部署配置
│   └── kubernetes/      # K8s 部署文件
├── Dockerfile
├── docker-compose.yml
└── Makefile
```

## 开发

### 构建

```bash
make build
```

### 运行测试

```bash
make test
```

### 代码格式化

```bash
make fmt
```

### 清理

```bash
make clean
```

## 许可证

MIT License

## 贡献

欢迎提交 Issue 和 Pull Request！
