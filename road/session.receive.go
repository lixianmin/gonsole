package road

import (
	"context"
	"fmt"
	"github.com/lixianmin/gonsole/ifs"
	"github.com/lixianmin/gonsole/road/codec"
	"github.com/lixianmin/gonsole/road/component"
	"github.com/lixianmin/gonsole/road/message"
	"github.com/lixianmin/gonsole/road/route"
	"github.com/lixianmin/gonsole/road/serialize"
	"github.com/lixianmin/gonsole/road/util"
	"github.com/lixianmin/got/iox"
	"github.com/lixianmin/logo"
	"reflect"
)

/********************************************************************
created:    2020-08-31
author:     lixianmin

Copyright (C) - All Rights Reserved
*********************************************************************/

func (my *sessionImpl) initialize() {
	var msgBuffer = &iox.Buffer{}
	my.conn.SetOnReadHandler(func(data []byte, err error) {
		if err != nil {
			logo.Info("close session(%d) by err=%q", my.id, err)
			_ = my.Close()
			return
		}

		_, _ = msgBuffer.Write(data)
		if err1 := my.onReceivedMessage(msgBuffer); err1 != nil {
			logo.Info("close session(%d) by onReceivedMessage(), err=%q", my.id, err1)
			_ = my.Close()
			return
		}
	})
}

func (my *sessionImpl) onReceivedMessage(buffer *iox.Buffer) error {
	packets, err := my.app.packetDecoder.Decode(buffer)
	if err != nil {
		var err1 = fmt.Errorf("failed to decode message: %s", err.Error())
		return err1
	}

	if !my.rateLimiter.Allow() {
		return ErrKickedByRateLimit
	}

	// process all packet
	for i := range packets {
		var p = packets[i]
		//logo.JsonI("p", p)
		switch p.Kind {
		case codec.Handshake:
			if err := my.onReceivedHandshake(p); err != nil {
				return err
			}
		case codec.HandshakeAck, codec.Heartbeat:
			// 1. HandshakeAck：回复heartbeat是为了激活js的setTimeout()定时发送heartbeat的功能，在此之前是不应该定时发送heartbeat的
			// 2. Heartbeat: 回复heartbeat是因为现在server只有一个goroutine，被用在了阻塞式读取网络数据，因此server缺少定时发送heartbeat的能力，转而采用client主动heartbeat而server回复的方案
			if err := my.onReceivedHeartbeat(); err != nil {
				return err
			}
		case codec.Data:
			if err := my.onReceivedData(p); err != nil {
				return err
			}
		}
	}

	return nil
}

func (my *sessionImpl) onReceivedHandshake(p *codec.Packet) error {
	var err = my.writeBytes(my.app.handshakeData)
	if err == nil {
		my.onHandShaken.Invoke()
	}

	return err
}

func (my *sessionImpl) onReceivedHeartbeat() error {
	// 发送心跳包，如果网络是通的，收到心跳返回时会刷新 lastAt
	if err := my.writeBytes(my.app.heartbeatPacketData); err != nil {
		return fmt.Errorf("failed to write to conn: %s", err.Error())
	}

	// 注意：libpitaya的heartbeat部分是问题的，只能在应用层自己做ping/pong
	//logo.Debug("session(%d) sent heartbeat", my.id)
	return nil
}

func (my *sessionImpl) onReceivedData(p *codec.Packet) error {
	var item, err = my.decodeReceivedData(p)
	if err != nil {
		var err1 = fmt.Errorf("failed to process packet: %s", err.Error())
		return err1
	}

	// 取handler，准备处理协议
	var handler, err2 = my.app.getHandler(item.route)
	if err2 != nil {
		return err2
	}

	// 这个err3不能立即返回，需要变成后面的data中的err并输出到client
	// 注意err4与err3并没有什么关系，err3是业务逻辑错误，并不引起session.Close()
	var payload, err3 = processReceivedData(item, handler, my.app.serializer)
	needReply := item.msg.Type != message.Notify
	if needReply {
		var msg = message.Message{Type: message.Response, Id: item.msg.Id, Data: payload}
		var data, err4 = my.encodeMessageMayError(msg, err3)
		if err4 != nil {
			return err4
		}

		return my.writeBytes(data)
	}

	return nil
}

func (my *sessionImpl) decodeReceivedData(p *codec.Packet) (receivedItem, error) {
	msg, err := message.Decode(p.Data)
	if err != nil {
		return receivedItem{}, err
	}

	r, err := route.Decode(msg.Route)
	if err != nil {
		return receivedItem{}, err
	}

	var ctx = context.WithValue(context.Background(), ifs.CtxKeySession, my)

	var item = receivedItem{
		ctx:   ctx,
		route: r,
		msg:   msg,
	}

	return item, nil
}

func processReceivedData(data receivedItem, handler *component.Handler, serializer serialize.Serializer) ([]byte, error) {
	// First unmarshal the handler argument that will be passed to
	// both handler and pipeline functions
	var arg, err = unmarshalHandlerArg(handler, serializer, data.msg.Data)
	if err != nil {
		return nil, err
	}

	var args []reflect.Value
	if arg != nil {
		args = []reflect.Value{handler.Receiver, reflect.ValueOf(data.ctx), reflect.ValueOf(arg)}
		// 如果request实现了IRequestPart接口，则处理一下
		if part, ok := arg.(IRequestPart); ok {
			if err2 := part.OnAdded(data.ctx, arg); err2 != nil {
				return nil, err2
			}
		}
	} else {
		args = []reflect.Value{handler.Receiver, reflect.ValueOf(data.ctx)}
	}

	response, err3 := util.PCall(handler.Method, args)
	if err3 != nil {
		return nil, err3
	}

	ret, err4 := util.SerializeOrRaw(serializer, response)
	if err4 != nil {
		return nil, err4
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
