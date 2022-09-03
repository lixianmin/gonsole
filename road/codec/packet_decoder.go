package codec

/********************************************************************
created:    2022-09-03
author:     lixianmin

Copyright (C) - All Rights Reserved
*********************************************************************/

// PacketDecoder interface
type PacketDecoder interface {
	Decode(data []byte) ([]*Packet, error)
}
