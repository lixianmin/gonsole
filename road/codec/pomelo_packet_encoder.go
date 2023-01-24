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

	// >16M返回
	if len(data) > MaxPacketSize {
		return nil, ErrPacketSizeExceed
	}

	//构建packet对象
	var size = int32(len(data))
	var p = &Packet{Kind: kind, Size: size}

	//构建buf，大小为（1+3）+包的大小
	var buf = make([]byte, HeadSize+size)
	buf[0] = p.Kind
	buf[1] = byte((size >> 16) & 0xFF)
	buf[2] = byte((size >> 8) & 0xFF)
	buf[3] = byte(size & 0xFF)

	//填充数据
	copy(buf[HeadSize:], data)
	return buf, nil
}
