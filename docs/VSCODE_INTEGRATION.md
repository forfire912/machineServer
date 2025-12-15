# VS Code 集成指南

本文档详细说明如何将 MachineServer 的实时控制台流（WebSocket）集成到 VS Code 插件中，实现仿真器输出的实时显示和交互。

## 1. 概述

MachineServer 通过 WebSocket 提供实时的控制台输出流。VS Code 插件可以通过连接此 WebSocket，将仿真器（QEMU/Renode/OpenOCD）的 stdout/stderr 或串口输出显示在 VS Code 的界面上。

- **WebSocket URL**: `ws://<host>:<port>/api/v1/sessions/{session_id}/stream/console`
- **数据方向**: 双向（目前主要用于下行输出，支持上行输入扩展）
- **数据格式**: 原始文本/二进制流

## 2. 准备工作

在您的 VS Code 插件项目（Node.js 环境）中，建议安装 `ws` 库来处理 WebSocket 连接。

```bash
npm install ws
npm install --save-dev @types/ws
```

## 3. 方案一：使用 OutputChannel (只读日志)

适用于只需要查看日志输出，不需要交互或颜色支持的场景。

### 实现代码

```typescript
import * as vscode from 'vscode';
import WebSocket from 'ws';

let outputChannel: vscode.OutputChannel;
let activeSocket: WebSocket | undefined;

export function activate(context: vscode.ExtensionContext) {
    // 1. 创建输出面板
    outputChannel = vscode.window.createOutputChannel("MachineServer Console");
    
    // 2. 注册连接命令
    let disposable = vscode.commands.registerCommand('machineserver.connectConsole', async () => {
        const sessionId = await vscode.window.showInputBox({
            placeHolder: "Enter Session ID",
            prompt: "Connect to simulation console stream"
        });

        if (sessionId) {
            connectToStream(sessionId);
        }
    });

    context.subscriptions.push(disposable);
}

function connectToStream(sessionId: string) {
    // 清理旧连接
    if (activeSocket) {
        activeSocket.close();
    }

    // 替换为实际的服务器地址
    const wsUrl = `ws://localhost:8080/api/v1/sessions/${sessionId}/stream/console`;
    
    outputChannel.show(true);
    outputChannel.appendLine(`[System] Connecting to ${wsUrl}...`);

    try {
        activeSocket = new WebSocket(wsUrl);

        activeSocket.on('open', () => {
            outputChannel.appendLine("[System] Connected to console stream.");
        });

        activeSocket.on('message', (data: any) => {
            // 将接收到的数据直接写入输出面板
            outputChannel.append(data.toString());
        });

        activeSocket.on('error', (error) => {
            outputChannel.appendLine(`[System] Error: ${error.message}`);
        });

        activeSocket.on('close', () => {
            outputChannel.appendLine("[System] Connection closed.");
            activeSocket = undefined;
        });

    } catch (e) {
        vscode.window.showErrorMessage(`Failed to connect: ${e}`);
    }
}

export function deactivate() {
    if (activeSocket) {
        activeSocket.close();
    }
}
```

## 4. 方案二：使用 Terminal (交互式 + 颜色支持)

适用于需要与仿真器 Shell 交互，或者需要显示 ANSI 颜色代码（如 Linux 启动日志）的场景。

### 实现代码

```typescript
import * as vscode from 'vscode';
import WebSocket from 'ws';

export function activate(context: vscode.ExtensionContext) {
    let disposable = vscode.commands.registerCommand('machineserver.openTerminal', async () => {
        const sessionId = await vscode.window.showInputBox({
            placeHolder: "Enter Session ID",
            prompt: "Open simulation terminal"
        });

        if (sessionId) {
            createSimulationTerminal(sessionId);
        }
    });

    context.subscriptions.push(disposable);
}

function createSimulationTerminal(sessionId: string) {
    const wsUrl = `ws://localhost:8080/api/v1/sessions/${sessionId}/stream/console`;
    let socket: WebSocket | undefined;

    // 创建伪终端 (Pseudoterminal)
    const writeEmitter = new vscode.EventEmitter<string>();
    const closeEmitter = new vscode.EventEmitter<number>();

    const pty: vscode.Pseudoterminal = {
        onDidWrite: writeEmitter.event,
        onDidClose: closeEmitter.event,
        
        open: () => {
            writeEmitter.fire(`\x1b[36mConnecting to session ${sessionId}...\x1b[0m\r\n`);
            
            try {
                socket = new WebSocket(wsUrl);

                socket.on('open', () => {
                    writeEmitter.fire('\x1b[32mConnected.\x1b[0m\r\n');
                });

                socket.on('message', (data: any) => {
                    // 写入终端，支持 ANSI 颜色代码
                    // 注意：需要确保换行符是 \r\n
                    let text = data.toString();
                    if (text.indexOf('\r') === -1) {
                        text = text.replace(/\n/g, '\r\n');
                    }
                    writeEmitter.fire(text);
                });

                socket.on('close', () => {
                    writeEmitter.fire('\r\n\x1b[31mConnection closed.\x1b[0m\r\n');
                    closeEmitter.fire(0);
                });

                socket.on('error', (err) => {
                    writeEmitter.fire(`\r\n\x1b[31mError: ${err.message}\x1b[0m\r\n`);
                });

            } catch (e) {
                writeEmitter.fire(`Error: ${e}\r\n`);
            }
        },
        
        close: () => {
            if (socket) {
                socket.close();
            }
        },
        
        handleInput: (data: string) => {
            // 处理用户输入，发送回服务器（如果后端支持输入）
            if (socket && socket.readyState === WebSocket.OPEN) {
                socket.send(data);
            }
        }
    };

    // 创建并显示终端
    const terminal = vscode.window.createTerminal({ 
        name: `Sim: ${sessionId.substring(0, 8)}`, 
        pty 
    });
    terminal.show();
}
```

## 5. 最佳实践

1.  **自动重连**: 可以在 `close` 事件中实现指数退避重连机制，以应对网络波动。
2.  **心跳检测**: 虽然 WebSocket 协议有 Ping/Pong，但在应用层实现心跳可以更可靠地检测死链接。
3.  **多会话管理**: 建议使用 Map 管理多个会话的连接，避免资源泄漏。
4.  **认证集成**: 如果 MachineServer 开启了认证，连接 WebSocket 时需要在 URL 参数或协议头中携带 Token。
    *   URL 方式: `ws://...?token=YOUR_JWT_TOKEN` (需要后端支持)
    *   Header 方式: `ws` 库支持自定义 Headers。

```typescript
const socket = new WebSocket(wsUrl, {
    headers: {
        'Authorization': `Bearer ${token}`
    }
});
```
