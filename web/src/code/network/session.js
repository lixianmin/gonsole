'use strict'

/********************************************************************
 created:    2023-06-09
 author:     lixianmin

 Copyright (C) - All Rights Reserved
 *********************************************************************/
import {newOctetsStream, SeekOrigin} from "@src/code/iox/octets_stream";
import {newOctetsReader} from "@src/code/iox/octets_reader";
import {decode, encode} from "@src/code/serde/tools";
import {PacketKind} from "@src/code/serde/consts";
import {newJsonSerde} from "@src/code/serde/json_serde";
import {newHeartbeat} from "@src/code/network/heartbeat";
import {newOctetsWriter} from "@src/code/iox/octets_writer";

export function newSession(url) {
    const socket = new WebSocket(url)
    socket.binaryType = 'arraybuffer'
    socket.onopen = onopen
    socket.onmessage = onmessage
    socket.onerror = onerror
    socket.onclose = onclose

    const serde = newJsonSerde()
    const reader = newOctetsReader(newOctetsStream())
    const writer = newOctetsWriter(newOctetsStream())

    let handshake = undefined
    const heartbeat = newHeartbeat()

    function onopen() {

    }

    function onmessage(evt) {
        const data = new Uint8Array(evt.data)
        const stream = reader.stream
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

    function send(bytes) {
        socket.send(bytes.buffer)
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
            case PacketKind.Heartbeat:
                console.log(pack)
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
        handshake = serde.deserialize(pack.data)
        buildKindRoutes()
        startHeartbeat()

        function buildKindRoutes() {
            const kind_routes = new Map()

            for (let [route, kind] of Object.entries(handshake.route_kinds)) {
                kind_routes.set(kind, route)
            }

            handshake.kind_routes = kind_routes
            delete handshake.route_kinds
        }

        function startHeartbeat() {
            const interval = handshake.heartbeat * 1000 // unit: ms
            const pack = {kind: PacketKind.Heartbeat}
            setInterval(() => {
                sendPacket(pack)
            }, interval)
        }


        heartbeat.interval = handshake.heartbeat * 1000 // unit: ms
        console.log(handshake)
    }

    function sendPacket(pack) {
        const stream = writer.stream
        stream.reset()
        encode(writer, pack)

        stream.seek(0, SeekOrigin.Begin)
        const bytes = stream.bytes
        socket.send(bytes.buffer)
    }

    return {
        sendPacket: sendPacket,
    }
}