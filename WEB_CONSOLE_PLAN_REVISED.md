# Web Console Development Plan (Revised) - 仿真基础设施管理平台

本文档根据最新需求进行了重新规划。Web 控制台将聚焦于 **“仿真环境管理”** 与 **“板卡可视化配置”**，剥离具体的调试功能（交由 IDE 完成）。

**核心理念**:
*   **Web Console (管理平面)**: 负责基础设施的定义、资源的分配、环境的生命周期管理以及与 IDE 的连接桥梁。
*   **IDE (数据平面)**: 负责代码编写、编译、固件加载、断点调试及单步执行。

## 1. 功能架构 (Functional Architecture)

### 1.1 核心模块
1.  **板卡设计器 (Visual Board Designer)**: 可视化配置硬件参数，生成机器定义。
2.  **环境实例管理 (Instance Management)**: 仿真会话的创建、监控与销毁。
3.  **连接向导 (Connection Wizard)**: 为 IDE 提供连接信息和配置代码片段。

## 2. 详细功能规划

### 2.1 板卡设计器 (Visual Board Designer) - **新增核心**
提供图形化界面来定义 `board_config`，屏蔽底层 JSON 细节。

*   **预设选型**:
    *   从 `GET /capabilities` 获取支持的 `Boards` 列表。
    *   用户可选择 "基于现有板卡修改" 或 "从头创建"。
*   **可视化配置项**:
    *   **处理器 (CPU)**: 下拉选择架构 (e.g., Cortex-M4, RISC-V)。
    *   **内存映射 (Memory Map)**:
        *   交互式表格/块图：定义 RAM/ROM 区域。
        *   输入：起始地址 (Hex), 大小 (Size), 权限 (RWX)。
        *   *校验*: 自动检测地址冲突。
    *   **外设 (Peripherals)**:
        *   组件库：列出可用外设 (UART, ETH, GPIO)。
        *   拖拽/添加：将外设映射到特定地址空间。
        *   属性配置：设置中断号 (IRQ)、时钟频率等。
*   **配置预览与导出**:
    *   实时显示生成的 JSON/YAML 配置。
    *   支持 "保存为模板" (保存到 LocalStorage 或后端数据库)。

### 2.2 仿真环境管理 (Environment Management)
*   **实例列表**:
    *   展示运行中的仿真器实例。
    *   状态指示: `Provisioning`, `Running`, `Stopped`, `Error`.
    *   资源监控: CPU/内存占用 (如果后端支持)。
*   **生命周期控制**:
    *   **启动 (Provision)**: 使用选定的板卡模板启动实例。
    *   **电源控制**: Power On / Power Off / Reset。
    *   **快照管理**: 创建环境快照，以便快速回滚环境状态。
*   **日志监控**:
    *   查看仿真器本身的运行日志 (Standard Output)，用于排查启动失败原因（非串口输出）。

### 2.3 IDE 连接桥梁 (IDE Bridge)
帮助用户快速将本地 IDE 连接到远程仿真环境。

*   **连接信息面板**:
    *   **GDB 端口**: 显示分配的端口 (e.g., `3333`)。
    *   **服务地址**: 显示服务器 IP/域名。
*   **配置生成器 (Config Generator)**:
    *   **VS Code**: 自动生成 `.vscode/launch.json` 内容。
        ```json
        {
            "type": "cppdbg",
            "miDebuggerServerAddress": "192.168.1.100:3333",
            ...
        }
        ```
    *   **CLion / GDB CLI**: 生成对应的连接命令字符串。
    *   **一键复制**: 提供 "Copy Config" 按钮。

## 3. 交互流程 (User Flow)

1.  **配置阶段**: 用户进入 "板卡设计器" -> 选择 "Renode" -> 配置 CPU/内存/外设 -> 保存为 "My-IoT-Board"。
2.  **启动阶段**: 进入 "实例管理" -> 点击 "新建实例" -> 选择 "My-IoT-Board" -> 系统分配资源并启动 -> 状态变为 "Running"。
3.  **连接阶段**: 点击实例详情 -> 查看 "IDE 连接信息" -> 复制 VS Code 配置 -> 粘贴到本地项目。
4.  **调试阶段**: (在本地 IDE 中) 编译代码 -> 点击 Debug -> IDE 通过 GDB 连接到 Web Console 管理的实例 -> 开始调试。

## 4. API 需求分析 (API Requirements)

为了支持上述功能，现有 API 可能需要以下增强（或在前端组合调用）：

| 功能 | 现有 API | 需求/备注 |
| :--- | :--- | :--- |
| **获取组件库** | `GET /capabilities` | 需确保返回足够详细的元数据（如外设的可配置属性）。 |
| **保存板卡模板** | 无 | 暂时存放在前端 LocalStorage，未来可增加 `POST /board-templates`。 |
| **创建实例** | `POST /sessions` | 支持传入完整的 JSON `board_config`。 |
| **获取连接信息** | `GET /sessions/:id` | 响应需明确包含 `gdb_port` 和宿主机 IP（或前端配置）。 |

## 5. 开发阶段 (Phases)

### Phase 1: 基础管理与连接 (MVP)
*   实现 **实例列表** 和 **生命周期控制** (Start/Stop/Delete)。
*   实现 **IDE 连接信息面板** (展示 GDB 端口和 VS Code 配置片段)。
*   仅支持选择 **预定义板卡** (Pre-defined Boards) 启动。

### Phase 2: 可视化配置 (Visual Config)
*   开发 **板卡设计器** UI。
*   实现内存映射和外设配置的表单逻辑。
*   支持使用自定义配置启动会话。

### Phase 3: 模板与持久化
*   实现板卡配置的保存与管理 (Templates)。
*   优化用户体验 (地址冲突检测、配置校验)。
