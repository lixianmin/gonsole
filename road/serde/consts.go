package serde

/********************************************************************
created:    2023-06-05
author:     lixianmin

Copyright (C) - All Rights Reserved
*********************************************************************/

const (
	Handshake = 1    // 连接建立后, 服务器主动发送handshake
	Heartbeat = 2    // client定期发送心跳
	Kick      = 3    // server踢人
	Userdata  = 1000 // 用户自定义的类型, 从这里开始
)
