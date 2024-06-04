package intern

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

type TcpLink struct {
	commonLink
}

func NewTcpLink(conn net.Conn) *TcpLink {
	var my = &TcpLink{
		commonLink: commonLink{
			conn: conn,
		},
	}

	return my
}

func (my *TcpLink) GoLoop(kickInterval time.Duration, onReadHandler OnReadHandler) {
	defer loom.DumpIfPanic()
	defer func() {
		_ = my.conn.Close()
		_ = my.Close()
	}()

	var buffer = make([]byte, 1024)
	var stream = &iox.OctetsStream{}
	var reader = iox.NewOctetsReader(stream)

	for atomic.LoadInt32(&my.isClosed) == 0 {
		my.resetReadDeadline(kickInterval)
		var num, err1 = my.conn.Read(buffer)
		if err1 != nil {
			onReadHandler(nil, err1)
			return
		}

		_ = stream.Write(buffer[:num])
		onReadHandler(reader, nil)
		stream.Tidy()
	}
}

func (my *TcpLink) Write(data []byte) (int, error) {
	// net.TCPConn本身的是thread safe的，只要每次写入的都是完整的message，就不需要并发控制
	var num, err = my.conn.Write(data)
	return num, err
}
