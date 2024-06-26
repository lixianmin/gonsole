package road

import (
	"fmt"
	//"github.com/lixianmin/gonsole/road"
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
	go my.link.GoLoop(my.manger.kickInterval, func(reader *iox.OctetsReader, err error) {
		if err != nil {
			logo.Info("close session(%d) by err=%q", my.id, err)
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
		if err4 := my.onReceivedUserdata(pack); err4 != nil {
			return err4
		}
	} else if pack.Kind == serde.Heartbeat {
		// 现在server只有一个goroutine用于阻塞式读取网络数据，因此server缺少定时发送heartbeat的能力，因此采用client主动heartbeat而server回复的方案
		if _, err6 := my.link.Write(my.manger.heartbeatBuffer); err6 != nil {
			return err6
		}
	} else if pack.Kind == serde.HandshakeRe {
		if err5 := my.onReceivedHandshakeRe(pack); err5 != nil {
			return err5
		}
	}

	return nil
}

func (my *sessionImpl) onReceivedHandshakeRe(input serde.Packet) error {
	var info serde.JsonHandshakeRe
	var err = convert.FromJsonE(input.Data, &info)
	if err != nil {
		return err
	}

	var s = my.manger.GetSerde(info.Serde)
	if s == nil {
		return ErrInvalidSerde
	}

	my.setSerde(s)

	var handler = my.onHandShakenHandler
	if handler != nil {
		handler()
	}
	return nil
}

func (my *sessionImpl) onReceivedUserdata(input serde.Packet) error {
	// client发来的消息, 必须有handler, 因此一定有kind才是合理的. server推送的消息可以没有kind
	var handler = my.manger.GetHandlerByKind(input.Kind)
	if handler == nil {
		return ErrEmptyHandler
	}

	if my.serde == nil {
		return ErrInvalidSerde
	}

	// 这个err不能立即返回，这是业务逻辑错误, 应该输出到client, 而不应该引发session.Close()
	var payload, err = processReceivedPacket(input, my.ctxValue, handler, my.serde)
	var output = serde.Packet{
		Kind:      input.Kind,
		RequestId: input.RequestId,
	}

	if err == nil {
		output.Data = payload
	} else if err1, ok := err.(*Error); ok {
		output.Code = convert.Bytes(err1.Code)
		output.Data = convert.Bytes(err1.Message)
	} else {
		output.Code = convert.Bytes("PlainError")
		output.Data = convert.Bytes(err.Error())
	}

	return my.sendPacket(output)
}

func processReceivedPacket(pack serde.Packet, ctxValue reflect.Value, handler *component.Handler, serde serde.Serde) ([]byte, error) {
	// First unmarshal the handler argument that will be passed to
	// both handler and pipeline functions
	var arg, err = unmarshalHandlerArg(handler, serde, pack.Data)
	if err != nil {
		return nil, err
	}

	var args []reflect.Value
	if arg != nil {
		args = []reflect.Value{handler.Receiver, ctxValue, reflect.ValueOf(arg)}
	} else {
		args = []reflect.Value{handler.Receiver, ctxValue}
	}

	var response, err2 = PCall(handler.Method, args)
	if err2 != nil {
		return nil, err2
	}

	var ret, err3 = serializeOrRaw(serde, response)
	if err3 != nil {
		return nil, err3
	}

	return ret, nil
}

func unmarshalHandlerArg(handler *component.Handler, serde serde.Serde, payload []byte) (interface{}, error) {
	if handler.IsRawArg {
		return payload, nil
	}

	var arg interface{}
	if handler.Type != nil {
		arg = reflect.New(handler.Type.Elem()).Interface()
		err := serde.Deserialize(payload, arg)
		if err != nil {
			return nil, err
		}
	}

	return arg, nil
}
