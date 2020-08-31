package network

import (
	"context"
	"encoding/json"
	"github.com/lixianmin/gonsole/logger"
	"github.com/lixianmin/gonsole/network/acceptor"
	"github.com/lixianmin/gonsole/network/component"
	"github.com/lixianmin/gonsole/network/conn/codec"
	"github.com/lixianmin/gonsole/network/conn/message"
	"github.com/lixianmin/gonsole/network/conn/packet"
	"github.com/lixianmin/gonsole/network/route"
	"github.com/lixianmin/gonsole/network/serialize"
	"github.com/lixianmin/gonsole/network/service"
	"github.com/lixianmin/gonsole/network/util"
	"github.com/lixianmin/gonsole/network/util/compression"
	"github.com/lixianmin/got/loom"
	"reflect"
	"sync"
	"time"
)

/********************************************************************
created:    2020-08-28
author:     lixianmin

Copyright (C) - All Rights Reserved
*********************************************************************/

var (
	// hbd contains the heartbeat packet data
	hbd []byte
	// hrd contains the handshake response data
	hrd  []byte
	once sync.Once
)

type (
	Agent struct {
		conn           acceptor.PlayerConn
		packetEncoder  codec.PacketEncoder
		packetDecoder  codec.PacketDecoder
		messageEncoder message.Encoder
		serializer     serialize.Serializer
		sendingChan    chan sendingItem
		receivedChan   chan receivedItem
		wc             loom.WaitClose
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

func NewAgent(conn acceptor.PlayerConn,
	packetEncoder codec.PacketEncoder,
	packetDecoder codec.PacketDecoder,
	messageEncoder message.Encoder,
	serializer serialize.Serializer) *Agent {

	const bufferSize = 16
	var agent = &Agent{
		conn:           conn,
		packetEncoder:  packetEncoder,
		packetDecoder:  packetDecoder,
		messageEncoder: messageEncoder,
		serializer:     serializer,
		sendingChan:    make(chan sendingItem, bufferSize),
		receivedChan:   make(chan receivedItem, bufferSize),
	}

	once.Do(func() {
		hbdEncode(2*time.Second, packetEncoder, messageEncoder.IsCompressionEnabled(), serializer.GetName())
	})

	loom.Go(agent.goReceive)
	loom.Go(agent.goSend)
	loom.Go(agent.goProcess)
	return agent
}

func (my *Agent) goProcess(later *loom.Later) {
	for {
		select {
		case data := <-my.receivedChan:
			my.processReceived(data)
		case <-my.wc.C():
			return
		}
	}
}

func (my *Agent) processReceived(data receivedItem) {
	ret, err := my.processReceivedImpl(data)
	if data.msg.Type != message.Notify {
		if err != nil {
			logger.Info("Failed to process handler message: %s", err.Error())
		} else {
			err := my.ResponseMID(data.ctx, data.msg.ID, ret)
			if err != nil {

			}
		}
	}
}

func (my *Agent) processReceivedImpl(data receivedItem) ([]byte, error) {
	handler, err := service.GetHandler(data.route)
	if err != nil {
		return nil, err
	}

	// First unmarshal the handler arg that will be passed to
	// both handler and pipeline functions
	arg, err := unmarshalHandlerArg(handler, my.serializer, data.msg.Data)
	if err != nil {
		return nil, err
	}

	args := []reflect.Value{handler.Receiver, reflect.ValueOf(data.ctx)}
	if arg != nil {
		args = append(args, reflect.ValueOf(arg))
	}

	resp, err := util.Pcall(handler.Method, args)

	ret, err := serializeReturn(my.serializer, resp)
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

func (my *Agent) Close() {
	my.wc.Close(func() {
		_ = my.conn.Close()
	})
}

func hbdEncode(heartbeatTimeout time.Duration, packetEncoder codec.PacketEncoder, dataCompression bool, serializerName string) {
	hData := map[string]interface{}{
		"code": 200,
		"sys": map[string]interface{}{
			"heartbeat":  heartbeatTimeout.Seconds(),
			"dict":       message.GetDictionary(),
			"serializer": serializerName,
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

	hrd, err = packetEncoder.Encode(packet.Handshake, data)
	if err != nil {
		panic(err)
	}

	hbd, err = packetEncoder.Encode(packet.Heartbeat, nil)
	if err != nil {
		panic(err)
	}
}
