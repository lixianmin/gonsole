package epoll

import (
	"github.com/lixianmin/gonsole/road/codec"
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
			logo.JsonI("err", err)
			return
		}

		_, _ = input.Write(buffer[:num])
		if err2 := my.onReceiveData(input); err2 != nil {
			logo.JsonI("err2", err2)
			return
		}
	}
}

func (my *TcpConn) GetReceivedChan() <-chan Message {
	return my.receivedChan
}

func (my *TcpConn) onReceiveData(input *iox.Buffer) error {
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

		select {
		case my.receivedChan <- Message{Data: frameData}:
		case <-my.wc.C():
			return nil
		}

		input.Next(totalSize)
		data = input.Bytes()
	}

	input.Tidy()
	return nil
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

// LocalAddr returns the local address.
func (my *TcpConn) LocalAddr() net.Addr {
	return my.conn.LocalAddr()
}

// RemoteAddr returns the remote address.
func (my *TcpConn) RemoteAddr() net.Addr {
	return my.conn.RemoteAddr()
}
