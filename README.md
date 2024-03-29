## gonsole
基于websocket的远程控制台系统



-----
#### 0x1 简述

历数各类软件系统，你会发现每一个牛B的系统都会自带一个控制台，用于观察系统状态和调整系统参数，比如Linux, MySQL等。

1. 支持自定义command：`server.RegisterCommand(cmd)`
1. 支持自定义topic，订阅后可周期性推送数据到控制台：`server.RegisterTopic(topic)`
1. 安全验证：对于关键的系统命令，可以设置`cmd.IsPublic=false`，这一类命令只能使用`auth`验证后才能使用
1. 历史命令：输入history查看历史命令，输入 !98 执行历史命令列表中的第98命令
1. Tab键命令补全



----

#### 0x2 基本命令图示

##### 01 帮助中心 help

<img src="https://raw.githubusercontent.com/lixianmin/gonsole/master/res/images/help.png?raw=true"  style="zoom:50%" />



##### 02 日志列表 log.list

<img src="https://raw.githubusercontent.com/lixianmin/gonsole/master/res/images/log.list.png?raw=true"  style="zoom:50%" />



##### 03 命令输入框

<img src="https://raw.githubusercontent.com/lixianmin/gonsole/master/res/images/inputbox.png?raw=true"  style="zoom:50%" />




----
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



----

#### 0x4 Road Map

1. 引入完整的登录验证方式
2. ~~将项目中的js逐步过渡为Vue框架, 目标是梳理代码框架, 通过import减小代码单元的大小~~
3. ~~逐步使用typescript代替javascript, 引入编译机制~~
4. 升级golang以引入泛型机制. 但这件事情在centos的yum默认支持到1.18+之前不能考虑. 目前(2022-09-03) 最新版本是golang 1.19。(centos目前官方不再更新)
5. ~~逐步移除gaio这个库, 在golang 1.17+的centos上编译会报错, 它升级太慢了. [相关issue](https://github.com/xtaci/gaio/issues/21)~~
6. 引入对https的支持, 或者至少设计出完整的支持方案. 在golang库中直接支持可能比在nginx上支持要更下简单一些, 毕竟会减少对nginx的依赖. ~~另外, gaio这个库似乎不支持https~~


---
#### 0x5 感谢JetBrains的免费License支持

<img src="https://resources.jetbrains.com/storage/products/company/brand/logos/jb_beam.png" style="zoom:20%" />
