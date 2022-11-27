package epoll

import (
	"github.com/lixianmin/got/iox"
	"github.com/lixianmin/got/loom"
	"net"
	"sync"
	"time"
)

/********************************************************************
created:    2020-12-06
author:     lixianmin

Copyright (C) - All Rights Reserved
*********************************************************************/

type TcpConn struct {
	conn              net.Conn
	heartbeatInterval time.Duration
	onReadHandler     OnReadHandler
	wc                loom.WaitClose
	writeLock         sync.Mutex
}

func newTcpConn(conn net.Conn, heartbeatInterval time.Duration) *TcpConn {
	var my = &TcpConn{
		conn:              conn,
		heartbeatInterval: heartbeatInterval,
		onReadHandler:     emptyOnReadHandler,
	}

	go my.goLoop()
	return my
}

func (my *TcpConn) goLoop() {
	defer loom.DumpIfPanic()
	defer my.Close()

	var buffer = make([]byte, 1024)
	var input = &iox.Buffer{}

	for !my.wc.IsClosed() {
		var num, err = my.conn.Read(buffer)
		if err != nil {
			my.onReadHandler(nil, err)
			//logo.JsonI("err", err)
			return
		}

		_ = my.conn.SetReadDeadline(time.Now().Add(my.heartbeatInterval * 3))
		_, _ = input.Write(buffer[:num])
		if err2 := onReceiveMessage(input, my.onReadHandler); err2 != nil {
			//logo.JsonI("err2", err2)
			return
		}
	}
}

func (my *TcpConn) SetOnReadHandler(handler OnReadHandler) {
	if handler != nil {
		my.onReadHandler = handler
	}
}

func (my *TcpConn) Write(data []byte) (int, error) {
	// 同一个conn在不同的协程中异步write可能导致panic，原先采用N协程处理M个链接（N<M)的方案，现在改为lock处理并发问题
	my.writeLock.Lock()
	var num, err = my.conn.Write(data)
	my.writeLock.Unlock()

	return num, err
}

// Close closes the connection.
// Any blocked Read or Write operations will be unblocked and return errors.
func (my *TcpConn) Close() error {
	return my.wc.Close(func() error {
		return my.conn.Close()
	})
}

// RemoteAddr returns the remote address.
func (my *TcpConn) RemoteAddr() net.Addr {
	return my.conn.RemoteAddr()
}
