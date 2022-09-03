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

	p := &Packet{Kind: kind, Size: int32(len(data))}
	buf := make([]byte, p.Size+HeaderLength)
	buf[0] = p.Kind

	copy(buf[1:HeaderLength], IntToBytes(int(p.Size)))
	copy(buf[HeaderLength:], data)

	return buf, nil
}
