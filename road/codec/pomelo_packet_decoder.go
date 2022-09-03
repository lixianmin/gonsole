/********************************************************************
created:    2022-09-03
author:     lixianmin

Copyright (C) - All Rights Reserved
*********************************************************************/

package codec

import (
	"bytes"
)

// PomeloPacketDecoder reads and decodes data slice following pomelo's protocol
type PomeloPacketDecoder struct{}

// NewPomeloPacketDecoder returns a new decoder that used for decode bytes slice.
func NewPomeloPacketDecoder() *PomeloPacketDecoder {
	return &PomeloPacketDecoder{}
}

func (c *PomeloPacketDecoder) forward(buf *bytes.Buffer) (int, PacketType, error) {
	header := buf.Next(HeadLength)
	return ParseHeader(header)
}

// Decode decode the bytes slice to packet.Packet(s)
func (c *PomeloPacketDecoder) Decode(data []byte) ([]*Packet, error) {
	buf := bytes.NewBuffer(nil)
	buf.Write(data)

	var (
		packets []*Packet
		err     error
	)

	// todo 这种玩法会丢失数据的
	// check length
	if buf.Len() < HeadLength {
		return nil, nil
	}

	// first time
	size, typ, err := c.forward(buf)
	if err != nil {
		return nil, err
	}

	for size <= buf.Len() {
		p := &Packet{Type: typ, Length: size, Data: buf.Next(size)}
		packets = append(packets, p)

		// if no more packets, break
		if buf.Len() < HeadLength {
			break
		}

		size, typ, err = c.forward(buf)
		if err != nil {
			return nil, err
		}
	}

	return packets, nil
}
