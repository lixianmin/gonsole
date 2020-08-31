package network

import (
	"encoding/json"
	"github.com/lixianmin/gonsole/logger"
	"github.com/lixianmin/gonsole/network/component"
	"github.com/lixianmin/gonsole/network/conn/codec"
	"github.com/lixianmin/gonsole/network/conn/message"
	"github.com/lixianmin/gonsole/network/conn/packet"
	"github.com/lixianmin/gonsole/network/serialize"
	"github.com/lixianmin/gonsole/network/service"
	"github.com/lixianmin/gonsole/network/util/compression"
	"github.com/lixianmin/got/loom"
	"time"
)

/********************************************************************
created:    2020-08-28
author:     lixianmin

Copyright (C) - All Rights Reserved
*********************************************************************/

type (
	App struct {
		commonSessionArgs
		acceptor Acceptor

		handlerService *service.HandlerService
	}
)

func NewApp(args AppArgs) *App {
	checkAppArgs(&args)

	var common = commonSessionArgs{
		packetDecoder:    codec.NewPomeloPacketDecoder(),
		packetEncoder:    codec.NewPomeloPacketEncoder(),
		messageEncoder:   message.NewMessagesEncoder(args.DataCompression),
		serializer:       serialize.NewJsonSerializer(),
		heartbeatTimeout: args.HeartbeatTimeout,
	}

	var app = &App{
		commonSessionArgs: common,
		acceptor:          args.Acceptor,
		handlerService:    service.NewHandlerService(),
	}

	app.heartbeatDataEncode(args.DataCompression)
	loom.Go(app.goLoop)
	return app
}

func checkAppArgs(args *AppArgs) {
	if args.HeartbeatTimeout == 0 {
		args.HeartbeatTimeout = 10 * time.Second
	}
}

func (my *App) goLoop(later *loom.Later) {
	for {
		select {
		case conn := <-my.acceptor.GetConnChan():
			NewSession(conn, my.commonSessionArgs)
		}
	}
}

func (my *App) Register(c component.Component, options ...component.Option) {
	var err = my.handlerService.Register(c, options)
	if err != nil {
		logger.Warn("Failed to register handler: %s", err.Error())
	}
}

func (my *App) heartbeatDataEncode(dataCompression bool) {
	hData := map[string]interface{}{
		"code": 200,
		"sys": map[string]interface{}{
			"heartbeat":  my.heartbeatTimeout.Seconds(),
			"dict":       message.GetDictionary(),
			"serializer": my.serializer.GetName(),
		},
	}

	data, err := json.Marshal(hData)
	if err != nil {
		panic(err)
	}

	if dataCompression {
		compressedData, err := compression.DeflateData(data)
		if err != nil {
			panic(err)
		}

		if len(compressedData) < len(data) {
			data = compressedData
		}
	}

	my.handshakeResponseData, err = my.packetEncoder.Encode(packet.Handshake, data)
	if err != nil {
		panic(err)
	}

	my.heartbeatPacketData, err = my.packetEncoder.Encode(packet.Heartbeat, nil)
	if err != nil {
		panic(err)
	}
}
