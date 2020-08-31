package network

import (
	"context"
	"github.com/lixianmin/gonsole/logger"
	"github.com/lixianmin/gonsole/network/conn/message"
	"github.com/lixianmin/gonsole/network/conn/packet"
	"github.com/lixianmin/gonsole/network/route"
	"github.com/lixianmin/got/loom"
)

/********************************************************************
created:    2020-08-31
author:     lixianmin

Copyright (C) - All Rights Reserved
*********************************************************************/

func (my *Session) goReceive(later *loom.Later) {
	for {
		msg, err := my.conn.GetNextMessage()

		if err != nil {
			logger.Info("Error reading next available message: %s", err.Error())
			return
		}

		packets, err := my.packetDecoder.Decode(msg)
		if err != nil {
			logger.Info("Failed to decode message: %s", err.Error())
			return
		}

		if len(packets) < 1 {
			logger.Warn("Read no packets, data: %v", msg)
			continue
		}

		// process all packet
		for i := range packets {
			var p = packets[i]
			var item, err = my.processReceivedPacket(p)
			if err != nil {
				logger.Info("Failed to process packet: %s", err.Error())
				return
			}

			if p.Type == packet.Data {
				select {
				case my.receivedChan <- item:
				case <-my.wc.C():
					return
				}
			}
		}
	}
}

func (my *Session) processReceivedPacket(p *packet.Packet) (receivedItem, error) {
	switch p.Type {
	case packet.Handshake:
		return my.onReceivedHandshake(p)
	case packet.HandshakeAck:
		logger.Debug("Receive handshake ACK")
	case packet.Data:
		return my.onReceivedDataPacket(p)
	case packet.Heartbeat:
		// expected
	}

	return receivedItem{}, nil
}

func (my *Session) onReceivedHandshake(p *packet.Packet) (receivedItem, error) {
	_, err := my.conn.Write(my.handshakeResponseData)
	if err != nil {
		return receivedItem{}, err
	}

	//logger.Debug("Received handshake packet")
	return receivedItem{}, nil
}

func (my *Session) onReceivedDataPacket(p *packet.Packet) (receivedItem, error) {
	msg, err := message.Decode(p.Data)
	if err != nil {
		return receivedItem{}, err
	}

	r, err := route.Decode(msg.Route)
	if err != nil {
		return receivedItem{}, err
	}

	var item = receivedItem{
		ctx:   context.Background(),
		route: r,
		msg:   msg,
	}

	return item, nil
}
