package road

import (
	"fmt"
	"github.com/lixianmin/gonsole/road/component"
	"github.com/lixianmin/gonsole/road/serde"
	"github.com/lixianmin/got/convert"
	"github.com/lixianmin/got/iox"
	"github.com/lixianmin/logo"
	"reflect"
)

/********************************************************************
created:    2020-08-31
author:     lixianmin

Copyright (C) - All Rights Reserved
*********************************************************************/

func (my *sessionImpl) startGoLoop() {
	go my.link.GoLoop(my.manager.kickInterval, func(reader *iox.OctetsReader, err error) {
		if err != nil {
			logo.Debug("close session(%d) by err=%q", my.id, err)
			_ = my.Close()
			return
		}

		if err1 := my.onReceivedData(reader); err1 != nil {
			logo.Info("close session(%d) by onReceivedData(), err=%q", my.id, err1)
			_ = my.Close()
			return
		}
	})
}

func (my *sessionImpl) onReceivedData(reader *iox.OctetsReader) error {
	var packets, err1 = serde.DecodePacket(reader)
	if err1 != nil {
		var err2 = fmt.Errorf("failed to decode message: %s", err1.Error())
		return err2
	}

	for _, pack := range packets {
		if err3 := my.onReceivedPacket(pack); err3 != nil {
			return err3
		}
	}

	return nil
}

func (my *sessionImpl) onReceivedPacket(pack serde.Packet) error {
	if pack.Kind >= serde.UserBase {
		if err1 := my.onReceivedUserdata(pack); err1 != nil {
			return err1
		}
	} else if pack.Kind == serde.Heartbeat {
		// 现在server只有一个goroutine用于阻塞式读取网络数据，因此server缺少定时发送heartbeat的能力，因此采用client主动heartbeat而server回复的方案
		if _, err2 := my.link.Write(my.manager.heartbeatBuffer); err2 != nil {
			return err2
		}
	} else if pack.Kind == serde.Echo {
		if err3 := my.onReceivedEcho(pack); err3 != nil {
			return err3
		}
	} else if pack.Kind == serde.HandshakeRe {
		if err4 := my.onReceivedHandshakeRe(pack); err4 != nil {
			return err4
		}
	}

	return nil
}

func (my *sessionImpl) onReceivedEcho(input serde.Packet) error {
	var requestId = input.RequestId
	var handler func() = nil

	my.handlerLock.Lock()
	{
		handler = my.echoHandlers[requestId]
		delete(my.echoHandlers, requestId)
	}
	my.handlerLock.Unlock()

	if handler != nil {
		defer func() {
			if rec := recover(); rec != nil {
				logo.JsonE("requestId", requestId, "recover", rec)
			}
		}()

		handler()
	} else {
		logo.JsonW("title", "echo handler is nil", "requestId", requestId)
	}

	return nil
}

func (my *sessionImpl) onReceivedHandshakeRe(input serde.Packet) error {
	var info serde.JsonHandshakeRe
	var err = convert.FromJsonE(input.Data, &info)
	if err != nil {
		return err
	}

	var s = my.manager.CreateSerde(info.Serde, my)
	if s == nil {
		return NewError("InvalidSerde", "info.Serde=%s", info.Serde)
	}

	my.setSerde(s)
	my.onEventHandShaken()
	return nil
}

func (my *sessionImpl) onEventHandShaken() {
	my.handlerLock.Lock()
	defer my.handlerLock.Unlock()

	{
		for _, handler := range my.onHandShakenHandlers {
			handler()
		}
		my.onHandShakenHandlers = nil
	}
}

func (my *sessionImpl) onReceivedUserdata(input serde.Packet) error {
	// client发来的消息, 必须有handler, 因此一定有kind才是合理的. server推送的消息可以没有kind
	var handler = my.manager.GetHandlerByKind(input.Kind)
	if handler == nil {
		return ErrEmptyHandler
	}

	if my.serde == nil {
		return ErrNilSerde
	}

	// 遍历拦截器
	for _, interceptor := range my.manager.interceptors {
		if err1 := interceptor(my, handler.Route); err1 != nil {
			return err1
		}
	}

	// 这个err不能立即返回，这是业务逻辑错误, 应该输出到client, 而不应该引发session.Close()
	var payload, err2 = processReceivedPacket(input, my.ctxValue, handler, my.serde)
	var output = serde.Packet{
		Kind:      input.Kind,
		RequestId: input.RequestId,
	}

	if err2 == nil {
		output.Data = payload
	} else if err3, ok := err2.(*Error); ok {
		output.Code = convert.Bytes(err3.Code)
		output.Data = convert.Bytes(err3.Message)
	} else {
		output.Code = convert.Bytes("PlainError")
		output.Data = convert.Bytes(err2.Error())
	}

	return my.sendPacket(output)
}

func processReceivedPacket(input serde.Packet, ctxValue reflect.Value, handler *component.Handler, serde serde.Serde) ([]byte, error) {
	// First unmarshal the handler argument that will be passed to both handler and pipeline functions
	var requestArg, err1 = unmarshalHandlerArg(handler, serde, input.Data)
	if err1 != nil {
		return nil, err1
	}

	var args []reflect.Value
	if requestArg != nil {
		args = []reflect.Value{handler.Receiver, ctxValue, reflect.ValueOf(requestArg)}
	} else {
		args = []reflect.Value{handler.Receiver, ctxValue}
	}

	var response, err2 = callMethod(handler.Method, args)
	if err2 != nil {
		return nil, err2
	}

	var ret, err3 = serializeOrRaw(serde, response)
	if err3 != nil {
		return nil, err3
	}

	return ret, nil
}

func unmarshalHandlerArg(handler *component.Handler, serde serde.Serde, payload []byte) (any, error) {
	if handler.IsRawArg {
		return payload, nil
	}

	var arg any
	if handler.Type != nil {
		arg = reflect.New(handler.Type.Elem()).Interface()
		err := serde.Deserialize(payload, arg)
		if err != nil {
			return nil, err
		}
	}

	return arg, nil
}
