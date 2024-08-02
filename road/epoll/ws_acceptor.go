package epoll

import (
	"github.com/gobwas/ws"
	"github.com/lixianmin/gonsole/road/intern"
	"github.com/lixianmin/logo"
	"net/http"
	"sync"
)

/********************************************************************
created:    2020-12-07
author:     lixianmin

Copyright (C) - All Rights Reserved
*********************************************************************/

type WsAcceptor struct {
	serveMux   IServeMux
	servePath  string
	linkChan   chan intern.Link
	isClosed   int32
	listenOnce sync.Once
}

func NewWsAcceptor(serveMux IServeMux, servePath string, opts ...AcceptorOption) *WsAcceptor {
	if serveMux == nil {
		logo.Error("serveMux is nil")
		return nil
	}

	var options = newAcceptorOptions()
	for _, opt := range opts {
		opt(&options)
	}

	var my = &WsAcceptor{
		serveMux:  serveMux,
		servePath: servePath,
		linkChan:  make(chan intern.Link, options.LinkChanSize),
	}

	return my
}

func (my *WsAcceptor) serveHTTP(w http.ResponseWriter, r *http.Request) {
	// Upgrade connection
	conn, _, _, err := ws.UpgradeHTTP(r, w)
	if err != nil {
		logo.JsonI("err", err)
		return
	}

	my.linkChan <- intern.NewWsLink(conn)
}

func (my *WsAcceptor) Listen() {
	my.listenOnce.Do(func() {
		// 这个相当于listener，每次创建一个新的链接
		my.serveMux.HandleFunc(my.servePath, my.serveHTTP)
	})
}

func (my *WsAcceptor) GetLinkChan() chan intern.Link {
	return my.linkChan
}
