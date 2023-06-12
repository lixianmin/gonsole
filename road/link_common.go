package road

import (
	"net"
	"sync/atomic"
	"time"
)

/********************************************************************
created:    2022-11-27
author:     lixianmin

Copyright (C) - All Rights Reserved
*********************************************************************/

type commonLink struct {
	conn     net.Conn
	isClosed int32
}

func (my *commonLink) resetReadDeadline(heartbeatInterval time.Duration) {
	// 不知道为什么经常报:  close session(27) by err="read tcp 192.168.31.96:8888->192.168.31.96:59562: i/o timeout"
	// 1. 改为SetDeadline(),即同时改写read/write的deadline也是无效的
	// 2. 时间已经设置为3倍的heartbeat时间了
	//
	// i/o timeout有可能是chrome的throttling mechanism导致的, 当chrome tab在background的时候, setInterval()的调用间隔可能会放大到1min, 就很容易超时了
	_ = my.conn.SetReadDeadline(time.Now().Add(heartbeatInterval * 3))
}

// Close closes the connection.
// Any blocked Read or Write operations will be unblocked and return errors.
func (my *commonLink) Close() error {
	atomic.StoreInt32(&my.isClosed, 1)
	return nil
}

// RemoteAddr returns the remote address.
func (my *commonLink) RemoteAddr() net.Addr {
	return my.conn.RemoteAddr()
}
