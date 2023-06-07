package epoll

import (
	"github.com/lixianmin/got/iox"
	"github.com/lixianmin/got/loom"
	"net"
	"sync/atomic"
	"time"
)

/********************************************************************
created:    2020-12-06
author:     lixianmin

Copyright (C) - All Rights Reserved
*********************************************************************/

type TcpConn struct {
	commonConn
}

func newTcpConn(conn net.Conn, heartbeatInterval time.Duration) *TcpConn {
	var my = &TcpConn{
		commonConn: commonConn{
			conn:              conn,
			heartbeatInterval: heartbeatInterval,
		},
	}

	return my
}

func (my *TcpConn) GoLoop(onReadHandler OnReadHandler) {
	defer loom.DumpIfPanic()
	defer func() {
		_ = my.conn.Close()
		_ = my.Close()
	}()

	var buffer = make([]byte, 1024)
	var stream = &iox.OctetsStream{}
	var reader = iox.NewOctetsReader(stream)

	for atomic.LoadInt32(&my.isClosed) == 0 {
		var num, err1 = my.conn.Read(buffer)
		if err1 != nil {
			onReadHandler(nil, err1)
			return
		}

		my.resetReadDeadline()
		_ = stream.Write(buffer[:num])
		onReadHandler(reader, nil)
		stream.Tidy()
	}
}

func (my *TcpConn) Write(data []byte) (int, error) {
	// net.TCPConn本身的是thread safe的，只要每次写入的都是完整的message，就不需要并发控制
	var num, err = my.conn.Write(data)
	return num, err
}
