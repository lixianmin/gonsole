package epoll

import (
	"github.com/lixianmin/got/iox"
	"github.com/lixianmin/got/loom"
	"github.com/lixianmin/logo"
	"net"
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

// Write writes data to the connection.
// Write can be made to time out and return an Error with Timeout() == true
// after a fixed time limit; see SetDeadline and SetWriteDeadline.
func (my *TcpConn) Write(b []byte) (int, error) {
	return my.conn.Write(b)
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
