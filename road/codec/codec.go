package codec

import "github.com/lixianmin/got/iox"

/********************************************************************
created:    2023-06-05
author:     lixianmin

Copyright (C) - All Rights Reserved
*********************************************************************/

type Codec interface {
	Decode(buffer *iox.Buffer) ([]*Packet, error)
	Encode(kind int32, data []byte) ([]byte, error)
}
