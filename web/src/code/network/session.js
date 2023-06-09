/********************************************************************
 created:    2023-06-09
 author:     lixianmin

 Copyright (C) - All Rights Reserved
 *********************************************************************/
import {newOctetsStream, SeekOrigin} from "@src/code/iox/octets_stream";
import {newOctetsReader} from "@src/code/iox/octets_reader";
import {decode} from "@src/code/serde/tools";
import {PacketKind} from "@src/code/serde/consts";
import {newJsonSerde} from "@src/code/serde/json_serde";

export function newSession(url) {
    const serde = newJsonSerde()
    const socket = new WebSocket(url)
    socket.binaryType = 'arraybuffer'
    socket.onopen = onopen
    socket.onmessage = onmessage
    socket.onerror = onerror
    socket.onclose = onclose

    const stream = newOctetsStream()
    const reader = newOctetsReader(stream)

    function onopen() {

    }

    function onmessage(evt) {
        const data = new Uint8Array(evt.data)
        stream.write(data, 0, data.length)
        stream.seek(0, SeekOrigin.Begin)
        // console.log(`onmessage: data.length=${data.length}, stream=`, stream)

        onReceivedData(reader)
        stream.tidy()
    }

    function onerror() {

    }

    function onclose() {

    }

    function send(data) {
        socket.send(data.buffer)
    }

    function onReceivedData(reader) {
        const packets = decode(reader)
        // console.log(packets)
        for (const pack of packets) {
            onReceivedPacket(pack)
        }
    }

    function onReceivedPacket(pack) {
        switch (pack.kind) {
            case PacketKind.Handshake:
                onReceivedHandshake(pack)
                break
            case PacketKind.Kick:
                break
            default:
                if (pack.kind >= PacketKind.UserDefined) {

                }
                break
        }
    }

    function onReceivedHandshake(pack) {
        const handshake = serde.deserialize(pack.data)
        console.log(handshake)
    }

    return {}
}