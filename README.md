# MachineServer - 统一仿真微服务平台

[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
[![Python 3.8+](https://img.shields.io/badge/python-3.8+-blue.svg)](https://www.python.org/downloads/)

## 项目概述

**MachineServer** 是一个面向多后端的统一仿真微服务平台，旨在提供一致的 RESTful API 接口用于：

- 🖥️ **嵌入式处理器仿真** - 支持多种嵌入式处理器架构的仿真
- ⚙️ **程序执行和调试** - 提供程序运行控制和调试功能
- 📊 **代码覆盖率分析** - 实时代码覆盖率统计和报告
- 🔗 **系统级协同仿真** - 支持多组件协同仿真

## 特性

- 🌐 统一的 RESTful API 接口
- 🔌 支持多种仿真后端
- 📡 实时状态监控和控制
- 📈 详细的性能和覆盖率报告
- 🔧 灵活的配置管理
- 🚀 高性能异步处理

## 快速开始

### 安装

```bash
# 克隆仓库
git clone https://github.com/forfire912/machineServer.git
cd machineServer

# 安装依赖
pip install -r requirements.txt

# 或使用 setup.py 安装
pip install -e .
```

### 运行服务器

```bash
python app.py
```

服务器将在 `http://localhost:5000` 启动

## API 接口文档

### 1. 嵌入式处理器仿真

#### 创建仿真实例
```http
POST /api/v1/simulation/create
Content-Type: application/json

{
  "processor_type": "arm",
  "config": {
    "architecture": "cortex-m4",
    "frequency": 100000000
  }
}
```

#### 启动仿真
```http
POST /api/v1/simulation/{id}/start
```

#### 停止仿真
```http
POST /api/v1/simulation/{id}/stop
```

#### 获取仿真状态
```http
GET /api/v1/simulation/{id}/status
```

### 2. 程序执行和调试

#### 加载程序
```http
POST /api/v1/execution/load
Content-Type: application/json

{
  "simulation_id": "sim_123",
  "program_path": "/path/to/program.elf"
}
```

#### 执行步进
```http
POST /api/v1/execution/{id}/step
```

#### 设置断点
```http
POST /api/v1/execution/{id}/breakpoint
Content-Type: application/json

{
  "address": "0x08000100"
}
```

#### 读取寄存器
```http
GET /api/v1/execution/{id}/registers
```

#### 读取内存
```http
GET /api/v1/execution/{id}/memory?address=0x08000000&size=256
```

### 3. 代码覆盖率分析

#### 开始覆盖率收集
```http
POST /api/v1/coverage/{id}/start
```

#### 获取覆盖率报告
```http
GET /api/v1/coverage/{id}/report
```

#### 导出覆盖率数据
```http
GET /api/v1/coverage/{id}/export?format=json
```

### 4. 系统级协同仿真

#### 创建协同仿真
```http
POST /api/v1/cosimulation/create
Content-Type: application/json

{
  "components": [
    {
      "type": "processor",
      "config": {...}
    },
    {
      "type": "peripheral",
      "config": {...}
    }
  ]
}
```

#### 同步仿真步进
```http
POST /api/v1/cosimulation/{id}/sync-step
```

## 配置

服务器配置可以通过 `config.yaml` 文件进行设置：

```yaml
server:
  host: 0.0.0.0
  port: 5000
  debug: false

simulation:
  max_instances: 10
  timeout: 3600

logging:
  level: INFO
  file: machineserver.log
```

## 项目结构

```
machineServer/
├── app.py                      # 主应用入口
├── config.yaml                 # 配置文件
├── requirements.txt            # Python 依赖
├── setup.py                    # 安装脚本
├── README.md                   # 项目文档
├── machineserver/              # 主包目录
│   ├── __init__.py
│   ├── api/                    # API 路由
│   │   ├── __init__.py
│   │   ├── simulation.py       # 仿真 API
│   │   ├── execution.py        # 执行调试 API
│   │   ├── coverage.py         # 覆盖率 API
│   │   └── cosimulation.py     # 协同仿真 API
│   ├── core/                   # 核心模块
│   │   ├── __init__.py
│   │   ├── simulation_manager.py
│   │   ├── execution_engine.py
│   │   ├── coverage_analyzer.py
│   │   └── cosim_coordinator.py
│   └── utils/                  # 工具函数
│       ├── __init__.py
│       ├── config.py
│       └── logger.py
└── tests/                      # 测试目录
    └── __init__.py
```

## 使用示例

### Python 客户端示例

```python
import requests

# 创建仿真实例
response = requests.post('http://localhost:5000/api/v1/simulation/create', json={
    'processor_type': 'arm',
    'config': {
        'architecture': 'cortex-m4',
        'frequency': 100000000
    }
})
sim_id = response.json()['simulation_id']

# 启动仿真
requests.post(f'http://localhost:5000/api/v1/simulation/{sim_id}/start')

# 获取状态
status = requests.get(f'http://localhost:5000/api/v1/simulation/{sim_id}/status')
print(status.json())
```

## 开发

### 运行测试

```bash
pytest tests/
```

### 代码格式化

```bash
black machineserver/
```

## 贡献

欢迎提交 Issue 和 Pull Request！

## 许可证

本项目采用 MIT 许可证 - 详见 [LICENSE](LICENSE) 文件

## 联系方式

- 项目主页: https://github.com/forfire912/machineServer
- 问题反馈: https://github.com/forfire912/machineServer/issues

---

**MachineServer** - 让嵌入式系统仿真更简单、更强大！
