package codec

import (
	"fmt"
)

/********************************************************************
created:    2022-09-03
author:     lixianmin

Copyright (C) - All Rights Reserved
*********************************************************************/

// PacketKind represents the packet's type such as: handshake and so on.
type PacketKind = byte

type Packet struct {
	Kind PacketKind
	Size int32
	Data []byte
}

func (p *Packet) String() string {
	return fmt.Sprintf("Kind: %d, Size: %d, Data: %s", p.Kind, p.Size, string(p.Data))
}
