package road

import (
	"maps"
	"math/rand"

	"github.com/lixianmin/gonsole/road/serde"
	"github.com/lixianmin/got/convert"
)

/********************************************************************
created:    2020-08-31
author:     lixianmin

2022-11-26，
1. 从下面死锁的描述来看，同一个connection在一个协程中read，在另一个协程中write是没问题的
2. 但是，同一个connection在不同的协程中异步写是会导致panic的，因此才有了session_sender
3. 但是，除了现在采用启动N协程的处理M个链接（N<M)外，还可以使用lock解决并发问题，修改之

为什么要摘出这样一个类出来？
同一个链接的read/write 不能放到同一个goroutine中。write很直接，但read的handler
中写什么代码是不确定的，如果其中调用到conn.Write()，就有可能形成死锁：
1. read的handler想结束返回就需要write成功
2. 但现在sendingChan满了，所以write无法成功，因此read的handler也无法结束
3. 因为read的handler无法结束，就导致同一个goroutine中的sendingChan的数据无法提取
4. sendingChan中的数据无法提取出来，就一直是满的

Copyright (C) - All Rights Reserved
*********************************************************************/

func (my *sessionImpl) Send(route string, v any) error {
	if route == "" {
		return ErrInvalidRoute
	}

	if my.wc.IsClosed() {
		return nil
	}

	if my.serde == nil {
		return ErrNilSerde
	}

	// 可能是并发访问my.routeKinds, 所以对my.routeKinds的修改都是通过clone实现的
	var kind, ok = my.routeKinds[route]
	var pack = serde.Packet{Kind: kind}
	if !ok {
		// 因为notify相关的逻辑经常使用SendByRoute(), 因此长的route还是挺费的. 但是:
		// 1. kind1不允许在manger中动态计算, 因为manager只有一份, 一个session计算了新kind1, 其它session并不知道, 就不同步了
		// 2. 如果像componentHandler一样, 启动时写死的话, 使用方式会比较别扭
		// 3. 考虑每个session单独计算并各自保存route kind1, 看起来似乎可行
		var kind1, err1 = my.sendRouteKind(route)
		if err1 != nil {
			return err1
		}

		pack.Kind = kind1
	}

	var err2, isError2 = v.(error)
	if !isError2 {
		var payload, err3 = serializeOrRaw(my.serde, v)
		if err3 != nil {
			return err3
		}

		pack.Data = payload
	} else if err4, ok := v.(*Error); ok {
		pack.Code = convert.Bytes(err4.Code)
		pack.Data = convert.Bytes(err4.Message)
	} else {
		pack.Code = convert.Bytes("PlainError")
		pack.Data = convert.Bytes(err2.Error())
	}

	var err5 = my.sendPacket(pack)
	return err5
}

func (my *sessionImpl) sendRouteKind(route string) (int32, error) {
	var kind int32 = 0
	var isSent = false
	my.writeLock.Lock()
	{
		// double check lock
		if kind, isSent = my.routeKinds[route]; !isSent {
			kind = int32(len(my.routeKinds)) + serde.UserBase

			var cloned = maps.Clone(my.routeKinds)
			cloned[route] = kind
			my.routeKinds = cloned // golang中指针赋值是原子操作
		}
	}
	my.writeLock.Unlock()

	if !isSent {
		var v = &serde.JsonRouteKind{
			Kind:  kind,
			Route: route,
		}

		var data, err1 = convert.ToJsonE(v)
		if err1 != nil {
			return 0, err1
		}

		var pack = serde.Packet{Kind: serde.RouteKind, Data: data}
		var err2 = my.sendPacket(pack)
		if err2 != nil {
			return 0, err2
		}
	}

	return kind, nil
}

//func (my *sessionImpl) SendByKind(kind int32, v interface{}) error {
//	if my.wc.IsClosed() {
//		return nil
//	}
//
//	if my.serde == nil {
//		return ErrInvalidSerde
//	}
//
//	var data, err1 = my.serde.Serialize(v)
//	if err1 != nil {
//		return err1
//	}
//
//	var pack = serde.Packet{Kind: kind, Data: data}
//	var err2 = my.sendPacket(pack)
//	return err2
//}

// Kick 强踢下线
func (my *sessionImpl) Kick(reason string) error {
	if my.wc.IsClosed() {
		return nil
	}

	var pack = serde.Packet{Kind: serde.Kick, Data: convert.Bytes(reason)}
	var err1 = my.sendPacket(pack)
	if err1 != nil {
		return err1
	}

	// 发送kick后，关闭链接
	var err2 = my.Close()
	return err2
}

func (my *sessionImpl) Handshake() error {
	if my.wc.IsClosed() {
		return nil
	}

	var nonce = fetchNonce()
	var info = serde.JsonHandshake{
		Nonce:     nonce,
		Heartbeat: float32(my.manager.heartbeatInterval.Seconds()),
		Routes:    my.manager.routes,
		Gid:       my.manager.gid,
		SessionId: my.id, // server的很多日志都是基于sid的, client打印一下这个值, 用于跟server配对
	}

	// all supported serde names
	for name := range my.manager.serdeBuilders {
		info.Serdes = append(info.Serdes, name)
	}

	// handshake这个协议一定使用json去发, 后续的协议则可以替换为其它serde方法
	var data, err1 = convert.ToJsonE(info)
	if err1 != nil {
		return err1
	}

	my.Attachment().Set(keyNonce, nonce)
	var pack = serde.Packet{Kind: serde.Handshake, Data: data}
	var err2 = my.sendPacket(pack)

	if err2 != nil {
		_ = my.Close()
	}

	return err2
}

func (my *sessionImpl) Echo(handler func()) error {
	if my.wc.IsClosed() || handler == nil {
		return nil
	}

	if my.serde == nil {
		return ErrNilSerde
	}

	var requestId = echoIdGenerator.Add(1)
	my.handlerLock.Lock()
	{
		my.echoHandlers[requestId] = handler
	}
	my.handlerLock.Unlock()

	var pack = serde.Packet{Kind: serde.Echo, RequestId: requestId}
	var err3 = my.sendPacket(pack)
	return err3
}

func (my *sessionImpl) sendPacket(pack serde.Packet) error {
	my.writeLock.Lock()
	defer my.writeLock.Unlock()

	var writer = my.writer
	var stream = writer.Stream()
	stream.Reset()
	serde.EncodePacket(writer, pack)

	var buffer = stream.Bytes()
	var _, err = my.link.Write(buffer)
	return err
}

func (my *sessionImpl) setSerde(serde serde.Serde) {
	my.serde = serde
}

func fetchNonce() int32 {
	// nonce一定不为0
	for {
		var nonce = rand.Int31()
		if nonce != 0 {
			return nonce
		}
	}
}
