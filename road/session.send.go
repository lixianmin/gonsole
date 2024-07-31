package road

import (
	"github.com/lixianmin/gonsole/road/serde"
	"github.com/lixianmin/got/convert"
	"maps"
	"math/rand"
	"sync/atomic"
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
	if my.wc.IsClosed() {
		return nil
	}

	if my.serde == nil {
		return ErrNilSerde
	}

	var data, err1 = my.serde.Serialize(v)
	if err1 != nil {
		return err1
	}

	// 可能是并发访问my.routeKinds
	var kind, ok = my.routeKinds[route]
	var pack = serde.Packet{Kind: kind, Data: data}
	if !ok {
		// 因为notify相关的逻辑经常使用SendByRoute(), 因此长的route还是挺费的. 但是:
		// 1. kind不允许在manger中动态计算, 因为manager只有一份, 一个session计算了新kind, 其它session并不知道, 就不同步了
		// 2. 如果像componentHandler一样, 启动时写死的话, 使用方式会比较别扭
		// 3. 考虑每个session单独计算并各自保存route kind, 看起来似乎可行
		var kind2, err2 = my.sendRouteKind(route)
		if err2 != nil {
			return err2
		}

		pack.Kind = kind2
	}

	var err3 = my.sendPacket(pack)
	return err3
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

	var data, err1 = my.serde.Serialize(reason)
	if err1 != nil {
		return err1
	}

	var pack = serde.Packet{Kind: serde.Kick, Data: data}
	var err2 = my.sendPacket(pack)
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

	var requestId = atomic.AddInt32(&echoIdGenerator, 1)
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
