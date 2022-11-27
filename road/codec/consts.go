package codec

/********************************************************************
created:    2022-09-03
author:     lixianmin

Copyright (C) - All Rights Reserved
*********************************************************************/

// Codec constants.
const (
	HeaderSize    = 4
	MaxPacketSize = 1 << 24 //16MB
)

// client handshake → server handshake → client handshakeAck → server heartbeat
const (
	Handshake    = 0x01 // Handshake represents a handshake: request(client) <====> handshake response(server)
	HandshakeAck = 0x02 // HandshakeAck represents a handshake ack from client to server
	Heartbeat    = 0x03 // Heartbeat represents a heartbeat
	Data         = 0x04 // Data represents a common data packet
	Kick         = 0x05 // Kick represents a kick-off packet, disconnect message from server
)
