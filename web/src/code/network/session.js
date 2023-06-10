'use strict'

/********************************************************************
 created:    2023-06-09
 author:     lixianmin

 Copyright (C) - All Rights Reserved
 *********************************************************************/
import {newOctetsStream, SeekOrigin} from "@src/code/iox/octets_stream";
import {newOctetsReader} from "@src/code/iox/octets_reader";
import {decode, encode} from "@src/code/network/tools";
import {PacketKind} from "@src/code/network/consts";
import {newJsonSerde} from "@src/code/network/json_serde";
import {newOctetsWriter} from "@src/code/iox/octets_writer";

export function newSession() {
    const _serde = newJsonSerde()
    const _reader = newOctetsReader(newOctetsStream())
    const _writer = newOctetsWriter(newOctetsStream())

    const _kindRoutes = new Map()
    const _routeKinds = new Map()
    const _routeHandlers = new Map()

    let _onConnected = undefined
    let _socket = undefined

    function connect(url, onConnected) {
        _onConnected = onConnected
        _socket = new WebSocket(url)
        _socket.binaryType = 'arraybuffer'
        _socket.onopen = onopen
        _socket.onmessage = onmessage
        _socket.onerror = onerror
        _socket.onclose = onclose
    }

    function onopen() {

    }

    function onmessage(evt) {
        const data = new Uint8Array(evt.data)
        const stream = _reader.stream
        stream.write(data, 0, data.length)
        stream.seek(0, SeekOrigin.Begin)
        // console.log(`onmessage: data.length=${data.length}, stream=`, stream)

        onReceivedData(_reader)
        stream.tidy()
    }

    function onerror() {

    }

    function onclose() {

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
                // console.log(pack)
                break
            case PacketKind.Kick:
                _socket.close(0, 'kicked')
                break
            default:
                onReceivedUserdata(pack)
                break
        }
    }

    function onReceivedHandshake(pack) {
        const handshake = _serde.deserialize(pack.data)
        buildKindRoutes()
        startHeartbeat()

        if (_onConnected) {
            _onConnected(handshake.nonce)
        }

        function buildKindRoutes() {
            _kindRoutes.clear()
            _routeKinds.clear()
            for (let [route, kind] of Object.entries(handshake.route_kinds)) {
                _kindRoutes.set(kind, route)
                _routeKinds.set(route, kind)
            }
        }

        function startHeartbeat() {
            const interval = handshake.heartbeat * 1000 // unit: ms
            const pack = {kind: PacketKind.Heartbeat}
            setInterval(() => {
                sendPacket(pack)
            }, interval)
        }

        console.log('handshake', handshake)
    }

    function onReceivedUserdata(pack) {
        if (pack.kind >= PacketKind.Userdata) {
            const route = _kindRoutes.get(pack.kind)
            const handler = _routeHandlers.get(route)
            if (handler) {
                const item = _serde.deserialize(pack.data)
                let response = undefined
                let err = undefined
                if (pack.code) {
                    err = {
                        code: _serde.deserialize(pack.code),
                        message: item
                    }
                } else {
                    response = item
                }

                // console.log(response, err)
                handler(response, err)
            }
        }
    }

    function sendPacket(pack) {
        const stream = _writer.stream
        stream.reset()
        encode(_writer, pack)

        stream.seek(0, SeekOrigin.Begin)
        const bytes = stream.bytes
        _socket.send(bytes.buffer)
    }

    function request(route, bean, callback = undefined) {
        const kind = _routeKinds.get(route)
        if (kind) {
            const data = _serde.serialize(bean)
            const pack = {kind: kind, data: data}
            sendPacket(pack)

            if (callback) {
                _routeHandlers.set(route, callback)
            }
        }
    }

    function on(route, handler) {
        _routeHandlers.set(route, handler)
    }

    return {
        connect: connect,
        request: request,
        on: on,
    }
}