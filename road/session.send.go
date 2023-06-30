package road

import (
	"github.com/lixianmin/gonsole/ifs"
	"github.com/lixianmin/gonsole/road/serde"
	"github.com/lixianmin/got/convert"
	"math/rand"
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

func (my *sessionImpl) SendByRoute(route string, v interface{}) error {
	if my.wc.IsClosed() {
		return nil
	}

	if my.serde == nil {
		return ErrInvalidSerde
	}

	var data, err1 = my.serde.Serialize(v)
	if err1 != nil {
		return err1
	}

	var kind, ok = my.manger.GetKindByRoute(route)
	var pack = serde.Packet{Kind: kind, Data: data}
	if !ok {
		var routeData = convert.Bytes(route)
		pack.Kind = serde.RouteBase + int32(len(routeData))
		pack.Route = routeData
	}

	var err2 = my.sendPacket(pack)
	return err2
}

func (my *sessionImpl) SendByKind(kind int32, v interface{}) error {
	if my.wc.IsClosed() {
		return nil
	}

	if my.serde == nil {
		return ErrInvalidSerde
	}

	var data, err1 = my.serde.Serialize(v)
	if err1 != nil {
		return err1
	}

	var pack = serde.Packet{Kind: kind, Data: data}
	var err2 = my.sendPacket(pack)
	return err2
}

// Kick 强踢下线
func (my *sessionImpl) Kick() error {
	if my.wc.IsClosed() {
		return nil
	}

	var _, err = my.link.Write(my.manger.kickBuffer)
	return err
}

func (my *sessionImpl) Handshake() error {
	if my.wc.IsClosed() {
		return nil
	}

	// nonce一定不为0
	var nonce = rand.Int31()
	for nonce == 0 {
		nonce = rand.Int31()
	}

	var info = serde.JsonHandshake{
		Nonce:     nonce,
		Heartbeat: float32(my.manger.heartbeatInterval.Seconds()),
		Routes:    my.manger.routes,
	}

	// all supported serde names
	for _, s := range my.manger.serdes {
		info.Serdes = append(info.Serdes, s.GetName())
	}

	// handshake这个协议一定使用json去发, 后续的协议则可以替换为其它serde方法
	var data, err1 = convert.ToJsonE(info)
	if err1 != nil {
		return err1
	}

	my.Attachment().Put(ifs.KeyNonce, nonce)
	var pack = serde.Packet{Kind: serde.Handshake, Data: data}
	var err2 = my.sendPacket(pack)

	if err2 != nil {
		_ = my.Close()
	}
	return err2
}

func (my *sessionImpl) HandshakeRe(serdeName string) error {
	var s = my.manger.GetSerde(serdeName)
	if s == nil {
		return ErrInvalidSerde
	}
	my.setSerde(s)

	var reply = serde.JsonHandshakeRe{Serde: serdeName}
	var data = convert.ToJson(reply)

	var pack = serde.Packet{Kind: serde.Handshake, Data: data}
	var err2 = my.sendPacket(pack)

	if err2 != nil {
		_ = my.Close()
	}
	return err2
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
