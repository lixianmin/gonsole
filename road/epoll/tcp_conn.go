package epoll

import (
	"github.com/lixianmin/gonsole/road/codec"
	"github.com/lixianmin/got/iox"
	"github.com/lixianmin/got/loom"
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
	input        *iox.Buffer
	wc           loom.WaitClose
}

func newTcpConn(conn net.Conn, receivedChanSize int) *TcpConn {
	var receivedChan = make(chan Message, receivedChanSize)
	var my = &TcpConn{
		conn:         conn,
		receivedChan: receivedChan,
		input:        &iox.Buffer{},
	}

	return my
}

func (my *TcpConn) sendErrorMessage(err error) {
	my.writeMessage(Message{Err: err})
}

func (my *TcpConn) GetReceivedChan() <-chan Message {
	return my.receivedChan
}

func (my *TcpConn) onReceiveData(buff []byte) error {
	var input = my.input
	var _, err = input.Write(buff)
	if err != nil {
		return err
	}

	var headLength = codec.HeaderLength
	var data = input.Bytes()

	for len(data) > headLength {
		var header = data[:headLength]
		msgSize, _, err := codec.ParseHeader(header)
		if err != nil {
			return err
		}

		var totalSize = headLength + msgSize
		if len(data) < totalSize {
			return nil
		}

		// 这里每次新建的frameData目前是省不下的, 原因是writeMessage()方法会把这个slice写到chan中并由另一个goroutine使用
		var frameData = make([]byte, totalSize)
		copy(frameData, data[:totalSize])

		my.writeMessage(Message{Data: frameData})
		input.Next(totalSize)
		data = input.Bytes()
	}

	my.input.Tidy()
	return nil
}

// Write writes data to the connection.
// Write can be made to time out and return an Error with Timeout() == true
// after a fixed time limit; see SetDeadline and SetWriteDeadline.
func (my *TcpConn) Write(b []byte) (int, error) {
	return my.conn.Write(b)
}

func (my *TcpConn) writeMessage(msg Message) {
	select {
	case my.receivedChan <- msg:
	case <-my.wc.C():
	}
}

// Close closes the connection.
// Any blocked Read or Write operations will be unblocked and return errors.
func (my *TcpConn) Close() error {
	return my.wc.Close(func() error {
		return my.conn.Close()
	})
}

// LocalAddr returns the local address.
func (my *TcpConn) LocalAddr() net.Addr {
	return my.conn.LocalAddr()
}

// RemoteAddr returns the remote address.
func (my *TcpConn) RemoteAddr() net.Addr {
	return my.conn.RemoteAddr()
}
