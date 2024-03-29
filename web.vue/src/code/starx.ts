/********************************************************************
 created:    2022-01-10
 author:     lixianmin

 Copyright (C) - All Rights Reserved
 *********************************************************************/
import {Packet} from "./packet"
import {PacketType} from "./packet_type"
import {strDecode, strEncode} from "./protocol"
import {Message} from "./message"
import {OctetsStream} from "./core/octets_stream"
import {MessageType} from "./message_type"
import {createHeartbeat} from "./heartbeat.js"

type PushHandlerFunc = (data: any) => void
type HandlerFunc = (data: Uint8Array) => void

export class StartX {
    public on(key: string, handler: PushHandlerFunc) {
        this.pushHandlers.set(key, handler)
    }

    public emit(key: string, args: any = '') {
        const handler = this.pushHandlers.get(key)
        if (handler != null) {
            handler(args)
        }
    }

    private processPackages(packages: Packet[]) {
        for (const pack of packages) {
            const handler = this.handlers.get(pack.type)
            if (handler != null) {
                handler(pack.body)
            }
        }
    }

    private defaultDecode(data) {
        const msg = Message.decode(data)

        if (msg.id > 0) {
            msg.route = this.routeMap.get(msg.id)
            this.routeMap.delete(msg.id)

            if (!msg.route) {
                return
            }
        }

        msg.body = this.decompose(msg)
        return msg
    }

    private decompose(msg: Message) {
        let route = msg.route

        //Decompose route from dict
        if (msg.compressRoute) {
            if (!this.abbrs[route]) {
                return {}
            }

            route = msg.route = this.abbrs[route]
        }

        return JSON.parse(strDecode(msg.body))
    }

    private reset() {
        this.reconnect = false
        this.reconnectionDelay = 1000 * 5
        this.reconnectAttempts = 0
        clearTimeout(this.reconnectTimer)
    }

    private initData(data) {
        if (!data || !data.sys) {
            return
        }

        this.dict = data.sys.dict

        // init compress dict
        if (this.dict) {
            this.abbrs = {}

            for (const route in this.dict) {
                this.abbrs[this.dict[route]] = route
            }
        }
    }

    private handshakeInit(data) {
        if (data.sys && data.sys.heartbeat) {
            this.heartbeat.interval = data.sys.heartbeat * 1000
        }

        this.initData(data)

        if (typeof this.handshakeCallback === 'function') {
            this.handshakeCallback(data.user)
        }
    }

    private processMessage(msg: Message) {
        if (msg.id) {
            // if there is an id, then find the callback function with the request
            const callback = this.callbacks.get(msg.id)
            this.callbacks.delete(msg.id)

            if (typeof callback === 'function') {
                callback(msg.body)
            }
        } else { // server push message
            const handler = this.pushHandlers.get(msg.route)
            if (typeof handler !== "undefined") {
                handler(msg.body)
            } else {
                console.log(`cannot find handler for route=${msg.route}, msg=`, msg)
            }
        }
    }

    private send(packet: Uint8Array) {
        if (this.socket != null) {
            // console.trace("send:", packet)
            this.socket.send(packet.buffer)
        } else {
            console.log("socket = null")
        }
    }

    private sendMessage(requestId: number, route, message: any) {
        // if (this.useCrypto) {
        //     message = JSON.stringify(message);
        //     var sig = window.rsa.signString(message, "sha256");
        //     message = JSON.parse(message);
        //     message['__crypto__'] = sig;
        // }

        let body = message
        if (this.encode) {
            body = this.encode(requestId, route, message)
        }

        const packet = Packet.encode(PacketType.Data, body)
        this.send(packet)
    }

    private connectInner(params, url: string) {
        console.log('connect to: ' + url)
        params = params || {}

        const DEFAULT_MAX_RECONNECT_ATTEMPTS = 10
        const maxReconnectAttempts = params.maxReconnectAttempts || DEFAULT_MAX_RECONNECT_ATTEMPTS;
        this.reconnectUrl = url

        const onopen = (event: Event) => {
            // console.log("onopen", event)
            if (this.reconnect) {
                this.emit('reconnect');
            }

            this.reset()
            // client主动handshake，把自己的参数告诉server，然后server会发送HandshakeAck发送heartbeatInterval等参数
            const packet = Packet.encode(PacketType.Handshake, strEncode(JSON.stringify(this.handshakeData)))
            this.send(packet)
        }

        const onmessage = (event: MessageEvent<ArrayBuffer>) => {
            const data = new Uint8Array(event.data)
            const stream = this.buffer

            stream.write(data, 0, data.length)
            stream.setPosition(0)

            const packets = Packet.decode(stream)
            this.processPackages(packets)
        }

        const onerror = (event: Event) => {
            this.emit('io-error', event)
            console.error('socket error: ', event)
        };

        const onclose = (event: CloseEvent) => {
            this.emit('close', event)
            this.emit('disconnect', event)
            console.log('socket close: ', event)

            if (params.reconnect && this.reconnectAttempts < maxReconnectAttempts) {
                this.reconnect = true
                this.reconnectAttempts++

                this.reconnectTimer = setTimeout(() => {
                    this.connectInner(params, this.reconnectUrl)
                }, this.reconnectionDelay)
                this.reconnectionDelay *= 2
            }
        };

        let socket = new WebSocket(url)
        socket.binaryType = 'arraybuffer'
        socket.onopen = onopen
        socket.onmessage = onmessage
        socket.onerror = onerror
        socket.onclose = onclose

        this.socket = socket
    }

    public connect(params: any, onConnected) {
        this.handshakeCallback = params.handshakeCallback
        this.onConnected = onConnected

        this.encode = params.encode || this.defaultEncode
        this.decode = params.decode || this.defaultDecode

        this.handshakeData.user = params.user;
        // if (params.encrypt) {
        //     this.useCrypto = true;
        //     rsa.generate(1024, "10001");
        //     this.handshakeData.sys.rsa = {
        //         rsa_n: rsa.n.toString(16),
        //         rsa_e: rsa.e
        //     };
        // }

        this.handlers.set(PacketType.Handshake, this.handleHandshake)  // 这是服务器推过来的，用于传递一些heartbeat interval之类的参数给client
        this.handlers.set(PacketType.Heartbeat, this.handleHeartbeat)
        this.handlers.set(PacketType.Data, this.handleData)
        this.handlers.set(PacketType.Kick, this.handleKick)
        this.connectInner(params, params.url)
    }

    private defaultEncode(requestId: number, route, message) {
        const type = requestId != 0 ? MessageType.Request : MessageType.Notify

        message = strEncode(JSON.stringify(message))

        let compressRoute = false
        if (this.dict && this.dict[route]) {
            route = this.dict[route]
            compressRoute = true
        }

        return Message.encode(requestId, type, compressRoute, route, message)
    }

    public disconnect() {
        if (this.socket != null) {
            this.socket.close()
            console.log('disconnect')
            this.socket = null
        }

        this.heartbeat.clearTimeout()
    }

    public request(route: string, message, callback) {
        // requestId不能是0, 否则会被认为是notify类型, 而不是request
        let requestId = ++this.requestIdGenerator
        this.sendMessage(requestId, route, message)

        this.callbacks.set(requestId, callback)
        this.routeMap.set(requestId, route)
    }

    public notify(route, message) {
        message = message || {}
        this.sendMessage(0, route, message)
    }

    private handleHandshake = (data: Uint8Array) => {
        let item = JSON.parse(strDecode(data))

        const RES_OLD_CLIENT = 501
        if (item.code === RES_OLD_CLIENT) {
            this.emit('error', 'client version not fulfill')
            return;
        }

        const RES_OK = 200
        if (item.code !== RES_OK) {
            this.emit('error', 'handshake fail');
            return
        }

        this.handshakeInit(item)

        const packet = Packet.encode(PacketType.HandshakeAck)
        this.send(packet)

        // handshakeAck之后，目前服务器会回复heartbeat，然后 就可以自动登录了
        const onConnected = this.onConnected
        if (onConnected != null) {
            const nonce = item.nonce
            onConnected(nonce)
        }
    }

    // 通过 => 定义 function, 使它可以在定义的时候捕获this, 而不是在使用的时候
    // https://www.typescriptlang.org/docs/handbook/functions.html#this-and-arrow-functions
    private handleHeartbeat = (data) => {
        setTimeout(() => {
            const packet = Packet.encode(PacketType.Heartbeat)
            this.send(packet)
        }, this.heartbeat.interval)

        this.resetHeartbeatTimeout()
    }

    private resetHeartbeatTimeout = () => {
        this.heartbeat.clearTimeout()
        this.heartbeat.setTimeout(() => {
            console.error('server heartbeat timeout')
            this.emit('heartbeat timeout')
            this.disconnect()
        })
    }

    private handleData = (data) => {
        let msg = data
        if (this.decode) {
            msg = this.decode(msg)
        }

        this.processMessage(msg)
        // this.resetHeartbeatTimeout()
    }

    private handleKick = (data: Uint8Array) => {
        data = JSON.parse(strDecode(data))
        this.emit('onKick', data)
    }

    private socket: WebSocket | null = null
    private buffer = new OctetsStream(8)
    private useCrypto = false
    private encode
    private decode
    private requestIdGenerator = 0

    private reconnectUrl = ""
    private reconnect = false
    private reconnectTimer: any
    private reconnectAttempts = 0
    private reconnectionDelay = 5000

    private handshakeData = {
        'sys': {
            type: 'js-websocket',
            version: '0.0.1',
            rsa: {}
        },
        'user': {}
    };

    private pushHandlers = new Map<string, PushHandlerFunc>()
    private handlers = new Map<number, HandlerFunc>()
    private routeMap = new Map<number, string>()
    private callbacks = new Map<number, any>()

    private abbrs = {}
    private dict = {}

    private heartbeat = createHeartbeat()
    private handshakeCallback
    private onConnected
}