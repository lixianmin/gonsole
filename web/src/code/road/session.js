'use strict'

/********************************************************************
 created:    2023-06-09
 author:     lixianmin

 Copyright (C) - All Rights Reserved
 *********************************************************************/
import {newOctetsStream, SeekOrigin} from "@src/code/iox/octets_stream";
import {newOctetsReader} from "@src/code/iox/octets_reader";
import {decode, encode} from "@src/code/road/packet_tools";
import {PacketKind} from "@src/code/road/consts";
import {newJsonSerde} from "@src/code/road/json_serde";
import {newOctetsWriter} from "@src/code/iox/octets_writer";

export function newSession() {
    const _serde = newJsonSerde()
    const _reader = newOctetsReader(newOctetsStream())
    const _writer = newOctetsWriter(newOctetsStream())

    const _kindRoutes = new Map()
    const _routeKinds = new Map()
    const _requestHandlers = new Map()

    let _onConnected = undefined
    let _socket = undefined
    let _heartbeatIntervalId = 0
    let _reconnect = undefined
    let _isVisible = true
    let _requestIdGenerator = 0

    function connect(url, onConnected) {
        _reconnect = function () {
            _reader.stream.reset()
            _writer.stream.reset()
            _kindRoutes.clear()
            _routeKinds.clear()
            _requestHandlers.clear()
            _onConnected = onConnected

            _socket = new WebSocket(url)
            _socket.binaryType = 'arraybuffer'
            _socket.onopen = onopen
            _socket.onmessage = onmessage
            _socket.onerror = onerror
            _socket.onclose = onclose
        }

        _reconnect()
    }

    function onopen(evt) {

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

    function onerror(evt) {
        stopSendingHeartbeat()
        console.error('onerror:', evt)
    }

    function onclose(evt) {
        stopSendingHeartbeat()
        if (_isVisible) {
            _reconnect()
        }

        // console.log('onclose:', evt)
    }

    document.addEventListener("visibilitychange", function () {
        _isVisible = document.visibilityState === 'visible'
        const readState = _socket.readyState
        if (_isVisible && (readState === WebSocket.CLOSING || readState === WebSocket.CLOSED)) {
            stopSendingHeartbeat()
            _reconnect()
        }
        // console.log('visibilitychange', document.visibilityState)
    });

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
            case PacketKind.RouteKind:
                onReceivedRouteKind(pack)
                break
            case PacketKind.Echo:
                onReceivedEcho(pack)
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
        handshakeRe()

        if (_onConnected) {
            _onConnected(handshake.nonce)
        }

        function buildKindRoutes() {
            _kindRoutes.clear()
            _routeKinds.clear()

            const routes = handshake.routes
            const size = routes.length
            for (let i = 0; i < size; i++) {
                const kind = PacketKind.UserBase + i
                const route = routes[i]
                _kindRoutes.set(kind, route)
                _routeKinds.set(route, kind)
            }
        }

        function startHeartbeat() {
            const interval = handshake.heartbeat * 1000 // unit: ms
            const pack = {kind: PacketKind.Heartbeat}
            _heartbeatIntervalId = setInterval(() => {
                sendPacket(pack)
            }, interval)
        }

        function handshakeRe() {
            const reply = {'serde': 'json'}
            const data = _serde.serialize(reply)
            const pack = {kind: PacketKind.HandshakeRe, data: data}
            sendPacket(pack)
        }

        console.log('handshake', handshake)
    }

    function stopSendingHeartbeat() {
        if (_heartbeatIntervalId > 0) {
            clearInterval(_heartbeatIntervalId)
            _heartbeatIntervalId = 0
        }
    }

    function onReceivedRouteKind(pack) {
        const bean = _serde.deserialize(pack.data)
        _kindRoutes.set(bean.kind, bean.route)
        _routeKinds.set(bean.route, bean.kind)
    }

    function onReceivedEcho(pack) {
        sendPacket(pack);
        console.log("pack", pack)
    }

    function onReceivedUserdata(pack) {
        if (pack.kind < PacketKind.UserBase) {
            return
        }

        const handler = fetchHandler(pack)
        if (!handler) {
            console.error(`can not find handler, kind=${pack.kind}, requestId=${pack.requestId}`)
            return
        }

        let response = undefined
        let err = undefined
        const hasError = pack.code
        if (hasError) {
            err = {
                code: _serde.bytes2String(pack.code),
                message: _serde.bytes2String(pack.data)
            }
        } else {
            response = _serde.deserialize(pack.data)
        }

        // console.log(response, err)
        handler(response, err)
    }

    function fetchHandler(pack) {
        const requestId = pack.requestId
        if (requestId !== 0) {
            const handler = _requestHandlers.get(requestId)
            _requestHandlers.delete(requestId)
            return handler
        } else {
            const route = _kindRoutes.get(pack.kind)
            return _requestHandlers.get(route)
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

    function request(route, request, handler = undefined) {
        const data = _serde.serialize(request)
        const requestId = ++_requestIdGenerator

        const kind = _routeKinds.get(route)
        const pack = {kind: kind, requestId: requestId, data: data}
        // console.log('route', route, 'kind', kind)
        // if (!kind) {
        //     const routeData = _serde.string2bytes(route)
        //     pack.kind = PacketKind.RouteBase + routeData.length
        //     pack.route = routeData
        // }

        if (handler) {
            _requestHandlers.set(requestId, handler)
        }

        sendPacket(pack)
    }

    function on(route, handler) {
        _requestHandlers.set(route, handler)
    }

    return {
        connect: connect,
        request: request,
        on: on,
    }
}