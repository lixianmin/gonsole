package epoll

import (
	"github.com/lixianmin/gonsole/road/intern"
	"github.com/lixianmin/got/loom"
	"github.com/lixianmin/logo"
	"net"
	"sync/atomic"
)

/********************************************************************
created:    2020-12-06
author:     lixianmin

Copyright (C) - All Rights Reserved
*********************************************************************/

type TcpAcceptor struct {
	linkChan chan intern.Link
	isClosed int32
}

func NewTcpAcceptor(address string, opts ...AcceptorOption) *TcpAcceptor {
	var options = newAcceptorOptions()
	for _, opt := range opts {
		opt(&options)
	}

	var my = &TcpAcceptor{
		linkChan: make(chan intern.Link, options.LinkChanSize),
	}

	go my.goLoop(address)
	return my
}

func (my *TcpAcceptor) goLoop(address string) {
	defer loom.DumpIfPanic()

	listener, err := net.Listen("tcp", address)
	if err != nil {
		logo.Warn("failed to listen on address=%q, err=%q", address, err)
		return
	}
	defer listener.Close()

	// while this acceptor is not closed
	for atomic.LoadInt32(&my.isClosed) == 0 {
		conn, err := listener.Accept()
		if err != nil {
			logo.Info("failed to accept TCP connection: %q", err)
			continue
		}

		// tcp链接对no delay的默认值就是true, 因此不需要设置
		//tcpConn, ok := conn.(*net.TCPConn)
		//_ = tcpConn.SetNoDelay(true)

		my.linkChan <- intern.NewTcpLink(conn)
	}
}

func (my *TcpAcceptor) Close() error {
	atomic.StoreInt32(&my.isClosed, 1)
	return nil
}

func (my *TcpAcceptor) GetLinkChan() chan intern.Link {
	return my.linkChan
}
