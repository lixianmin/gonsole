package network

import (
	"context"
	"fmt"
	"github.com/lixianmin/gonsole/logger"
	"github.com/lixianmin/gonsole/network/conn/message"
	"github.com/lixianmin/gonsole/network/conn/packet"
	"github.com/lixianmin/gonsole/network/util"
)

/********************************************************************
created:    2020-08-31
author:     lixianmin

Copyright (C) - All Rights Reserved
*********************************************************************/

func (a *Agent) ResponseMID(ctx context.Context, mid uint, v interface{}) error {
	return a.send(pendingMessage{ctx: ctx, typ: message.Response, mid: mid, payload: v, err: false})
}

func (my *Agent) Push(route string, v interface{}) error {
	return my.send(pendingMessage{typ: message.Push, route: route, payload: v})
}

func (my *Agent) send(pendingMsg pendingMessage) error {
	defer func() {
		if e := recover(); e != nil {
			logger.Info(e)
		}
	}()

	m, err := my.getMessageFromPendingMessage(pendingMsg)
	if err != nil {
		return err
	}

	// packet encode
	p, err := my.packetEncodeMessage(m)
	if err != nil {
		return err
	}

	pending := pendingWrite{
		ctx:  pendingMsg.ctx,
		data: p,
	}

	if pendingMsg.err {
		pending.err = fmt.Errorf("has pending error")
	}

	select {
	case <-my.wc.C():
	case my.chSend <- pending:
	}

	return nil
}

func (my *Agent) getMessageFromPendingMessage(pm pendingMessage) (*message.Message, error) {
	payload, err := util.SerializeOrRaw(my.serializer, pm.payload)
	if err != nil {
		return nil, err
	}

	// construct message and encode
	m := &message.Message{
		Type:  pm.typ,
		Data:  payload,
		Route: pm.route,
		ID:    pm.mid,
		Err:   pm.err,
	}

	return m, nil
}

func (my *Agent) packetEncodeMessage(m *message.Message) ([]byte, error) {
	em, err := my.messageEncoder.Encode(m)
	if err != nil {
		return nil, err
	}

	// packet encode
	p, err := my.packetEncoder.Encode(packet.Data, em)
	if err != nil {
		return nil, err
	}
	return p, nil
}
