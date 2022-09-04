package codec

import "github.com/lixianmin/got/iox"

/********************************************************************
created:    2022-09-03
author:     lixianmin

Copyright (C) - All Rights Reserved
*********************************************************************/

// PacketDecoder interface
type PacketDecoder interface {
	Decode(buffer *iox.Buffer) ([]*Packet, error)
}
