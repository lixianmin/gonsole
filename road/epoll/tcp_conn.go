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
			onReadHandler:     emptyOnReadHandler,
		},
	}

	go my.goLoop()
	return my
}

func (my *TcpConn) goLoop() {
	defer loom.DumpIfPanic()
	defer func() {
		_ = my.conn.Close()
		_ = my.Close()
	}()

	var buffer = make([]byte, 1024)
	var input = &iox.Buffer{}

	for atomic.LoadInt32(&my.isClosed) == 0 {
		var num, err = my.conn.Read(buffer)
		if err != nil {
			my.onReadHandler(nil, err)
			//logo.JsonI("err", err)
			return
		}

		my.resetReadDeadline()
		_, _ = input.Write(buffer[:num])
		if err2 := my.onReceiveMessage(input); err2 != nil {
			//logo.JsonI("err2", err2)
			return
		}
	}
}

func (my *TcpConn) Write(data []byte) (int, error) {
	// net.TCPConn本身的是thread safe的，只要每次写入的都是完整的message，就不需要并发控制
	var num, err = my.conn.Write(data)
	return num, err
}
