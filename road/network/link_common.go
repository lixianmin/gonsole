package network

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
