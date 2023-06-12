package epoll

import (
	"github.com/gobwas/ws"
	"github.com/lixianmin/gonsole/road"
	"github.com/lixianmin/gonsole/road/internal"
	"github.com/lixianmin/logo"
	"net/http"
)

/********************************************************************
created:    2020-12-07
author:     lixianmin

Copyright (C) - All Rights Reserved
*********************************************************************/

type WsAcceptor struct {
	linkChan chan road.Link
	isClosed int32
}

func NewWsAcceptor(serveMux IServeMux, servePath string, opts ...AcceptorOption) *WsAcceptor {
	var options = newAcceptorOptions()
	for _, opt := range opts {
		opt(&options)
	}

	var my = &WsAcceptor{
		linkChan: make(chan road.Link, options.LinkChanSize),
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

	my.linkChan <- internal.NewWsLink(conn)
}

func (my *WsAcceptor) GetLinkChan() chan road.Link {
	return my.linkChan
}
