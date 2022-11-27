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
	// 同一个conn在不同的协程中异步write可能导致panic，原先采用N协程处理M个链接（N<M)的方案，现在改为lock处理并发问题
	my.writeLock.Lock()
	var num, err = my.conn.Write(data)
	my.writeLock.Unlock()

	return num, err
}
