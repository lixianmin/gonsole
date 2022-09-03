// Copyright (c) nano Author and TFG Co. All Rights Reserved.
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in all
// copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.

package codec

import (
	"github.com/lixianmin/gonsole/ifs"
	"github.com/lixianmin/gonsole/road/internal"
)

// PomeloPacketEncoder struct
type PomeloPacketEncoder struct {
}

// NewPomeloPacketEncoder ctor
func NewPomeloPacketEncoder() *PomeloPacketEncoder {
	return &PomeloPacketEncoder{}
}

// Encode create a packet.Packet from  the raw bytes slice and then encode to bytes slice
// Protocol refs: https://github.com/NetEase/pomelo/wiki/Communication-Protocol
//
// -<type>-|--------<length>--------|-<data>-
// --------|------------------------|--------
// 1 byte packet type, 3 bytes packet data length(big end), and data segment
func (e *PomeloPacketEncoder) Encode(typ internal.PacketType, data []byte) ([]byte, error) {
	if typ < internal.Handshake || typ > internal.Kick {
		return nil, ifs.ErrWrongPomeloPacketType
	}

	if len(data) > MaxPacketSize {
		return nil, ifs.ErrPacketSizeExceed
	}

	p := &internal.Packet{Type: typ, Length: len(data)}
	buf := make([]byte, p.Length+HeadLength)
	buf[0] = p.Type

	copy(buf[1:HeadLength], IntToBytes(p.Length))
	copy(buf[HeadLength:], data)

	return buf, nil
}
