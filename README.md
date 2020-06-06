## gonsole
基本websocket的golang控制台

-----
#### 0x01 简述

1. 支持自定义控制台command：`server.RegisterCommand(cmd)`
1. 支持自定义topic，订阅后可周期性推送数据到控制台：`server.RegisterTopic(topic)`
1. 支持自定义日志：`gonsole.Init(log)`

----
#### 0x02 todo list
2. 在console.html页面，回车应该转到input box中
