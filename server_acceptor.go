package gonsole

import (
	"github.com/gorilla/websocket"
	"github.com/lixianmin/gonsole/logger"
	"github.com/lixianmin/road/acceptor"
	"net/http"
)

/********************************************************************
created:    2020-08-31
author:     lixianmin

Copyright (C) - All Rights Reserved
*********************************************************************/

type serverAcceptor struct {
	upgrader *websocket.Upgrader
	connChan chan acceptor.PlayerConn
}

func newServerAcceptor(readBufferSize int, writeBufferSize int) *serverAcceptor {
	var upgrader = &websocket.Upgrader{
		ReadBufferSize:  readBufferSize,
		WriteBufferSize: writeBufferSize,
		CheckOrigin: func(r *http.Request) bool {
			return true
		},
	}

	var actor = &serverAcceptor{
		upgrader: upgrader,
		connChan: make(chan acceptor.PlayerConn, 8),
	}

	return actor
}

func (my *serverAcceptor) HandleWebsocket(mux IServeMux, handlePattern string) {
	mux.HandleFunc(handlePattern, func(writer http.ResponseWriter, request *http.Request) {
		conn, err := my.upgrader.Upgrade(writer, request, nil)
		if err != nil {
			logger.Error("[HandleWebsocket(%s)] connection upgrade failed, userAgent=%q, err=%q", request.RemoteAddr, request.UserAgent(), err)
			return
		}

		playerConn, err := acceptor.NewWSConn(conn)
		if err != nil {
			logger.Error("[HandleWebsocket(%s)] failed to create new ws connection: %s", request.RemoteAddr, err.Error())
			return
		}

		my.connChan <- playerConn
	})
}

func (my *serverAcceptor) GetConnChan() chan acceptor.PlayerConn {
	return my.connChan
}

func (my *serverAcceptor) ListenAndServe() {
	
}
