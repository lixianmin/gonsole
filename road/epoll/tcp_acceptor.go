package epoll

import (
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
	connChan chan IConn
	isClosed int32
}

func NewTcpAcceptor(address string, opts ...AcceptorOption) *TcpAcceptor {
	var options = acceptorOptions{
		ConnChanSize:   16,
		PollBufferSize: 1024,
	}

	for _, opt := range opts {
		opt(&options)
	}

	var my = &TcpAcceptor{
		connChan: make(chan IConn, options.ConnChanSize),
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

		my.connChan <- newTcpConn(conn)
	}
}

func (my *TcpAcceptor) Close() error {
	atomic.StoreInt32(&my.isClosed, 1)
	return nil
}

func (my *TcpAcceptor) GetConnChan() chan IConn {
	return my.connChan
}
