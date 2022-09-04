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
type PomeloPacketDecoder struct{}

// NewPomeloPacketDecoder returns a new decoder that used for decode bytes slice.
func NewPomeloPacketDecoder() *PomeloPacketDecoder {
	var my = &PomeloPacketDecoder{}
	return my
}

// Decode decode the bytes slice to packet.Packet(s)
func (my *PomeloPacketDecoder) Decode(buffer *iox.Buffer) ([]*Packet, error) {
	defer buffer.Tidy()

	var packets []*Packet = nil
	for {
		if buffer.Len() < HeaderLength {
			return packets, nil
		}

		buffer.MakeCheckpoint()
		var size, kind, err = ParseHeader(buffer.Next(HeaderLength))
		if err != nil {
			return nil, err
		}

		if buffer.Len() < size {
			buffer.RestoreCheckpoint()
			return packets, nil
		}

		var p = &Packet{Kind: kind, Size: int32(size), Data: buffer.Next(size)}
		packets = append(packets, p)
	}
}
