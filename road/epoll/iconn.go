package epoll

import "net"

/********************************************************************
created:    2020-09-06
author:     lixianmin

Copyright (C) - All Rights Reserved
*********************************************************************/

type OnReadHandler func(data []byte, err error)

var emptyOnReadHandler = func(data []byte, err error) {

}

type IConn interface {
	SetOnReadHandler(handler OnReadHandler)
	Write(data []byte) (int, error)
	Close() error
	RemoteAddr() net.Addr
}
