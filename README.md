# gonsole

Go 游戏服务器框架，提供完整的 TCP/WebSocket 长链接解决方案

## 项目简介

gonsole 是一个轻量级游戏服务器框架，为实时游戏提供高性能的长链接通信能力。配套的 C# 客户端库 [unicorn](https://github.com/lixianmin/unicorn) 可实现 Unity 游戏与服务器的无缝对接。

**核心能力**：
- **双协议支持**：TCP + WebSocket，满足不同场景需求
- **HTTP 服务**：同时提供 RESTful API 能力
- **远程控制台**：Web 端交互式命令行，用于调试和运维
- **RPC 框架**：基于反射的组件服务注册

## 架构概览

```
┌─────────────────────────────────────────────────────────────┐
│                        HTTP Layer                            │
│         静态资源 │ 控制台页面 │ 日志文件 │ RESTful API        │
└─────────────────────────────────────────────────────────────┘
                              │
┌─────────────────────────────┼───────────────────────────────┐
│                   Long Connection Layer                      │
│         TCP Acceptor              │          WsAcceptor       │
└─────────────────────────────┬────┴────────────┬──────────────┘
                              │                 │
┌─────────────────────────────┼─────────────────┼──────────────┐
│                        Road Layer                            │
│    Session │ Handshake │ Heartbeat │ Kick │ RPC │ Serde      │
└─────────────────────────────┼─────────────────────────────────┘
                              │
┌─────────────────────────────┼─────────────────────────────────┐
│                     Component Layer                          │
│     ConsoleService │ Custom Services (Player, Room, ...)     │
└───────────────────────────────────────────────────────────────┘
                              ▲
                              │ TCP / WebSocket
                              │
┌─────────────────────────────┴─────────────────────────────────┐
│              Unity Client (unicorn)                          │
│              https://github.com/lixianmin/unicorn             │
└───────────────────────────────────────────────────────────────┘
```

## 核心特性

### 长链接服务
- **TCP**：高性能原生协议，适合对延迟敏感的游戏
- **WebSocket**：Web 端兼容，支持 HTTP/2
- **握手协议**：Server → Client 推送配置（心跳间隔、序列化方式、路由表）
- **心跳机制**：Client 主动发送，Server 被动响应
- **踢人机制**：Server 主动断开，附带原因码

### RPC 服务
- **反射注册**：自动提取组件方法，生成路由
- **拦截器**：支持请求前拦截
- **请求/响应模式**：Client.Request() → Server → Client 回调
- **Notify 模式**：Client.Send() → Server（无响应）

### 远程控制台
- **自定义命令**：`server.RegisterCommand(cmd)` 注册业务命令
- **Topic 订阅**：周期性数据推送，如系统 Top 信息
- **权限控制**：`FlagPublic` 公开命令，需认证后执行
- **历史命令**：`history` 查看，`!98` 重复执行
- **Tab 补全**：命令自动补全
- **pprof 集成**：内置性能分析

### 安全认证
- **JWT 令牌**：登录成功后颁发，支持自动登录
- **密码摘要**：SHA256 + Nonce 异或 + Base64

## 环境要求

- Go 1.22+

## 快速开始

### 运行 Demo

```bash
make web
./bin/web
```

或直接运行：

```bash
go run examples/demo.go
```

- 控制台：http://127.0.0.1:8888/console
- TCP 端口：8889（默认）

### 代码示例

```go
package main

import (
	"net/http"

	"github.com/lixianmin/gonsole"
	"github.com/lixianmin/gonsole/road"
)

func main() {
	var webPort = 8888
	var mux = http.NewServeMux()
	var server = gonsole.NewServer(mux,
		gonsole.WithPort(webPort),
		gonsole.WithPageTemplate("console.html"),
		gonsole.WithUserPasswords(map[string]string{"admin": "secret"}),
		gonsole.WithEnablePProf(true),
	)

	// 注册控制台命令
	server.RegisterCommand(&gonsole.Command{
		Name:  "hi",
		Note:  "打印问候语",
		Flag:  gonsole.FlagPublic,
		Handler: func(session road.Session, args []string) (*gonsole.Response, error) {
			return gonsole.NewBeanResponse(map[string]string{"text": "hello"}), nil
		},
	})

	// 注册 RPC 服务
	// server.RegisterService(&PlayerService{})

	http.ListenAndServe(":8888", mux)
}
```

## 项目结构

```
gonsole/
├── road/                 # 网络层核心
│   ├── epoll/            # TCP/WebSocket Acceptor
│   ├── intern/           # 连接封装
│   ├── serde/            # 序列化（JSON等）
│   ├── component/        # RPC 服务注册
│   └── client/           # Go 客户端
├── beans/                # 内置命令实现
├── jwtx/                 # JWT 认证
├── tools/                # 工具函数
├── web/                  # 前端资源
├── examples/             # 示例代码
├── console.go            # 控制台入口
└── Makefile              # 构建脚本
```

## 常用命令

```bash
make test     # 运行测试
make web      # 构建 Web 服务
make build    # 构建项目
make vet      # 静态检查
make fmt      # 代码格式化
make clean    # 清理构建产物
```

## 配套客户端

**Unity/C# 客户端**：[unicorn](https://github.com/lixianmin/unicorn)

```csharp
// Unity 客户端连接示例
var session = new Session();
session.Connect("localhost", 8080, s => new JsonSerde(), 
    onHandShaken: () => Logo.Info("Connected"),
    onClosed: () => Logo.Info("Disconnected")
);

// RPC 调用
session.Call("player.move", new { x = 10, y = 20 }, (response, error) => {
    if (error == null) Logo.Info($"Result: {response}");
});
```

## 技术栈

| 层次 | 技术选型 |
|------|----------|
| 网络库 | gobwas/ws (WebSocket)、net (TCP) |
| 序列化 | JSON（可扩展 Protobuf 等） |
| 认证 | golang-jwt/jwt v5 |
| 系统监控 | shirou/gopsutil |
| 前端 | Solid.js + TypeScript + Vite |

## Roadmap

- [x] JWT + 密码认证
- [x] Solid.js + TypeScript 前端
- [x] Go 1.22+ 泛型支持
- [x] 自研 epoll 替代 gaio
- [ ] HTTPS 支持

## 许可证

[MIT License](LICENSE)
