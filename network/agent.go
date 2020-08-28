package network

import (
	"github.com/lixianmin/gonsole/logger"
	"github.com/lixianmin/gonsole/network/acceptor"
	"github.com/lixianmin/gonsole/network/conn/codec"
	"github.com/lixianmin/gonsole/network/conn/message"
	"github.com/lixianmin/gonsole/network/conn/packet"
	"github.com/lixianmin/got/loom"
)

/********************************************************************
created:    2020-08-28
author:     lixianmin

Copyright (C) - All Rights Reserved
*********************************************************************/

type Agent struct {
	conn    acceptor.PlayerConn
	decoder codec.PacketDecoder
	wc      *loom.WaitClose
}

func NewAgent(conn acceptor.PlayerConn, decoder codec.PacketDecoder) *Agent {
	var agent = &Agent{
		conn:    conn,
		decoder: decoder,
		wc:      loom.NewWaitClose(),
	}

	loom.Go(agent.goRead)
	loom.Go(agent.goWrite)
	return agent
}

func (my *Agent) goRead(later *loom.Later) {
	for {
		msg, err := my.conn.GetNextMessage()

		if err != nil {
			logger.Info("Error reading next available message: %s", err.Error())
			return
		}

		packets, err := my.decoder.Decode(msg)
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
			if err := my.processPacket(packets[i]); err != nil {
				logger.Info("Failed to process packet: %s", err.Error())
				return
			}
		}
	}
}

func (my *Agent) processPacket(p *packet.Packet) error {
	switch p.Type {
	case packet.Handshake:
		logger.Debug("Received handshake packet")
	case packet.HandshakeAck:
		logger.Debug("Receive handshake ACK")
	case packet.Data:
		msg, err := message.Decode(p.Data)
		if err != nil {
			return err
		}
		my.processMessage(msg)
	case packet.Heartbeat:
		// expected
	}

	return nil
}

func (my *Agent) processMessage(msg *message.Message) {
	//r, err := route.Decode(msg.Route)
	//if err != nil {
	//	logger.Warn("Failed to decode route: %s", err.Error())
	//	return
	//}
}

func (my *Agent) goWrite(later *loom.Later) {

}
