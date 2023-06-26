package road

import (
	"github.com/lixianmin/gonsole/road/serde"
	"net"
)

/********************************************************************
created:    2022-04-07
author:     lixianmin

Copyright (C) - All Rights Reserved
*********************************************************************/

type Session interface {
	Handshake() error                   // server主动向client发送服务器的配置信息
	HandshakeRe(serdeName string) error // client回复server的handshake协议
	Kick() error                        // server主动踢client
	PushByRoute(route string, v interface{}) error
	PushByKind(kind int32, v interface{}) error
	Close() error

	OnReceivingPacket(handler func(pack serde.Packet) error)
	OnClosed(handler func())

	Id() int64
	RemoteAddr() net.Addr
	Attachment() Attachment
}
