package epoll

import (
	"github.com/gobwas/ws"
	"github.com/lixianmin/logo"
	"net/http"
)

/********************************************************************
created:    2020-12-07
author:     lixianmin

Copyright (C) - All Rights Reserved
*********************************************************************/

type WsAcceptor struct {
	connChan         chan PlayerConn
	receivedChanSize int
	isClosed         int32
}

func NewWsAcceptor(serveMux IServeMux, servePath string, opts ...AcceptorOption) *WsAcceptor {
	var options = acceptorOptions{
		ConnChanSize:     16,
		ReceivedChanSize: 16,
		PollBufferSize:   1024,
	}

	for _, opt := range opts {
		opt(&options)
	}

	var my = &WsAcceptor{
		connChan:         make(chan PlayerConn, options.ConnChanSize),
		receivedChanSize: options.ReceivedChanSize,
	}

	// 这个相当于listener，每创建一个新的链接
	serveMux.HandleFunc(servePath, my.ServeHTTP)
	return my
}

func (my *WsAcceptor) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// Upgrade connection
	conn, _, _, err := ws.UpgradeHTTP(r, w)
	if err != nil {
		logo.JsonI("err", err)
		return
	}

	my.connChan <- newWsConn(conn, my.receivedChanSize)
}

func (my *WsAcceptor) GetConnChan() chan PlayerConn {
	return my.connChan
}
