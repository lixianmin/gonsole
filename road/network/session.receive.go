package network

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
	go my.link.GoLoop(my.manger.heartbeatInterval, func(reader *iox.OctetsReader, err error) {
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
	var packets, err1 = serde.Decode(reader)
	if err1 != nil {
		var err2 = fmt.Errorf("failed to decode message: %s", err1.Error())
		return err2
	}

	var handler = my.onReceivingPacketHandler
	for _, pack := range packets {
		if handler != nil {
			var err3 = handler(pack)
			if err3 == ErrPacketProcessed {
				continue
			} else if err3 != nil {
				return err3
			}
		}

		if pack.Kind >= serde.UserDefined {
			if err4 := my.onReceivedOther(pack); err4 != nil {
				return err4
			}
		} else if pack.Kind == serde.Heartbeat {
			// 现在server只有一个goroutine用于阻塞式读取网络数据，因此server缺少定时发送heartbeat的能力，因此采用client主动heartbeat而server回复的方案
			var pack = serde.Packet{Kind: serde.Heartbeat}
			if err5 := my.writePacket(pack); err5 != nil {
				return err5
			}
		}
	}

	return nil
}

func (my *sessionImpl) onReceivedOther(input serde.Packet) error {
	var handler = my.manger.GetHandlerByKind(input.Kind)
	if handler == nil {
		return ErrEmptyHandler
	}

	// 这个err不能立即返回，这是业务逻辑错误, 应该输出到client, 而不应该引发session.Close()
	var payload, err = processReceivedPacket(input, my.ctxValue, handler, my.manger.GetSerde())
	var output = serde.Packet{
		Kind: input.Kind,
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

	return my.writePacket(output)
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