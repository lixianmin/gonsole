package epoll

import (
	"github.com/lixianmin/got/iox"
	"github.com/lixianmin/got/loom"
	"github.com/lixianmin/logo"
	"net"
	"sync"
)

/********************************************************************
created:    2020-12-06
author:     lixianmin

Copyright (C) - All Rights Reserved
*********************************************************************/

type TcpConn struct {
	conn         net.Conn
	receivedChan chan Message
	wc           loom.WaitClose
	writeLock    sync.Mutex
}

func newTcpConn(conn net.Conn, receivedChanSize int) *TcpConn {
	var receivedChan = make(chan Message, receivedChanSize)
	var my = &TcpConn{
		conn:         conn,
		receivedChan: receivedChan,
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
			my.receivedChan <- Message{Err: err}
			logo.JsonI("err", err)
			return
		}

		_, _ = input.Write(buffer[:num])
		if err2 := onReceiveMessage(my.receivedChan, input); err2 != nil {
			my.receivedChan <- Message{Err: err2}
			logo.JsonI("err2", err2)
			return
		}
	}
}

func (my *TcpConn) GetReceivedChan() <-chan Message {
	return my.receivedChan
}

func (my *TcpConn) Write(b []byte) (int, error) {
	// 同一个conn在不同的协程中异步write可能导致panic，原先采用N协程处理M个链接（N<M)的方案，现在改为lock处理并发问题
	my.writeLock.Lock()
	var num, err = my.conn.Write(b)
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
