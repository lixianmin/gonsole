/********************************************************************
created:    2022-09-03
author:     lixianmin

Copyright (C) - All Rights Reserved
*********************************************************************/

package codec

import (
	"github.com/lixianmin/gonsole/ifs"
)

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
func (e *PomeloPacketEncoder) Encode(typ PacketType, data []byte) ([]byte, error) {
	if typ < Handshake || typ > Kick {
		return nil, ifs.ErrWrongPomeloPacketType
	}

	if len(data) > MaxPacketSize {
		return nil, ifs.ErrPacketSizeExceed
	}

	p := &Packet{Type: typ, Length: len(data)}
	buf := make([]byte, p.Length+HeadLength)
	buf[0] = p.Type

	copy(buf[1:HeadLength], IntToBytes(p.Length))
	copy(buf[HeadLength:], data)

	return buf, nil
}
