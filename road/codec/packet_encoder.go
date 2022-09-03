package codec

/********************************************************************
created:    2022-09-03
author:     lixianmin

Copyright (C) - All Rights Reserved
*********************************************************************/

// PacketEncoder interface
type PacketEncoder interface {
	Encode(typ PacketKind, data []byte) ([]byte, error)
}