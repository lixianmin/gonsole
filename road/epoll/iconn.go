package epoll

import "net"

/********************************************************************
created:    2020-09-06
author:     lixianmin

Copyright (C) - All Rights Reserved
*********************************************************************/

type OnReadHandler func(data []byte, err error)

type IConn interface {
	GoLoop(onReadHandler OnReadHandler)
	Write(data []byte) (int, error)
	Close() error
	RemoteAddr() net.Addr
}
