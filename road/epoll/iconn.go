package epoll

import (
	"github.com/lixianmin/got/iox"
	"net"
)

/********************************************************************
created:    2020-09-06
author:     lixianmin

Copyright (C) - All Rights Reserved
*********************************************************************/

type OnReadHandler func(reader *iox.OctetsReader, err error)

type IConn interface {
	GoLoop(onReadHandler OnReadHandler)
	Write(data []byte) (int, error)
	Close() error
	RemoteAddr() net.Addr
}
