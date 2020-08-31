package network

import (
	"context"
	"github.com/lixianmin/gonsole/logger"
	"github.com/lixianmin/gonsole/network/component"
	"github.com/lixianmin/gonsole/network/conn/codec"
	"github.com/lixianmin/gonsole/network/conn/message"
	"github.com/lixianmin/gonsole/network/route"
	"github.com/lixianmin/gonsole/network/serialize"
	"github.com/lixianmin/gonsole/network/service"
	"github.com/lixianmin/gonsole/network/util"
	"github.com/lixianmin/got/loom"
	"reflect"
	"sync/atomic"
	"time"
)

/********************************************************************
created:    2020-08-28
author:     lixianmin

Copyright (C) - All Rights Reserved
*********************************************************************/

type (
	commonSessionArgs struct {
		packetEncoder  codec.PacketEncoder
		packetDecoder  codec.PacketDecoder
		messageEncoder message.Encoder
		serializer     serialize.Serializer

		heartbeatTimeout      time.Duration
		heartbeatPacketData   []byte
		handshakeResponseData []byte
	}

	Session struct {
		commonSessionArgs
		conn         PlayerConn
		sendingChan  chan sendingItem
		receivedChan chan receivedItem
		lastAt       int64 // last heartbeat unix time stamp
		wc           loom.WaitClose
	}

	receivedItem struct {
		ctx   context.Context
		route *route.Route
		msg   *message.Message
	}

	sendingItem struct {
		ctx  context.Context
		data []byte
		err  error
	}

	// 未编码消息
	sendingInfo struct {
		ctx     context.Context
		typ     message.Type // message type
		route   string       // message route (push)
		mid     uint         // response message id (response)
		payload interface{}  // payload
		err     bool         // if its an error message
	}
)

func NewSession(conn PlayerConn, args commonSessionArgs) *Session {
	const bufferSize = 16
	var agent = &Session{
		commonSessionArgs: args,
		sendingChan:       make(chan sendingItem, bufferSize),
		receivedChan:      make(chan receivedItem, bufferSize),
		lastAt:            time.Now().Unix(),
	}

	loom.Go(agent.goReceive)
	loom.Go(agent.goSend)
	loom.Go(agent.goProcess)
	return agent
}

func (my *Session) goProcess(later *loom.Later) {
	for {
		select {
		case data := <-my.receivedChan:
			my.processReceived(data)
		case <-my.wc.C():
			return
		}
	}
}

func (my *Session) processReceived(data receivedItem) {
	ret, err := processReceivedImpl(data, my.serializer)
	if data.msg.Type != message.Notify {
		if err != nil {
			logger.Info("Failed to process handler message: %s", err.Error())
		} else {
			err := my.responseMID(data.ctx, data.msg.ID, ret)
			if err != nil {
				logger.Info(err)
			}
		}
	}
}

func processReceivedImpl(data receivedItem, serializer serialize.Serializer) ([]byte, error) {
	handler, err := service.GetHandler(data.route)
	if err != nil {
		return nil, err
	}

	// First unmarshal the handler arg that will be passed to
	// both handler and pipeline functions
	arg, err := unmarshalHandlerArg(handler, serializer, data.msg.Data)
	if err != nil {
		return nil, err
	}

	args := []reflect.Value{handler.Receiver, reflect.ValueOf(data.ctx)}
	if arg != nil {
		args = append(args, reflect.ValueOf(arg))
	}

	resp, err := util.Pcall(handler.Method, args)

	ret, err := serializeReturn(serializer, resp)
	if err != nil {
		return nil, err
	}

	return ret, nil
}

func unmarshalHandlerArg(handler *component.Handler, serializer serialize.Serializer, payload []byte) (interface{}, error) {
	if handler.IsRawArg {
		return payload, nil
	}

	var arg interface{}
	if handler.Type != nil {
		arg = reflect.New(handler.Type.Elem()).Interface()
		err := serializer.Unmarshal(payload, arg)
		if err != nil {
			return nil, err
		}
	}

	return arg, nil
}

func serializeReturn(serializer serialize.Serializer, v interface{}) ([]byte, error) {
	if data, ok := v.([]byte); ok {
		return data, nil
	}

	data, err := serializer.Marshal(v)
	if err != nil {
		return nil, err
	}

	return data, nil
}

func (my *Session) Close() {
	my.wc.Close(func() {
		_ = my.conn.Close()
	})
}

func (my *Session) refreshLastAt() {
	atomic.StoreInt64(&my.lastAt, time.Now().Unix())
}
