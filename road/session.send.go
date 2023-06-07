package road

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

func (my *sessionImpl) ShakeHand(heartbeat float32) error {
	if my.wc.IsClosed() {
		return nil
	}

	type Handshake struct {
		Nonce     int32   `json:"nonce"`
		Heartbeat float32 `json:"heartbeat"` // 心跳间隔. 单位: 秒
	}

	var nonce = rand.Int31()
	var item = Handshake{Nonce: nonce, Heartbeat: heartbeat}
	var data, err1 = my.manger.GetSerde().Serialize(item)
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
