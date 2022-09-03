package codec

/********************************************************************
created:    2022-09-03
author:     lixianmin

Copyright (C) - All Rights Reserved
*********************************************************************/

// PomeloPacketEncoder struct
type PomeloPacketEncoder struct {
}

// NewPomeloPacketEncoder ctor
func NewPomeloPacketEncoder() *PomeloPacketEncoder {
	return &PomeloPacketEncoder{}
}

// Encode create a packet.Packet from  the raw bytes slice and then encode to bytes slice
// Protocol refs: https://github.com/NetEase/pomelo/wiki/Communication-Protocol
//
// -<type>-|--------<length>--------|-<data>-
// --------|------------------------|--------
// 1 byte packet type, 3 bytes packet data length(big end), and data segment
func (my *PomeloPacketEncoder) Encode(kind PacketKind, data []byte) ([]byte, error) {
	if kind < Handshake || kind > Kick {
		return nil, ErrWrongPomeloPacketKind
	}

	if len(data) > MaxPacketSize {
		return nil, ErrPacketSizeExceed
	}

	var size = int32(len(data))
	var p = &Packet{Kind: kind, Size: size}

	var buf = make([]byte, HeaderLength+size)
	buf[0] = p.Kind
	buf[1] = byte((size >> 16) & 0xFF)
	buf[2] = byte((size >> 8) & 0xFF)
	buf[3] = byte(size & 0xFF)

	copy(buf[HeaderLength:], data)
	return buf, nil
}
