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
	Handshake() error // server主动向client发送服务器的配置信息
	Kick() error      // server主动踢client
	SendByRoute(route string, v interface{}) error
	SendByKind(kind int32, v interface{}) error

	OnHandShaken(handler func()) // 握手完成后
	OnClosed(handler func())     // 连接关闭后

	Id() int64
	RemoteAddr() net.Addr
	Attachment() Attachment
	Nonce() int32
}

type ClientSession interface {
	Session
	HandshakeRe(serdeName string, handshake serde.JsonHandshake) error // client回复server的handshake协议
	Close() error

	OnReceivedPacket(handler func(pack serde.Packet) error) // 通过该回调用, client使用session对象时自定义自己的处理方法
}
