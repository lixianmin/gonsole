package network

import (
	"github.com/gobwas/ws"
	"github.com/gobwas/ws/wsutil"
	"github.com/lixianmin/got/iox"
	"github.com/lixianmin/got/loom"
	"net"
	"sync"
	"sync/atomic"
	"time"
)

/********************************************************************
created:    2020-12-07
author:     lixianmin

Copyright (C) - All Rights Reserved
*********************************************************************/

type WsConn struct {
	commonConn
	writeLock sync.Mutex
}

func NewWsConn(conn net.Conn) *WsConn {
	var my = &WsConn{
		commonConn: commonConn{
			conn: conn,
		},
	}

	return my
}

func (my *WsConn) GoLoop(heartbeatInterval time.Duration, onReadHandler OnReadHandler) {
	defer loom.DumpIfPanic()
	defer func() {
		_ = my.conn.Close()
		_ = my.Close()
	}()

	var stream = &iox.OctetsStream{}
	var reader = iox.NewOctetsReader(stream)

	for atomic.LoadInt32(&my.isClosed) == 0 {
		data, _, err := wsutil.ReadData(my.conn, ws.StateServerSide)
		if err != nil {
			//if err == io.EOF || err == io.ErrUnexpectedEOF {
			//	continue
			//}

			onReadHandler(nil, err)
			//logo.JsonI("err", err)
			return
		}

		my.resetReadDeadline(heartbeatInterval)
		_ = stream.Write(data)
		onReadHandler(reader, nil)
		stream.Tidy()
	}
}

func (my *WsConn) Write(data []byte) (int, error) {
	// 同一个conn在不同的协程中异步write可能导致panic，原先采用N协程处理M个链接（N<M)的方案，现在改为lock处理并发问题
	// 底层的net.TCPConn的Write()是thread safe的，但是因为写web socket数据的时候，是分多次调用的，所以必须使用lock控制并发
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
