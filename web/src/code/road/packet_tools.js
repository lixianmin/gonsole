'use strict'

/********************************************************************
 created:    2023-06-09
 author:     lixianmin

 Copyright (C) - All Rights Reserved
 *********************************************************************/
import {SeekOrigin} from "@src/code/iox/octets_stream";
import {PacketKind} from "@src/code/road/consts";

export function encode(writer, pack) {
    writer.write7BitEncodedInt(pack.kind)

    // pack.kind是kind还是RouteBase+len(route), 这是在调用EncodePacket之前就准备好的
    // route需要是一个[]byte而不能是string, 因为需要在外围计算route的长度, 并赋值到kind, 计算前我们就已经拿到[]byte了
    if (pack.kind > PacketKind.RouteBase) {
        writer.stream.write(pack.route, 0, pack.route.length)
    }

    writer.write7BitEncodedInt(pack.requestId)
    writer.writeBytes(pack.code)
    writer.writeBytes(pack.data)
}

export function decode(reader) {
    const packets = []
    const stream = reader.stream

    while (true) {
        const lastPosition = stream.position
        try {
            const kind = reader.read7BitEncodedInt()

            let route = undefined
            if (kind > PacketKind.RouteBase) {
                const size = kind - PacketKind.RouteBase
                route = new Uint8Array(size)
                stream.read(route, 0, size)
            }

            const requestId = reader.read7BitEncodedInt()
            const code = reader.readBytes()
            const data = reader.readBytes()

            const pack = {kind, route, requestId, code, data}
            packets.push(pack)
        } catch (ex) {
            // console.log(ex, stream)
            stream.seek(lastPosition, SeekOrigin.Begin)
            return packets
        }
    }
}
