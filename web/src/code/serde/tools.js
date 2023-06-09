'use strict'

/********************************************************************
 created:    2023-06-09
 author:     lixianmin

 Copyright (C) - All Rights Reserved
 *********************************************************************/
import {SeekOrigin} from "@src/code/iox/octets_stream";

export function encode(writer, pack) {
    writer.write7BitEncodedInt(pack.kind)
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
            const code = reader.readBytes()
            const data = reader.readBytes()

            const pack = {kind, code, data}
            packets.push(pack)
        } catch (ex) {
            // console.log(ex, stream)
            stream.seek(lastPosition, SeekOrigin.Begin)
            return packets
        }
    }
}
