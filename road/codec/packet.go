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
	Kind   PacketKind
	Length int
	Data   []byte
}

func (p *Packet) String() string {
	return fmt.Sprintf("Kind: %d, Length: %d, Data: %s", p.Kind, p.Length, string(p.Data))
}
