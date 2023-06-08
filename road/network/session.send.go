package network

import (
	"github.com/lixianmin/gonsole/ifs"
	"github.com/lixianmin/gonsole/road/serde"
	"math/rand"
)

/********************************************************************
created:    2020-08-31
author:     lixianmin

Copyright (C) - All Rights Reserved
*********************************************************************/

func (my *sessionImpl) Push(route string, v interface{}) error {
	if my.wc.IsClosed() {
		return nil
	}

	var kind, ok = my.manger.GetKindByRoute(route)
	if !ok {
		return ErrInvalidRoute
	}

	var data, err1 = my.manger.GetSerde().Serialize(v)
	if err1 != nil {
		return err1
	}

	var pack = serde.Packet{Kind: kind, Data: data}
	var err2 = my.writePacket(pack)
	return err2
}

// Kick 强踢下线
func (my *sessionImpl) Kick() error {
	if my.wc.IsClosed() {
		return nil
	}

	var pack = serde.Packet{Kind: serde.Kick}
	var err = my.writePacket(pack)
	return err
}

func (my *sessionImpl) Handshake() error {
	if my.wc.IsClosed() {
		return nil
	}

	var nonce = rand.Int31()
	var info = serde.HandshakeInfo{
		Nonce:      nonce,
		Heartbeat:  float32(my.manger.heartbeatInterval.Seconds()),
		RouteKinds: my.manger.routeKinds,
	}

	var data, err1 = my.manger.GetSerde().Serialize(info)
	if err1 != nil {
		return err1
	}

	my.Attachment().Put(ifs.KeyNonce, nonce)
	var pack = serde.Packet{Kind: serde.Handshake, Data: data}
	var err2 = my.writePacket(pack)

	if err2 != nil {
		_ = my.Close()
	}
	return err2
}

func (my *sessionImpl) writePacket(pack serde.Packet) error {
	my.writeLock.Lock()
	defer my.writeLock.Unlock()

	var writer = my.writer
	var stream = writer.GetStream()
	stream.Reset()
	serde.Encode(writer, pack)

	var buffer = stream.Bytes()
	var _, err = my.conn.Write(buffer)
	return err
}
