package network

import (
	"context"
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
	"github.com/lixianmin/got/loom"
	"reflect"
)

/********************************************************************
created:    2020-08-28
author:     lixianmin

Copyright (C) - All Rights Reserved
*********************************************************************/

type Agent struct {
	conn        acceptor.PlayerConn
	decoder     codec.PacketDecoder
	serializer  serialize.Serializer
	messageChan chan unhandledMessage
	wc          *loom.WaitClose
}

func NewAgent(conn acceptor.PlayerConn, decoder codec.PacketDecoder, serializer serialize.Serializer) *Agent {
	var agent = &Agent{
		conn:        conn,
		decoder:     decoder,
		serializer:  serializer,
		messageChan: make(chan unhandledMessage, 8),
		wc:          loom.NewWaitClose(),
	}

	loom.Go(agent.goRead)
	loom.Go(agent.goWrite)
	loom.Go(agent.goDispatch)
	return agent
}

func (my *Agent) goRead(later *loom.Later) {
	for {
		msg, err := my.conn.GetNextMessage()

		if err != nil {
			logger.Info("Error reading next available message: %s", err.Error())
			return
		}

		packets, err := my.decoder.Decode(msg)
		if err != nil {
			logger.Info("Failed to decode message: %s", err.Error())
			return
		}

		if len(packets) < 1 {
			logger.Warn("Read no packets, data: %v", msg)
			continue
		}

		// process all packet
		for i := range packets {
			if err := my.processPacket(packets[i]); err != nil {
				logger.Info("Failed to process packet: %s", err.Error())
				return
			}
		}
	}
}

func (my *Agent) processPacket(p *packet.Packet) error {
	switch p.Type {
	case packet.Handshake:
		logger.Debug("Received handshake packet")
	case packet.HandshakeAck:
		logger.Debug("Receive handshake ACK")
	case packet.Data:
		msg, err := message.Decode(p.Data)
		if err != nil {
			return err
		}
		my.processRawMessage(msg)
	case packet.Heartbeat:
		// expected
	}

	return nil
}

func (my *Agent) processRawMessage(msg *message.Message) {
	r, err := route.Decode(msg.Route)
	if err != nil {
		logger.Warn("Failed to decode route: %s", err.Error())
		return
	}

	message1 := unhandledMessage{
		ctx:   context.Background(),
		route: r,
		msg:   msg,
	}

	my.messageChan <- message1
}

func (my *Agent) goWrite(later *loom.Later) {

}

func (my *Agent) goDispatch(later *loom.Later) {
	for {
		select {
		case msg := <-my.messageChan:
			my.dispatchMessage(msg.ctx, msg.route, msg.msg)
		}
	}
}

func (my *Agent) dispatchMessage(ctx context.Context, route *route.Route, msg *message.Message) ([]byte, error) {
	h, err := service.GetHandler(route)
	if err != nil {
		return nil, err
	}

	// First unmarshal the handler arg that will be passed to
	// both handler and pipeline functions
	arg, err := unmarshalHandlerArg(h, my.serializer, msg.Data)
	if err != nil {
		return nil, err
	}

	args := []reflect.Value{h.Receiver, reflect.ValueOf(ctx)}
	if arg != nil {
		args = append(args, reflect.ValueOf(arg))
	}

	resp, err := util.Pcall(h.Method, args)

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
