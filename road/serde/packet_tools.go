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

func EncodePacket(writer *iox.OctetsWriter, pack Packet) {
	_ = writer.Write7BitEncodedInt(pack.Kind)

	// pack.Kind是kind还是RouteBase+len(route), 这是在调用EncodePacket之前就准备好的
	if pack.Kind > RouteBase {
		_ = writer.Stream().Write(pack.Route)
	}

	_ = writer.Write7BitEncodedInt(pack.RequestId)
	_ = writer.WriteBytes(pack.Code)
	_ = writer.WriteBytes(pack.Data)
}

func DecodePacket(reader *iox.OctetsReader) ([]Packet, error) {
	var packets []Packet = nil
	var stream = reader.Stream()

	for {
		var lastPosition = stream.Position()

		var kind, err = reader.Read7BitEncodedInt()
		if err == iox.ErrNotEnoughData {
			rewindStream(stream, lastPosition)
			return packets, nil
		}

		var route []byte = nil
		if kind > RouteBase {
			var size = kind - RouteBase
			var data = make([]byte, size)
			var num, err2 = stream.Read(data)
			if err2 == iox.ErrNotEnoughData || num != int(size) {
				rewindStream(stream, lastPosition)
				return packets, nil
			}

			route = data
		}

		requestId, err := reader.Read7BitEncodedInt()
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

		var pack = Packet{Kind: kind, Route: route, RequestId: requestId, Code: code, Data: data}
		packets = append(packets, pack)
	}
}

func rewindStream(stream *iox.OctetsStream, position int) {
	_, _ = stream.Seek(int64(position), io.SeekStart)
}
