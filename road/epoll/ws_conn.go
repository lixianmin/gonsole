package epoll

import (
	"github.com/gobwas/ws"
	"github.com/gobwas/ws/wsutil"
	"github.com/lixianmin/got/iox"
	"github.com/lixianmin/got/loom"
	"github.com/lixianmin/logo"
	"net"
)

/********************************************************************
created:    2020-12-07
author:     lixianmin

Copyright (C) - All Rights Reserved
*********************************************************************/

type WsConn struct {
	conn         net.Conn
	receivedChan chan Message
	wc           loom.WaitClose
}

func newWsConn(conn net.Conn, receivedChanSize int) *WsConn {
	var receivedChan = make(chan Message, receivedChanSize)
	var my = &WsConn{
		conn:         conn,
		receivedChan: receivedChan,
	}

	go my.goLoop()
	return my
}

func (my *WsConn) goLoop() {
	defer loom.DumpIfPanic()
	defer my.Close()

	var input = &iox.Buffer{}
	for !my.wc.IsClosed() {
		data, _, err := wsutil.ReadData(my.conn, ws.StateServerSide)
		if err != nil {
			my.receivedChan <- Message{Err: err}
			logo.JsonI("err", err)
			return
		}

		_, _ = input.Write(data)
		if err2 := onReceiveMessage(my.receivedChan, input); err2 != nil {
			my.receivedChan <- Message{Err: err2}
			logo.JsonI("err2", err2)
			return
		}
	}
}

func (my *WsConn) GetReceivedChan() <-chan Message {
	return my.receivedChan
}

// Write writes data to the connection.
// Write can be made to time out and return an Error with Timeout() == true
// after a fixed time limit; see SetDeadline and SetWriteDeadline.
func (my *WsConn) Write(b []byte) (int, error) {
	var frame = ws.NewBinaryFrame(b)
	var err = ws.WriteFrame(my.conn, frame)
	if err != nil {
		return 0, err
	}

	return len(b), nil
}

// Close closes the connection.
// Any blocked Read or Write operations will be unblocked and return errors.
func (my *WsConn) Close() error {
	return my.wc.Close(func() error {
		return my.conn.Close()
	})
}

// RemoteAddr returns the remote address.
func (my *WsConn) RemoteAddr() net.Addr {
	return my.conn.RemoteAddr()
}
