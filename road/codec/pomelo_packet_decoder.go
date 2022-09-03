package codec

import (
	"github.com/lixianmin/got/iox"
)

/********************************************************************
created:    2022-09-03
author:     lixianmin

Copyright (C) - All Rights Reserved
*********************************************************************/

// PomeloPacketDecoder reads and decodes data slice following pomelo's protocol
type PomeloPacketDecoder struct {
	buffer *iox.Buffer
}

// NewPomeloPacketDecoder returns a new decoder that used for decode bytes slice.
func NewPomeloPacketDecoder() *PomeloPacketDecoder {
	var my = &PomeloPacketDecoder{
		buffer: &iox.Buffer{},
	}

	return my
}

// Decode decode the bytes slice to packet.Packet(s)
func (my *PomeloPacketDecoder) Decode(data []byte) ([]*Packet, error) {
	var buf = my.buffer
	if _, err := buf.Write(data); err != nil {
		return nil, err
	}

	var packets []*Packet = nil
	for {
		if buf.Len() < HeaderLength {
			return packets, nil
		}

		buf.MakeCheckpoint()
		var size, kind, err = ParseHeader(buf.Next(HeaderLength))
		if err != nil {
			return nil, err
		}

		if buf.Len() < size {
			buf.RestoreCheckpoint()
			return packets, nil
		}

		var p = &Packet{Kind: kind, Size: int32(size), Data: buf.Next(size)}
		packets = append(packets, p)
	}
}
