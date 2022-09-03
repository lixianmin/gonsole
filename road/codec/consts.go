package codec

/********************************************************************
created:    2022-09-03
author:     lixianmin

Copyright (C) - All Rights Reserved
*********************************************************************/

// Codec constants.
const (
	HeadLength    = 4
	MaxPacketSize = 1 << 24 //16MB
)

const (
	_ PacketKind = iota
	// Handshake represents a handshake: request(client) <====> handshake response(server)
	Handshake = 0x01

	// HandshakeAck represents a handshake ack from client to server
	HandshakeAck = 0x02

	// Heartbeat represents a heartbeat
	Heartbeat = 0x03

	// Data represents a common data packet
	Data = 0x04

	// Kick represents a kick off packet
	Kick = 0x05 // disconnect message from server
)
