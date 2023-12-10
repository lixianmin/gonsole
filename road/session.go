package road

import (
	"net"
)

/********************************************************************
created:    2022-04-07
author:     lixianmin

Copyright (C) - All Rights Reserved
*********************************************************************/

type Session interface {
	Handshake() error // server主动向client发送服务器的配置信息
	Kick() error      // server主动踢client
	Send(route string, v any) error

	OnHandShaken(handler func()) // 握手完成后
	OnClosed(handler func())     // 连接关闭后

	Id() int64
	RemoteAddr() net.Addr
	Attachment() Attachment
	Nonce() int32
}
