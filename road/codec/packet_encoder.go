/********************************************************************
created:    2022-09-03
author:     lixianmin

Copyright (C) - All Rights Reserved
*********************************************************************/

package codec

// PacketEncoder interface
type PacketEncoder interface {
	Encode(typ PacketType, data []byte) ([]byte, error)
}
