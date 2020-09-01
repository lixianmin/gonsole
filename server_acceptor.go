package gonsole

import (
	"github.com/gorilla/websocket"
	"github.com/lixianmin/gonsole/logger"
	"github.com/lixianmin/gonsole/network"
	"net/http"
)

/********************************************************************
created:    2020-08-31
author:     lixianmin

Copyright (C) - All Rights Reserved
*********************************************************************/

type ServerAcceptor struct {
	upgrader *websocket.Upgrader
	connChan chan network.PlayerConn
}

func newServerAcceptor(readBufferSize int, writeBufferSize int) *ServerAcceptor {
	var upgrader = &websocket.Upgrader{
		ReadBufferSize:  readBufferSize,
		WriteBufferSize: writeBufferSize,
		CheckOrigin: func(r *http.Request) bool {
			return true
		},
	}

	var acceptor = &ServerAcceptor{
		upgrader: upgrader,
		connChan: make(chan network.PlayerConn, 8),
	}

	return acceptor
}

func (my *ServerAcceptor) HandleWebsocket(mux IServeMux, handlePattern string) {
	mux.HandleFunc(handlePattern, func(writer http.ResponseWriter, request *http.Request) {
		conn, err := my.upgrader.Upgrade(writer, request, nil)
		if err != nil {
			logger.Error("[HandleWebsocket(%s)] connection upgrade failed, userAgent=%q, err=%q", request.RemoteAddr, request.UserAgent(), err)
			return
		}

		playerConn, err := network.NewWSConn(conn)
		if err != nil {
			logger.Error("[HandleWebsocket(%s)] failed to create new ws connection: %s", request.RemoteAddr, err.Error())
			return
		}

		my.connChan <- playerConn
	})
}

func (my *ServerAcceptor) GetConnChan() chan network.PlayerConn {
	return my.connChan
}
