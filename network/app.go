package network

import (
	"github.com/lixianmin/gonsole/network/acceptor"
	"github.com/lixianmin/gonsole/network/conn/codec"
	"github.com/lixianmin/got/loom"
)

/********************************************************************
created:    2020-08-28
author:     lixianmin

Copyright (C) - All Rights Reserved
*********************************************************************/

type AppArgs struct {
	listenAddress string
}

type App struct {
	acceptor      acceptor.Acceptor
	packetDecoder codec.PacketDecoder
	packetEncoder codec.PacketEncoder
}

func NewApp(args AppArgs) *App {
	checkAppArgs(&args)
	var app = &App{
		acceptor:      acceptor.NewWSAcceptor(args.listenAddress),
		packetDecoder: codec.NewPomeloPacketDecoder(),
		packetEncoder: codec.NewPomeloPacketEncoder(),
	}

	loom.Go(app.goLoop)
	loom.Go(app.goListenAndServe)
	return app
}

func checkAppArgs(args *AppArgs) {

}

func (my *App) goLoop(later *loom.Later) {
	for {
		select {
		case conn := <-my.acceptor.GetConnChan():
			NewAgent(conn, my.packetDecoder)
		}
	}
}

func (my *App) goListenAndServe(later *loom.Later) {
	my.acceptor.ListenAndServe()
}
