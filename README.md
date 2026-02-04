## gonsole

基于 WebSocket 的远程控制台系统，遵循[项目开发宪法](specs/constitution.md)的严格架构规范。

---

#### 0x1 简述

历数各类软件系统，你会发现每一个牛B的系统都会自带一个控制台，用于观察系统状态和调整系统参数，比如 Linux, MySQL 等。

**核心功能**：
1. 支持自定义 Command：`server.RegisterCommand(cmd)`
2. 支持自定义 Topic，订阅后可周期性推送数据：`server.RegisterTopic(topic)`
3. 安全验证：关键命令设置 `FlagPublic`，需 `auth` 验证后使用
4. 历史命令：输入 `history` 查看，`!98` 执行第 98 条
5. Tab 键命令补全
6. 内置 pprof 性能分析

**架构特点**（符合[项目开发宪法](specs/constitution.md)）：
- **无全局可变状态**：所有依赖显式注入（重构后 Session ID 生成器、Git 构建信息等均已组件化）
- **单一职责**：Console 采用 Facade 模式，职责委托给 CommandManager / TopicManager
- **显式错误处理**：无错误被静默忽略，均有日志或返回
- **表格驱动测试**：单元测试采用标准表格驱动风格
- **标准库优先**：最小化外部依赖，使用 `net/http` 等标准库

---

#### 0x2 基本命令图示

##### 01 帮助中心 help

<img src="https://raw.githubusercontent.com/lixianmin/gonsole/master/res/images/help.png?raw=true"  style="zoom:50%" />

##### 02 日志列表 log.list

<img src="https://raw.githubusercontent.com/lixianmin/gonsole/master/res/images/log.list.png?raw=true"  style="zoom:50%" />

##### 03 命令输入框

<img src="https://raw.githubusercontent.com/lixianmin/gonsole/master/res/images/inputbox.png?raw=true"  style="zoom:50%" />

---

#### 0x3 Demo

1. 直接运行examples/demo/main.go
1. 在浏览器中输入 http://127.0.0.1:8888/console
1. 按提示在文件框中输入help命令，查看帮助信息
1. 可以通过查看main.go的源代码，学习如何注册command和topic

部分代码如下：

```go
func main() {
	var webPort = 8888
	var mux = http.NewServeMux()
	var server = gonsole.NewServer(mux,
		gonsole.WithPort(webPort),                                      // webserver端口
		gonsole.WithPageTemplate("console.html"),                       // 页面文件模板
		gonsole.WithUserPasswords(map[string]string{"xmli": "123456"}), // 认证使用的用户名密码
		gonsole.WithEnablePProf(true),                                  // 开启pprof
	)

	server.RegisterCommand(&gonsole.Command{
		Name:  "hi",
		Note:  "打印 hi console",
		Flag:  gonsole.FlagPublic, // 或 0 表示需认证
		Handler: func(session road.Session, args []string) (*gonsole.Response, error) {
			var bean = struct {
				Text string `json:"text"`
			}{Text: "hello world"}
			return gonsole.NewBeanResponse(bean), nil
		},
	})
}
```

---

#### 0x4 架构设计

**Console 采用 Facade 模式，职责清晰分离**：

```
Console (Facade)
├── CommandManager   # 命令注册与查找
├── TopicManager     # 主题订阅与推送
└── road.App         # WebSocket 网络层
```

**显式依赖注入示例**：

```go
// Manager 持有 Session ID 生成器（非全局变量）
type Manager struct {
    idGenerator atomic.Int64  // 每个实例独立
}

// ConsoleService 通过构造函数注入依赖
type ConsoleService struct {
    console *Console  // 显式注入，非全局获取
}
```

**规范文档**：
- [项目开发宪法](specs/constitution.md) - 核心开发原则
- [重构规范](specs/01.refactor/01.refactor.spec.md) - BDD 行为规范
- [技术方案](specs/01.refactor/01.refactor.plan.md) - 实现细节

---

#### 0x5 Road Map

1. ~~引入完整的登录验证方式~~ ✅ JWT + 密码认证已实现
2. ~~将项目中的 js 逐步过渡为 Vue 框架~~ ✅ 已完成
3. ~~逐步使用 TypeScript 代替 JavaScript~~ ✅ 已完成
4. ~~升级 golang 以引入泛型机制~~ ✅ 已要求 Go 1.22+
5. ~~逐步移除 gaio 库~~ ✅ 已替换为自研 epoll
6. 引入对 HTTPS 的支持，或设计完整支持方案
7. ~~重构消除全局变量~~ ✅ 已完成（详见 specs/01.refactor/）
8. ~~Console 职责拆分~~ ✅ 已完成（Facade + Manager 模式）

**环境要求**：Go 1.22+
