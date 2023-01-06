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
		var remains = buffer.Bytes()
		if len(remains) < HeaderSize {
			return packets, nil
		}

		var bodySize, kind, err = ParseHeader(remains[:HeaderSize])
		if err != nil {
			return nil, err
		}

		var totalSize = HeaderSize + bodySize
		if len(remains) < totalSize {
			return packets, nil
		}

		buffer.Next(totalSize)
		var p = &Packet{Kind: kind, Size: int32(bodySize), Data: remains[HeaderSize:totalSize]}
		packets = append(packets, p)
	}
}
