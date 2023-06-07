package epoll

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

type commonConn struct {
	conn              net.Conn
	heartbeatInterval time.Duration
	isClosed          int32
}

func (my *commonConn) resetReadDeadline() {
	_ = my.conn.SetReadDeadline(time.Now().Add(my.heartbeatInterval * 3))
}

// Close closes the connection.
// Any blocked Read or Write operations will be unblocked and return errors.
func (my *commonConn) Close() error {
	atomic.StoreInt32(&my.isClosed, 1)
	return nil
}

// RemoteAddr returns the remote address.
func (my *commonConn) RemoteAddr() net.Addr {
	return my.conn.RemoteAddr()
}
