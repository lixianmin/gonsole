package network

import (
	"context"
	"fmt"
	"github.com/lixianmin/gonsole/logger"
	"github.com/lixianmin/gonsole/network/conn/message"
	"github.com/lixianmin/gonsole/network/conn/packet"
	"github.com/lixianmin/gonsole/network/util"
	"github.com/lixianmin/got/loom"
)

/********************************************************************
created:    2020-08-31
author:     lixianmin

Copyright (C) - All Rights Reserved
*********************************************************************/

func (my *Agent) goSend(later *loom.Later) {
	defer func() {
		my.Close()
	}()

	for {
		select {
		case item := <-my.sendingChan:
			if _, err := my.conn.Write(item.data); err != nil {
				logger.Info("Failed to write in conn: %s", err.Error())
				return
			}
		case <-my.wc.C():
			return
		}
	}
}

func (my *Agent) ResponseMID(ctx context.Context, mid uint, v interface{}) error {
	return my.send(sendingInfo{ctx: ctx, typ: message.Response, mid: mid, payload: v, err: false})
}

func (my *Agent) Push(route string, v interface{}) error {
	return my.send(sendingInfo{typ: message.Push, route: route, payload: v})
}

func (my *Agent) send(info sendingInfo) error {
	defer func() {
		if e := recover(); e != nil {
			logger.Info(e)
		}
	}()

	payload, err := util.SerializeOrRaw(my.serializer, info.payload)
	if err != nil {
		return err
	}

	// construct message and encode
	m := &message.Message{
		Type:  info.typ,
		Data:  payload,
		Route: info.route,
		ID:    info.mid,
		Err:   info.err,
	}

	// packet encode
	p, err := my.packetEncodeMessage(m)
	if err != nil {
		return err
	}

	item := sendingItem{
		ctx:  info.ctx,
		data: p,
	}

	if info.err {
		item.err = fmt.Errorf("has pending error")
	}

	select {
	case <-my.wc.C():
	case my.sendingChan <- item:
	}

	return nil
}

func (my *Agent) packetEncodeMessage(msg *message.Message) ([]byte, error) {
	data, err := my.messageEncoder.Encode(msg)
	if err != nil {
		return nil, err
	}

	// packet encode
	p, err := my.packetEncoder.Encode(packet.Data, data)
	if err != nil {
		return nil, err
	}

	return p, nil
}
