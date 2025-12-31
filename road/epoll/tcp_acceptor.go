package epoll

import (
	"net"
	"sync"
	"sync/atomic"
	"time"

	"github.com/lixianmin/gonsole/road/intern"
	"github.com/lixianmin/got/loom"
	"github.com/lixianmin/logo"
)

/********************************************************************
created:    2020-12-06
author:     lixianmin

Copyright (C) - All Rights Reserved
*********************************************************************/

type TcpAcceptor struct {
	address    string
	linkChan   chan intern.Link
	isClosed   int32
	listenOnce sync.Once
}

func NewTcpAcceptor(address string, opts ...AcceptorOption) *TcpAcceptor {
	var options = newAcceptorOptions()
	for _, opt := range opts {
		opt(&options)
	}

	var my = &TcpAcceptor{
		address:  address,
		linkChan: make(chan intern.Link, options.LinkChanSize),
	}

	return my
}

func (my *TcpAcceptor) goLoop() {
	defer loom.DumpIfPanic()

	var listener, err = net.Listen("tcp", my.address)
	if err != nil {
		logo.Warn("failed to listen on address=%q, err=%q", my.address, err)
		return
	}
	defer listener.Close()

	// while this acceptor is not closed
	for atomic.LoadInt32(&my.isClosed) == 0 {
		var conn, err2 = listener.Accept()
		if err2 != nil {
			logo.Info("failed to accept TCP connection: %q", err2)
			continue
		}

		if tcpConn, ok := conn.(*net.TCPConn); ok {
			// tcp链接对no delay的默认值就是true, 显示设置是为了让读代码的人更安心
			_ = tcpConn.SetNoDelay(true)
			_ = tcpConn.SetKeepAlive(true)
			_ = tcpConn.SetKeepAlivePeriod(time.Minute)
		}

		my.linkChan <- intern.NewTcpLink(conn)
	}
}

func (my *TcpAcceptor) Close() error {
	atomic.StoreInt32(&my.isClosed, 1)
	return nil
}

func (my *TcpAcceptor) Listen() {
	my.listenOnce.Do(func() {
		go my.goLoop()
	})
}

func (my *TcpAcceptor) GetLinkChan() chan intern.Link {
	return my.linkChan
}
