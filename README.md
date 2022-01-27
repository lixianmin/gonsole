## gonsole
基于websocket的远程控制台系统



-----
#### 0x01 简述

历数各类软件系统，你会发现每一个牛B的系统都会自带一个控制台，用于观察系统状态和调整系统参数，比如Linux, MySQL等。

1. 支持自定义command：`server.RegisterCommand(cmd)`
1. 支持自定义topic，订阅后可周期性推送数据到控制台：`server.RegisterTopic(topic)`
1. 安全验证：对于关键的系统命令，可以设置`cmd.IsPublic=false`，这一类命令只能使用`auth`验证后才能使用
1. 历史命令：输入history查看历史命令，输入 !98 执行历史命令列表中的第98命令
1. Tab键命令补全



----

#### 0x02 基本命令图示

##### 01 帮助中心 help

<img src="https://raw.githubusercontent.com/lixianmin/gonsole/master/res/images/help.png?raw=true"  style="zoom:50%" />



##### 02 日志列表 log.list

<img src="https://raw.githubusercontent.com/lixianmin/gonsole/master/res/images/log.list.png?raw=true"  style="zoom:50%" />



##### 03 命令输入框

<img src="https://raw.githubusercontent.com/lixianmin/gonsole/master/res/images/inputbox.png?raw=true"  style="zoom:50%" />




----
#### 0x03 Demo
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
		Name:     "hi",
		Note:     "打印 hi console",
		IsPublic: false,
		Handler: func(client *gonsole.Client, texts [] string) {
			var bean struct {
				Text string
			}

			bean.Text = "hello world"
			client.SendBean(bean)
		},
	})
}
```

---
#### 0x04 感谢JetBrains的免费License支持

<img src="https://resources.jetbrains.com/storage/products/company/brand/logos/jb_beam.png" style="zoom:20%" />
