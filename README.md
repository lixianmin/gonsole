## gonsole
基于websocket的远程控制台系统



-----
#### 0x01 简述

历数各类软件系统，你会发现每一个牛B的系统都会自带一个控制台，用于观察系统状态和调整系统参数，比如Linux, MySQL等。

1. 支持自定义command：`server.RegisterCommand(cmd)`
1. 支持自定义topic，订阅后可周期性推送数据到控制台：`server.RegisterTopic(topic)`
1. 安全验证：对于关键的系统命令，可以设置`cmd.IsPublic=false`，这一类命令只能使用`auth`验证后才能使用



----

#### 0x02 基本命令图示

##### 01 help命令

<img src="https://raw.githubusercontent.com/lixianmin/gonsole/master/images/help.png?raw=true"  style="zoom:70%" />



----
#### 0x03 Demo
1. 直接运行examples/demo/main.go
1. 在浏览器中输入 http://127.0.0.1:8888/console
1. 按提示在文件框中输入help命令，查看帮助信息
1. 可以通过查看main.go的源代码，学习如何注册command和topic
