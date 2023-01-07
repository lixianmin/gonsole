package epoll

import (
	"github.com/lixianmin/gonsole/road/codec"
	"github.com/lixianmin/got/iox"
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

func (my *commonConn) onReceiveMessage(input *iox.Buffer, onReadHandler OnReadHandler) error {
	const headSize = codec.HeadSize
	var remains = input.Bytes()

	// 像heartbeat之类的协议，有可能只有head没有body，所以需要使用>=
	for len(remains) >= headSize {
		var header = remains[:headSize]
		bodySize, _, err := codec.ParseHead(header)
		if err != nil {
			onReadHandler(nil, err)
			return err
		}

		var totalSize = headSize + bodySize
		if len(remains) < totalSize {
			return nil
		}

		// 这里每次新建的frameData目前是省不下的, 原因是writeMessage()方法会把这个slice写到chan中并由另一个goroutine使用
		//var frameData = make([]byte, totalSize)
		//copy(frameData, data[:totalSize])
		// onReadHandler()会把data[]中的数据copy走，因此不再需要新生成一个frameData
		onReadHandler(remains[:totalSize], nil)

		input.Next(totalSize)
		remains = input.Bytes()
	}

	input.Tidy()
	return nil
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
