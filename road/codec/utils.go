package codec

/********************************************************************
created:    2022-09-03
author:     lixianmin

Copyright (C) - All Rights Reserved
*********************************************************************/

// ParseHeader parses a packet header and returns its dataLen and packetType or an error
func ParseHeader(header []byte) (int, PacketKind, error) {
	if len(header) != HeaderSize {
		return 0, 0x00, ErrInvalidPomeloHeader
	}

	kind := header[0]
	if kind < Handshake || kind > Kick {
		return 0, 0x00, ErrWrongPomeloPacketKind
	}

	var size = BytesToInt(header[1:])

	if size > MaxPacketSize {
		return 0, 0x00, ErrPacketSizeExceed
	}

	return size, kind, nil
}

// BytesToInt32 decode packet data length byte to int(Big end)
func BytesToInt(b []byte) int {
	var result = 0
	for _, v := range b {
		result = result<<8 + int(v)
	}

	return result
}

//// IntToBytes encode packet data length to bytes(Big end)
//func IntToBytes(n int) []byte {
//	buf := make([]byte, 3)
//	buf[0] = byte((n >> 16) & 0xFF)
//	buf[1] = byte((n >> 8) & 0xFF)
//	buf[2] = byte(n & 0xFF)
//	return buf
//}
