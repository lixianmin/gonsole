package epoll

import (
	"github.com/gobwas/ws"
	"github.com/gobwas/ws/wsutil"
	"github.com/lixianmin/got/iox"
	"github.com/lixianmin/got/loom"
	"net"
	"sync"
	"time"
)

/********************************************************************
created:    2020-12-07
author:     lixianmin

Copyright (C) - All Rights Reserved
*********************************************************************/

type WsConn struct {
	conn              net.Conn
	heartbeatInterval time.Duration
	onReadHandler     OnReadHandler
	writeLock         sync.Mutex
	wc                loom.WaitClose
}

func newWsConn(conn net.Conn, heartbeatInterval time.Duration) *WsConn {
	var my = &WsConn{
		conn:              conn,
		heartbeatInterval: heartbeatInterval,
		onReadHandler:     emptyOnReadHandler,
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
			//if err == io.EOF || err == io.ErrUnexpectedEOF {
			//	continue
			//}

			my.onReadHandler(nil, err)
			//logo.JsonI("err", err)
			return
		}

		_ = my.conn.SetReadDeadline(time.Now().Add(my.heartbeatInterval * 3))
		_, _ = input.Write(data)
		if err2 := onReceiveMessage(input, my.onReadHandler); err2 != nil {
			//logo.JsonI("err2", err2)
			return
		}
	}
}

func (my *WsConn) SetOnReadHandler(handler OnReadHandler) {
	if handler != nil {
		my.onReadHandler = handler
	}
}

func (my *WsConn) Write(data []byte) (int, error) {
	// 同一个conn在不同的协程中异步write可能导致panic，原先采用N协程处理M个链接（N<M)的方案，现在改为lock处理并发问题
	my.writeLock.Lock()
	var frame = ws.NewBinaryFrame(data)
	var err = ws.WriteFrame(my.conn, frame)
	my.writeLock.Unlock()

	//logo.JsonI("b", b)
	if err != nil {
		return 0, err
	}

	return len(data), nil
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
