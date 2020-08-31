package network

import (
	"github.com/lixianmin/gonsole/logger"
	"github.com/lixianmin/gonsole/network/acceptor"
	"github.com/lixianmin/gonsole/network/component"
	"github.com/lixianmin/gonsole/network/conn/codec"
	"github.com/lixianmin/gonsole/network/conn/message"
	"github.com/lixianmin/gonsole/network/serialize"
	"github.com/lixianmin/gonsole/network/service"
	"github.com/lixianmin/got/loom"
)

/********************************************************************
created:    2020-08-28
author:     lixianmin

Copyright (C) - All Rights Reserved
*********************************************************************/

type (
	AppArgs struct {
		ListenAddress   string
		DataCompression bool
	}

	App struct {
		acceptor       acceptor.Acceptor
		packetEncoder  codec.PacketEncoder
		packetDecoder  codec.PacketDecoder
		messageEncoder message.Encoder
		serializer     serialize.Serializer

		handlerService *service.HandlerService
		handlerComp    []regComp
	}

	regComp struct {
		comp component.Component
		opts []component.Option
	}
)

func NewApp(args AppArgs) *App {
	checkAppArgs(&args)
	var app = &App{
		acceptor:       acceptor.NewWSAcceptor(args.ListenAddress),
		packetDecoder:  codec.NewPomeloPacketDecoder(),
		packetEncoder:  codec.NewPomeloPacketEncoder(),
		messageEncoder: message.NewMessagesEncoder(args.DataCompression),
		serializer:     serialize.NewJsonSerializer(),

		handlerService: service.NewHandlerService(),
		handlerComp:    make([]regComp, 0, 4),
	}

	loom.Go(app.goLoop)
	return app
}

func checkAppArgs(args *AppArgs) {

}

func (my *App) goLoop(later *loom.Later) {
	for {
		select {
		case conn := <-my.acceptor.GetConnChan():
			NewAgent(conn, my.packetEncoder, my.packetDecoder, my.messageEncoder, my.serializer)
		}
	}
}

func (my *App) Start() {
	// register all components
	for _, c := range my.handlerComp {
		if err := my.handlerService.Register(c.comp, c.opts); err != nil {
			logger.Warn("Failed to register handler: %s", err.Error())
		}
	}

	my.acceptor.ListenAndServe()
}

func (my *App) Register(c component.Component, options ...component.Option) {
	my.handlerComp = append(my.handlerComp, regComp{c, options})
}
