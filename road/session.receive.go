package road

import (
	"fmt"
	"reflect"

	"github.com/lixianmin/gonsole/road/component"
	"github.com/lixianmin/gonsole/road/serde"
	"github.com/lixianmin/got/convert"
	"github.com/lixianmin/got/iox"
	"github.com/lixianmin/logo"
)

/********************************************************************
created:    2020-08-31
author:     lixianmin

Copyright (C) - All Rights Reserved
*********************************************************************/

var typeOfBytes = reflect.TypeOf(([]byte)(nil))

func (my *sessionImpl) startGoLoop() {
	go my.link.GoLoop(my.manager.kickInterval, func(reader *iox.OctetsReader, err error) {
		if err != nil {
			logo.Info("close session(%d) by err, addr=%s, err=%q", my.id, my.link.RemoteAddr(), err)
			_ = my.Close()
			return
		}

		if err1 := my.onReceivedData(reader); err1 != nil {
			logo.Info("close session(%d) by onReceivedData(), addr=%s, err=%q", my.id, my.link.RemoteAddr(), err1)
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

	// 这是业务逻辑错误, 应该输出到client, 不应该引发session.Close()
	var requestArg, err2 = unmarshalRequestArg(handler, my.serde, input.Data)
	if err2 != nil {
		return err2
	}

	// 3个参数+返回值的情况
	if handler.RespondMethodType == nil {
		var args = []reflect.Value{handler.Receiver, my.ctxValue, reflect.ValueOf(requestArg)}
		var response, err3 = callMethod(handler.Method, args)
		return my.respondWith(input, response, err3)
	}

	// 第4个参数是respond
	//var respond = func(response any, err error) {
	//	if err4 := my.respondWith(input, response, err); err4 != nil {
	//		logo.Info("close session(%d) by respondWith(), err=%q", my.id, err4)
	//		_ = my.Close()
	//	}
	//}

	// 因为使用了input这种upvalue, 因此很难直接缓存respondValue
	// 虽然MakeFunc的成本很高, 但这里主要会用于处理耗时比较长的调用LLM之类的请求, 因此也没有很急切优化的必要
	var respondValue = reflect.MakeFunc(handler.RespondMethodType, func(args []reflect.Value) []reflect.Value {
		var response = args[0].Interface()
		var err4, _ = args[1].Interface().(error)

		if err5 := my.respondWith(input, response, err4); err5 != nil {
			logo.Info("close session(%d) by respondWith(), err5=%q", my.id, err5)
			_ = my.Close()
		}
		return nil
	})

	var args = []reflect.Value{handler.Receiver, my.ctxValue, reflect.ValueOf(requestArg), respondValue}
	_, _ = callMethod(handler.Method, args)
	return nil
}

func (my *sessionImpl) respondWith(input serde.Packet, response any, err error) error {
	// respondWith()可能很久以后才回调过来的, 加点儿判断, 早一步拦截
	if my.wc.IsClosed() {
		return nil
	}

	// 这个方法有可能从非session线程中回调回来, 因此要求my.serder必须是thread-safe的
	var payload []byte
	if err == nil {
		payload, err = serializeOrRaw(my.serde, response)
	}

	var pack = serde.Packet{
		Kind:      input.Kind,
		RequestId: input.RequestId,
	}

	if err == nil {
		pack.Data = payload
	} else if err2, ok := err.(*Error); ok {
		pack.Code = convert.Bytes(err2.Code)
		pack.Data = convert.Bytes(err2.Message)
	} else {
		pack.Code = convert.Bytes("PlainError")
		pack.Data = convert.Bytes(err.Error())
	}

	return my.sendPacket(pack)
}

func unmarshalRequestArg(handler *component.Handler, serde serde.Serde, payload []byte) (any, error) {
	if handler.RequestType == typeOfBytes {
		return payload, nil
	}

	var request = reflect.New(handler.RequestType.Elem()).Interface()
	var err = serde.Deserialize(payload, request)
	if err != nil {
		return nil, err
	}

	return request, nil
}
