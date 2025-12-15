# MachineServer - 统一仿真微服务平台

[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
[![Go Report Card](https://goreportcard.com/badge/github.com/forfire912/machineServer)](https://goreportcard.com/report/github.com/forfire912/machineServer)

## 项目概述

**MachineServer** 是一个基于 Go 语言开发的高性能统一仿真微服务平台。它屏蔽了底层仿真器（QEMU, Renode, OpenOCD）的差异，向上层应用提供统一的 RESTful API，用于构建云端嵌入式开发、自动化测试及系统级协同仿真环境。

## 核心特性与能力矩阵

### 1. 多后端统一适配 (Multi-backend Support)
支持多种主流仿真引擎，通过适配器模式提供统一接口。

| 后端 | 适用场景 | 优势 | 限制 |
| :--- | :--- | :--- | :--- |
| **QEMU** | Linux/Android 仿真, 裸机开发 | 速度快，社区支持广，GDB 支持完善 | 外设模拟精度一般，多核同步较弱 |
| **Renode** | 物联网节点, 复杂 SoC, 多节点组网 | 外设模拟精确，支持时间确定性，脚本能力强 | 性能略低于 QEMU，学习曲线较陡 |
| **OpenOCD** | 硬件在环 (HIL), 真实板卡调试 | 直接操作真实硬件 | 依赖物理连接，无法进行纯软件快照 |

### 2. 能力发现 (Capability Discovery)
动态查询后端支持的硬件特性，支持前端 UI 动态渲染。

*   **能力**:
    *   **动态聚合**: 自动聚合所有已启用后端的元数据。
    *   **多维度信息**: 返回支持的处理器架构 (Processors)、板卡型号 (Boards)、外设类型 (Peripherals) 及总线协议 (BusTypes)。
*   **查询内容**:
    *   **板卡 (Boards)**: 如 `raspi3`, `stm32f4discovery`, `hifive1` 等。
    *   **处理器 (Processors)**: 如 `cortex-m4`, `cortex-a53`, `riscv64` 等。
    *   **外设 (Peripherals)**: 如 `uart`, `ethernet`, `gpio`, `spi` 等。
    *   **总线 (BusTypes)**: 如 `ahb`, `apb`, `axi`, `pci` 等。
*   **可用 API**:
    *   `GET /api/v1/capabilities`: 获取所有后端能力列表
*   **使用示例**:
    ```bash
    curl -X GET http://localhost:8080/api/v1/capabilities
    ```
*   **响应示例**:
    ```json
    [
      {
        "backend": "qemu",
        "boards": ["raspi3", "virt", "versatilepb"],
        "processors": ["cortex-a53", "cortex-m3", "riscv64"],
        "peripherals": ["uart", "virtio-net", "pl011"],
        "bus_types": ["pci", "usb"]
      },
      {
        "backend": "renode",
        "boards": ["stm32f4_discovery", "hifive1"],
        "processors": ["cortex-m4", "riscv32"],
        "features": ["time-travel"]
      }
    ]
    ```

#### 支持的硬件列表 (Supported Hardware List)

**QEMU Backend:**
*   **Boards**: `versatilepb`, `vexpress-a9`, `realview-eb`, `integratorcp`, `mps2-an385`, `mps2-an500`, `mps2-an511`, `stm32vldiscovery`, `stm32f405soc`, `netduino2`, `netduinoplus2`, `microbit`, `nrf51dk`, `raspi2`, `raspi3`, `virt`, `sifive_e`, `sifive_u`, `spike`, `pc`, `q35`, `isapc`
*   **Processors**: `cortex-m0`, `cortex-m3`, `cortex-m4`, `cortex-m7`, `cortex-m33`, `cortex-a7`, `cortex-a8`, `cortex-a9`, `cortex-a15`, `cortex-a53`, `cortex-a57`, `cortex-a72`, `arm926`, `arm1136`, `riscv32`, `riscv64`, `sifive-e31`, `sifive-u54`, `i386`, `x86_64`
*   **Peripherals**: `uart`, `pl011`, `16550a`, `gpio`, `pl061`, `spi`, `ssi`, `i2c`, `timer`, `sp804`, `arm_timer`, `adc`, `ethernet`, `smc91c111`, `lan9118`, `e1000`, `virtio-net`, `display`, `pl110`, `sd`, `pl181`, `sdhci`, `usb`, `usb-ehci`, `usb-ohci`, `virtio-blk`, `virtio-rng`
*   **Bus Types**: `ahb`, `apb`, `axi`, `pci`, `pcie`, `usb`, `i2c`, `spi`

**Renode Backend:**
*   **Boards**: `stm32f4_discovery`, `stm32f746g_disco`, `stm32f072b_disco`, `nucleo_f103rb`, `nucleo_l476rg`, `nrf52840dk`, `nrf52dk`, `microbit`, `hifive1`, `hifive1_revb`, `hifive_unleashed`, `sam_e70_xplained`, `polarfire_soc`, `imxrt1064_evk`, `k64f`, `arduino_uno`, `zedboard`, `pico`
*   **Processors**: `cortex-m0`, `cortex-m0+`, `cortex-m3`, `cortex-m4`, `cortex-m7`, `cortex-m23`, `cortex-m33`, `cortex-a7`, `cortex-a9`, `cortex-a53`, `cortex-a72`, `cortex-r5`, `cortex-r52`, `riscv32`, `riscv64`, `vexriscv`, `rocket`, `ariane`, `ibex`, `sparc`, `ppc`, `xtensa`, `x86`
*   **Peripherals**: `uart`, `usart`, `lpuart`, `gpio`, `spi`, `qspi`, `i2c`, `timer`, `rtc`, `watchdog`, `adc`, `dac`, `can`, `fdcan`, `ethernet`, `gem`, `macb`, `usb`, `usb-otg`, `sd-card`, `sdmmc`, `display`, `ltdc`, `radio`, `nrf-radio`, `ieee802.15.4`, `sensor`, `imu`, `temp-sensor`, `humidity-sensor`, `crypto`, `rng`, `aes`
*   **Bus Types**: `ahb`, `apb`, `axi`, `wishbone`, `pci`, `i2c`, `spi`, `uart`

### 3. 会话管理 (Session Management)
完整的仿真生命周期管理。

*   **能力**:
    *   **创建/删除**: 动态分配端口资源，隔离运行环境。
    *   **板卡配置**: 支持通过 JSON/YAML 动态指定板卡型号 (`board`) 及参数。
    *   **状态查询**: 实时获取会话状态 (Running, Stopped, Error)。
*   **可用 API**:
    *   `POST /api/v1/sessions`: 创建新会话
    *   `GET /api/v1/sessions`: 获取会话列表
    *   `GET /api/v1/sessions/:id`: 获取会话详情
    *   `DELETE /api/v1/sessions/:id`: 销毁会话
*   **使用示例**:
    ```bash
    # 创建会话
    curl -X POST http://localhost:8080/api/v1/sessions \
      -H "Content-Type: application/json" \
      -d '{"backend": "qemu", "board_config": {"board": "virt"}}'

    # 获取会话列表
    curl -X GET http://localhost:8080/api/v1/sessions
    ```
*   **限制**: 单节点最大并发会话数受 `config.yaml` 中 `max_sessions` 限制。

### 4. 板卡配置 (Board Configuration)
灵活的硬件定义系统，支持通过 JSON/YAML 动态配置仿真目标。

*   **能力**:
    *   **预定义板卡**: 直接使用后端支持的标准板卡（如 `raspi3`, `stm32f4discovery`）。
    *   **自定义配置**: 通过 JSON 描述处理器架构、内存映射及外设参数（需后端支持动态构建）。
*   **可用 API**:
    *   集成在 `POST /api/v1/sessions` 中，通过 `board_config` 字段传递。
*   **使用示例**:
    ```bash
    # 简单模式 (使用预定义板卡)
    curl -X POST http://localhost:8080/api/v1/sessions \
      -d '{"backend": "qemu", "board_config": {"board": "raspi3"}}'

    # 高级模式 (自定义参数)
    curl -X POST http://localhost:8080/api/v1/sessions \
      -d '{
        "backend": "renode",
        "board_config": {
          "board": "stm32f4_discovery",
          "cpu": "cortex-m4",
          "memory_map": [{"start": "0x08000000", "size": "1M"}]
        }
      }'
    ```

### 5. 仿真控制 (Simulation Control)
像操作真实开发板一样控制仿真器。

*   **能力**:
    *   **PowerOn**: 启动仿真进程 (Resume)。
    *   **PowerOff**: 优雅停止仿真 (Pause/Stop)。
    *   **Reset**: 复位目标机。
*   **可用 API**:
    *   `POST /api/v1/sessions/:id/poweron`: 启动/恢复仿真
    *   `POST /api/v1/sessions/:id/poweroff`: 停止/暂停仿真
    *   `POST /api/v1/sessions/:id/reset`: 复位仿真
*   **使用示例**:
    ```bash
    # 启动/恢复仿真
    curl -X POST http://localhost:8080/api/v1/sessions/sess_123/poweron

    # 暂停/停止仿真
    curl -X POST http://localhost:8080/api/v1/sessions/sess_123/poweroff

    # 复位
    curl -X POST http://localhost:8080/api/v1/sessions/sess_123/reset
    ```

### 6. 程序加载与调试 (Loading & Debugging)
支持多种格式的固件加载及源码级调试。

*   **能力**:
    *   **程序加载**: 支持 ELF, Binary, Intel HEX 格式。QEMU 使用 GDB Batch 加载，Renode/OpenOCD 使用原生命令。
    *   **GDB 集成**: 标准 GDB Remote Serial Protocol (RSP) 支持。系统自动为每个会话分配独立的 GDB 端口。
    *   **远程调试**: 支持任何兼容 GDB 的 IDE (VS Code, CLion) 或命令行工具进行远程连接。
    *   **实时流**: 通过 WebSocket 实时推送控制台 (UART) 输出、系统日志和状态变更事件。
*   **可用 API**:
    *   `POST /api/v1/programs`: 上传程序文件
    *   `POST /api/v1/sessions/:id/program`: 加载程序到会话
    *   `GET /api/v1/sessions/:id/stream/console`: WebSocket 控制台流
*   **使用示例**:
    ```bash
    # 1. 上传程序文件
    curl -X POST http://localhost:8080/api/v1/programs \
      -F "file=@firmware.elf"

    # 2. 加载程序到会话
    curl -X POST http://localhost:8080/api/v1/sessions/sess_123/program \
      -H "Content-Type: application/json" \
      -d '{"program_path": "/tmp/firmware.elf"}'

    # 3. 连接 GDB (假设端口为 3333)
    # gdb-multiarch firmware.elf -ex "target remote :3333"
    ```

### 7. 快照与恢复 (Snapshot & Restore)
保存仿真现场，用于快速启动或 Bug 复现。

*   **能力**:
    *   **创建快照**: 保存 CPU 寄存器、内存及外设状态。
    *   **恢复快照**: 将仿真器重置到快照点。
*   **支持情况**:
    *   ✅ QEMU (`savevm`/`loadvm`)
    *   ✅ Renode (`save`/`load`)
    *   ❌ OpenOCD (不支持)
*   **可用 API**:
    *   `POST /api/v1/sessions/:id/snapshots`: 创建快照
    *   `POST /api/v1/sessions/:id/restore`: 恢复快照
*   **使用示例**:
    ```bash
    # 创建快照
    curl -X POST http://localhost:8080/api/v1/sessions/sess_123/snapshots \
      -H "Content-Type: application/json" \
      -d '{"name": "boot_complete"}'

    # 恢复快照
    curl -X POST http://localhost:8080/api/v1/sessions/sess_123/restore \
      -H "Content-Type: application/json" \
      -d '{"snapshot_id": "boot_complete"}'
    ```

### 8. 覆盖率分析 (Coverage Analysis)
无需插桩的非侵入式代码覆盖率采集。

*   **能力**:
    *   **采集控制**: 动态开启/停止覆盖率记录。
    *   **报告生成**: 自动生成 LCOV (`.info`) 及 HTML 可视化报告。
    *   **实现原理**:
        *   **Renode**: 使用内置 `cpu LogCoverage` 功能。
        *   **QEMU/OpenOCD**: 基于 Semihosting 机制，由固件触发数据导出。
*   **可用 API**:
    *   `POST /api/v1/sessions/:id/coverage/start`: 开始采集
    *   `POST /api/v1/sessions/:id/coverage/stop`: 停止采集
*   **使用示例**:
    ```bash
    # 开始采集
    curl -X POST http://localhost:8080/api/v1/sessions/sess_123/coverage/start

    # 停止采集 (返回报告路径)
    curl -X POST http://localhost:8080/api/v1/sessions/sess_123/coverage/stop
    ```

### 9. 系统级协同仿真 (Co-Simulation)
支持多节点异构组网仿真。

*   **能力**:
    *   **多节点组网**: 在一个 Co-Sim 会话中管理多个 QEMU/Renode 实例。
    *   **同步方案 3 (Time-slice)**: 基于时间切片的并行同步 (`SyncTime`)，适合松耦合系统。
    *   **同步方案 4 (Event-driven)**: 基于事件注入的交互 (`InjectEvent`)，适合 GPIO/中断触发。
*   **可用 API**:
    *   `POST /api/v1/cosimulation`: 创建协同会话
    *   `GET /api/v1/cosimulation`: 获取协同会话列表
    *   `GET /api/v1/cosimulation/:id`: 获取协同会话详情
    *   `DELETE /api/v1/cosimulation/:id`: 删除协同会话
    *   `POST /api/v1/cosimulation/:id/start`: 启动协同仿真
    *   `POST /api/v1/cosimulation/:id/stop`: 停止协同仿真
    *   `POST /api/v1/cosimulation/:id/sync-step`: 执行指令步进同步
    *   `POST /api/v1/cosimulation/:id/sync-time`: 执行时间切片同步
    *   `POST /api/v1/cosimulation/:id/event`: 注入跨节点事件
*   **使用示例**:
    ```bash
    # 创建协同会话
    curl -X POST http://localhost:8080/api/v1/cosimulation \
      -H "Content-Type: application/json" \
      -d '{
        "nodes": [
          {"backend": "qemu", "board": "virt"},
          {"backend": "renode", "board": "hifive1"}
        ]
      }'

    # 启动协同仿真
    curl -X POST http://localhost:8080/api/v1/cosimulation/cosim_123/start
    ```

### 10. 异步作业队列 (Async Job Queue)
基于 Redis 的高性能异步任务处理系统。

*   **能力**:
    *   **任务解耦**: 将耗时的仿真任务（如长时间运行、覆盖率报告生成）放入后台队列。
    *   **状态追踪**: 实时查询任务执行进度和结果。
    *   **并发控制**: 通过 Worker 池管理并发任务数，防止资源过载。
*   **可用 API**:
    *   `GET /metrics`: 通过 Prometheus 指标 `simulation_jobs_queued` 监控队列深度
*   **配置**: 需在 `config.yaml` 中配置 Redis 连接信息。

### 11. 安全与运维 (Security & Ops)
企业级特性支持。

*   **身份认证**:
    *   **API Key**: 适合 CI/CD 集成 (`Authorization: ApiKey <key>`)。
    *   **JWT**: 适合用户登录 (`Authorization: Bearer <token>`)。
*   **审计日志**: 记录所有 API 操作的用户、IP、时间及动作。
*   **资源配额**:
    *   限制最大并发会话数。
    *   限制上传文件及快照大小。
*   **监控指标**: 集成 Prometheus，提供请求量、延迟、活跃会话数等指标 (`/metrics`)。
*   **可用 API**:
    *   `GET /health`: 健康检查
    *   `GET /metrics`: Prometheus 监控指标
*   **使用示例**:
    ```bash
    # 带 API Key 的请求
    curl -H "Authorization: ApiKey your-secret-key" http://localhost:8080/api/v1/sessions

    # 获取监控指标
    curl http://localhost:8080/metrics
    ```

## 外部工具集成与 IDE 配置 (External Tools & IDE Integration)

为了确保 MachineServer 能正确调度底层仿真器并支持调试，请遵循以下配置要求。

### 1. 后端工具要求 (Backend Requirements)

| 工具 | 版本要求 | 关键配置/依赖 | 备注 |
| :--- | :--- | :--- | :--- |
| **QEMU** | 5.0+ | 需安装对应架构的二进制文件 (如 `qemu-system-arm`, `qemu-system-riscv64`) | 必须支持 `-gdb tcp::port` 参数。建议安装 `qemu-system` 全套包。 |
| **Renode** | 1.12+ | Linux/macOS 下通常依赖 `mono` 运行时 | 需确保 `renode` 命令在 PATH 中可用，或在 `config.yaml` 中指定绝对路径。 |
| **OpenOCD** | 0.11+ | 需配置 USB 设备权限 (udev rules) | 仅用于连接真实硬件。需提供对应板卡的 `.cfg` 文件路径。 |

### 2. IDE 连接指南 (IDE Integration)

MachineServer 会为每个会话分配一个独立的 GDB 端口（在会话详情中返回 `gdb_port`）。您可以使用任何支持 GDB 协议的 IDE 进行连接。

#### Visual Studio Code 配置
使用 `cpptools` (C/C++) 或 `cortex-debug` 插件。

**.vscode/launch.json 示例**:
```json
{
    "version": "0.2.0",
    "configurations": [
        {
            "name": "Debug Remote Session",
            "type": "cppdbg",
            "request": "launch",
            "program": "${workspaceFolder}/build/firmware.elf",
            "MIMode": "gdb",
            "miDebuggerPath": "/usr/bin/gdb-multiarch",
            "miDebuggerServerAddress": "localhost:3333", // 替换为 API 返回的 gdb_port
            "cwd": "${workspaceFolder}",
            "setupCommands": [
                {
                    "description": "Enable pretty-printing for gdb",
                    "text": "-enable-pretty-printing",
                    "ignoreFailures": true
                }
            ]
        }
    ]
}
```

#### CLion 配置
1.  打开 **Run/Debug Configurations**。
2.  添加 **GDB Remote Debug**。
3.  **'target remote' args**: `localhost:3333` (替换为实际端口)。
4.  **Symbol file**: 选择本地编译的 ELF 文件。

#### 命令行 GDB
```bash
# 启动 GDB 并加载符号表
gdb-multiarch build/firmware.elf

# 在 GDB 提示符下连接
(gdb) target remote localhost:3333
(gdb) continue
```

## 快速开始

### 前置要求
*   Go 1.21+
*   QEMU (可选, `qemu-system-arm` 等)
*   Renode (可选)
*   OpenOCD (可选)

### 安装与运行

1.  **克隆仓库**
    ```bash
    git clone https://github.com/forfire912/machineServer.git
    cd machineServer
    ```

2.  **配置**
    复制并修改配置文件：
    ```bash
    cp configs/config.yaml config.yaml
    # 编辑 config.yaml 设置后端路径及认证信息
    ```

3.  **运行**
    ```bash
    go run cmd/server/main.go
    ```
    服务默认监听 `:8080`。

### Docker 部署

```bash
docker-compose up -d
```

## 未来规划 (Roadmap)

为了进一步提升平台的易用性与企业级能力，后续开发计划如下：

### Phase 1: 易用性与生态扩展 (Usability & Ecosystem)
*   **CLI 命令行工具 (`mctl`)**: 封装 REST API，支持 `mctl run -b raspi3 firmware.elf` 等快捷指令。
*   **Python SDK**: 提供 PyPI 包，方便自动化测试脚本集成 (e.g., `import machineserver`).
*   **Web 控制台 (Dashboard)**: 提供可视化界面，支持会话管理、在线终端 (xterm.js) 及性能监控图表。
*   **CI/CD 插件**: 官方支持 Jenkins, GitHub Actions, GitLab CI 插件，简化流水线配置。

### Phase 2: 高级仿真特性 (Advanced Simulation)
*   **高级网络模拟**:
    *   支持 TAP/TUN 模式，允许仿真设备接入宿主机网络。
    *   集成 VDE (Virtual Distributed Ethernet) 实现复杂的虚拟交换机组网。
*   **性能分析与追踪**:
    *   集成 Perf/FlameGraph 生成 CPU 热点图。
    *   支持指令级追踪 (Instruction Trace) 数据导出。
*   **外设直通 (Passthrough)**: 支持 USB、PCI 设备直通给仿真实例。

### Phase 3: 云原生与大规模集群 (Cloud Native & Scale)
*   **Kubernetes Operator**: 定义 `Simulation` CRD，实现仿真任务在 K8s 集群的自动调度与弹性伸缩。
*   **制品仓库 (Artifact Registry)**: 内置板卡配置 (`.repl`, `.dtb`) 和固件镜像的版本管理。
*   **多租户隔离**: 基于命名空间的资源隔离与计费统计。

## 配置说明 (`config.yaml`)

```yaml
server:
  port: 8080
  mode: debug # debug, release

auth:
  enabled: true
  api_keys: ["your-secret-key"]
  jwt_secret: "your-jwt-secret"

backends:
  qemu:
    enabled: true
    binary: "/usr/bin/qemu-system-arm"
  renode:
    enabled: true
    binary: "/usr/bin/renode"

resources:
  max_sessions: 10
  session_timeout: 3600

monitoring:
  enabled: true
```

## 许可证

MIT License
