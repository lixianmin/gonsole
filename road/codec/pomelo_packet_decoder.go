package codec

import (
	"bytes"
)

/********************************************************************
created:    2022-09-03
author:     lixianmin

Copyright (C) - All Rights Reserved
*********************************************************************/

// PomeloPacketDecoder reads and decodes data slice following pomelo's protocol
type PomeloPacketDecoder struct {
	buffer *bytes.Buffer
}

// NewPomeloPacketDecoder returns a new decoder that used for decode bytes slice.
func NewPomeloPacketDecoder() *PomeloPacketDecoder {
	var my = &PomeloPacketDecoder{
		buffer: bytes.NewBuffer(nil),
	}

	return my
}

// Decode decode the bytes slice to packet.Packet(s)
func (my *PomeloPacketDecoder) Decode(data []byte) ([]*Packet, error) {
	var buf = my.buffer
	buf.Write(data)

	var (
		packets []*Packet
		err     error
	)

	// check length
	if buf.Len() < HeaderLength {
		return nil, nil
	}

	// first time
	size, kind, err := ParseHeader(buf.Next(HeaderLength))
	if err != nil {
		return nil, err
	}

	for size <= buf.Len() {
		p := &Packet{Kind: kind, Length: size, Data: buf.Next(size)}
		packets = append(packets, p)

		// if no more packets, break
		if buf.Len() < HeaderLength {
			break
		}

		size, kind, err = ParseHeader(buf.Next(HeaderLength))
		if err != nil {
			return nil, err
		}
	}

	return packets, nil
}
