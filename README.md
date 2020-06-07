## gonsole
基本websocket的golang控制台

-----
#### 0x01 简述

1. 支持自定义控制台command：`server.RegisterCommand(cmd)`
1. 支持自定义topic，订阅后可周期性推送数据到控制台：`server.RegisterTopic(topic)`
1. 支持自定义日志：`gonsole.Init(log)`

----
#### 0x02 Demo
1. 直接运行examples/demo/main.go
1. 在浏览器中输入 http://127.0.0.1:8888/console
1. 按提示在文件框中输入help命令，查看帮助信息
1. 可以通过查看main.go的源代码，学习如何注册command和topic
