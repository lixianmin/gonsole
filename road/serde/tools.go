package serde

import (
	"github.com/lixianmin/got/iox"
	"io"
)

/********************************************************************
created:    2023-06-05
author:     lixianmin

Copyright (C) - All Rights Reserved
*********************************************************************/

func Encode(writer *iox.OctetsWriter, pack Packet) {
	_ = writer.Write7BitEncodedInt(pack.Kind)
	_ = writer.WriteBytes(pack.Code)
	_ = writer.WriteBytes(pack.Data)
}

func Decode(reader *iox.OctetsReader) ([]Packet, error) {
	var packets []Packet = nil
	var stream = reader.Stream()

	for {
		var lastPosition = stream.Position()

		var kind, err = reader.Read7BitEncodedInt()
		if err == iox.ErrNotEnoughData {
			rewindStream(stream, lastPosition)
			return packets, nil
		}

		code, err := reader.ReadBytes()
		if err == iox.ErrNotEnoughData {
			rewindStream(stream, lastPosition)
			return packets, nil
		}

		data, err := reader.ReadBytes()
		if err == iox.ErrNotEnoughData {
			rewindStream(stream, lastPosition)
			return packets, nil
		}

		var pack = Packet{Kind: kind, Code: code, Data: data}
		packets = append(packets, pack)
	}
}

func rewindStream(stream *iox.OctetsStream, position int) {
	_, _ = stream.Seek(int64(position), io.SeekStart)
}
