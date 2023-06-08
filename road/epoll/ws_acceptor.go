package epoll

import (
	"github.com/gobwas/ws"
	"github.com/lixianmin/gonsole/road/network"
	"github.com/lixianmin/logo"
	"net/http"
	"time"
)

/********************************************************************
created:    2020-12-07
author:     lixianmin

Copyright (C) - All Rights Reserved
*********************************************************************/

type WsAcceptor struct {
	connChan          chan network.Connection
	heartbeatInterval time.Duration
	isClosed          int32
}

func NewWsAcceptor(serveMux IServeMux, servePath string, opts ...AcceptorOption) *WsAcceptor {
	var options = newAcceptorOptions()
	for _, opt := range opts {
		opt(&options)
	}

	var my = &WsAcceptor{
		connChan:          make(chan network.Connection, options.ConnChanSize),
		heartbeatInterval: options.HeartbeatInterval,
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

	my.connChan <- newWsConn(conn, my.heartbeatInterval)
}

func (my *WsAcceptor) GetConnChan() chan network.Connection {
	return my.connChan
}

func (my *WsAcceptor) GetHeartbeatInterval() time.Duration {
	return my.heartbeatInterval
}
