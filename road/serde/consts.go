package serde

/********************************************************************
created:    2023-06-05
author:     lixianmin

Copyright (C) - All Rights Reserved
*********************************************************************/

const (
	Handshake   = 1    // 连接建立后, 服务器主动发送handshake
	Heartbeat   = 2    // 定期发送心跳, 主要是为了保活
	Kick        = 3    // kick, 直接踢人
	UserDefined = 1000 // 用户自定义的类型, 从这里开始
)
