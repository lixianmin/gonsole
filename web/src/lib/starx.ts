/********************************************************************
 created:    2022-01-10
 author:     lixianmin

 Copyright (C) - All Rights Reserved
 *********************************************************************/
import {Packet} from "./packet";
import {PacketType} from "./packet_type";
import {strdecode, strencode} from "./protocol";
import {Message} from "./message";
import {OctetsStream, SeekOrigin} from "./octets_stream";
import {MessageType} from "./message_type";

type PushHandlerFunc = (data: any) => void
type HandlerFunc = (data: string) => void

export default class StartX {
    public on(key: string, handler: PushHandlerFunc) {
        this.pushHandlers[key] = handler;
    }

    public emit(key: string, args: any = '') {
        const handler = this.pushHandlers[key] as PushHandlerFunc
        if (handler != null) {
            handler(args)
        }
    }

    private processPackages(packages: any) {
        for (let i = 0; i < packages.length; i++) {
            const pack = packages[i];
            const handler = this.handlers[pack.type] as HandlerFunc
            if (handler != null) {
                handler(pack.body)
            }
        }
    }

    private defaultDecode(data) {
        const msg = Message.decode(data)

        if (msg.id > 0) {
            msg.route = this.routeMap[msg.id]
            this.routeMap.delete(msg.id)

            if (!msg.route) {
                return;
            }
        }

        msg.body = this.decompose(msg);
        return msg;
    }

    private decompose(msg: Message) {
        let route = msg.route;

        //Decompose route from dict
        if (msg.compressRoute) {
            if (!this.abbrs[route]) {
                return {}
            }

            route = msg.route = this.abbrs[route]
        }

        return JSON.parse(strdecode(msg.body))
    }

    private reset() {
        this.reconnect = false;
        this.reconnectionDelay = 1000 * 5;
        this.reconnectAttempts = 0;
        clearTimeout(this.reconnectTimer);
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
            this.heartbeatInterval = data.sys.heartbeat * 1000;     // heartbeat interval
            this.heartbeatTimeout = this.heartbeatInterval * 2;     // max heartbeat timeout
        } else {
            this.heartbeatInterval = 0
            this.heartbeatTimeout = 0
        }

        this.initData(data)

        if (typeof this.handshakeCallback === 'function') {
            this.handshakeCallback(data.user)
        }
    }

    private processMessage(msg: Message) {
        // todo 这些需要测试才行
        if (msg.id) {
            // if there is an id, then find the callback function with the request
            const callback = this.callbacks[msg.id]
            this.callbacks.delete(msg.id)

            if (typeof callback === 'function') {
                callback(msg.body)
            }
        } else { // server push message
            const handler = this.pushHandlers[msg.route] as PushHandlerFunc
            if (typeof handler !== "undefined") {
                handler(msg.body)
            }
        }
    }

    private heartbeatTimeoutCb() {
        const gap = this.nextHeartbeatTimeout - Date.now();
        const gapThreshold = 100;   // heartbeat gap threshold
        if (gap > gapThreshold) {
            this.heartbeatTimeoutId = setTimeout(this.heartbeatTimeoutCb, gap);
        } else {
            console.error('server heartbeat timeout');
            this.emit('heartbeat timeout');
            this.disconnect();
        }
    }

    private send(packet: Uint8Array) {
        if (this.socket != null) {
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

    public connect(params, url: string, callback) {
        console.log('connect to: ' + url)
        params = params || {}

        const DEFAULT_MAX_RECONNECT_ATTEMPTS = 10
        const maxReconnectAttempts = params.maxReconnectAttempts || DEFAULT_MAX_RECONNECT_ATTEMPTS;
        this.reconnectUrl = url

        const onopen = (event) => {
            console.log("onopen", event)
            if (this.reconnect) {
                this.emit('reconnect');
            }

            this.reset()
            const packet = Packet.encode(PacketType.Handshake, strencode(JSON.stringify(this.handshakeBuffer)));
            this.send(packet)
        }

        const onmessage = (event: MessageEvent) => {
            let data = new Uint8Array(event.data)
            let stream = this.buffer

            stream.write(data, 0, data.length)
            stream.setPosition(0)
            this.processPackages(Packet.decode(stream))

            // new package arrived, update the heartbeat timeout
            if (this.heartbeatTimeout) {
                this.nextHeartbeatTimeout = Date.now() + this.heartbeatTimeout
            }
        }

        const onerror = (event) => {
            this.emit('io-error', event)
            console.error('socket error: ', event)
        };

        const onclose = (event) => {
            this.emit('close', event)
            this.emit('disconnect', event)
            console.log('socket close: ', event)

            if (params.reconnect && this.reconnectAttempts < maxReconnectAttempts) {
                this.reconnect = true
                this.reconnectAttempts++

                this.reconnectTimer = setTimeout(() => {
                    this.connect(params, this.reconnectUrl, callback)
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

    public init(params, callback) {
        this.initCallback = callback
        this.handshakeCallback = params.handshakeCallback

        this.encode = params.encode || this.defaultEncode
        this.decode = params.decode || this.defaultDecode

        this.handshakeBuffer.user = params.user;
        // if (params.encrypt) {
        //     this.useCrypto = true;
        //     rsa.generate(1024, "10001");
        //     this.handshakeBuffer.sys.rsa = {
        //         rsa_n: rsa.n.toString(16),
        //         rsa_e: rsa.e
        //     };
        // }

        this.handlers[PacketType.Heartbeat] = this.handleHeartBeat
        this.handlers[PacketType.Handshake] = this.handleHandshake
        this.handlers[PacketType.Data] = this.handleData
        this.handlers[PacketType.Kick] = this.handleKick
        this.connect(params, params.url, callback)
    }

    private defaultEncode(requestId: number, route, message) {
        const type = requestId != 0 ? MessageType.Request : MessageType.Notify

        message = strencode(JSON.stringify(message))

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

        if (this.heartbeatId) {
            clearTimeout(this.heartbeatId)
            this.heartbeatId = null
        }

        if (this.heartbeatTimeoutId) {
            clearTimeout(this.heartbeatTimeoutId)
            this.heartbeatTimeoutId = null
        }
    }

    public request(route: string, message, callback) {
        let requestId = this.requestIdGenerator++
        this.sendMessage(requestId, route, message)

        this.callbacks[requestId] = callback
        this.routeMap[requestId] = route
    }

    public notify(route, message) {
        message = message || {}
        this.sendMessage(0, route, message)
    }

    // 通过 => 定义 function, 使它可以在定义的时候捕获this, 而不是在使用的时候
    // https://www.typescriptlang.org/docs/handbook/functions.html#this-and-arrow-functions
    private handleHeartBeat = (data: Uint8Array) => {
        if (!this.heartbeatInterval) {
            // no heartbeat
            return;
        }

        if (this.heartbeatTimeoutId) {
            clearTimeout(this.heartbeatTimeoutId)
            this.heartbeatTimeoutId = null
        }

        if (this.heartbeatId) {
            // already in a heartbeat interval
            return;
        }

        this.heartbeatId = setTimeout(() => {
            this.heartbeatId = null
            const packet = Packet.encode(PacketType.Heartbeat)
            this.send(packet)

            this.nextHeartbeatTimeout = Date.now() + this.heartbeatTimeout
            this.heartbeatTimeoutId = setTimeout(this.heartbeatTimeoutCb, this.heartbeatTimeout)
        }, this.heartbeatInterval)
    }

    private handleHandshake = (data) => {
        let item = JSON.parse(strdecode(data))

        const RES_OLD_CLIENT = 501
        if (item.code === RES_OLD_CLIENT) {
            this.emit('error', 'client version not fullfill')
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

        if (this.initCallback) {
            this.initCallback(this.socket)
        }
    }

    private handleData = (data) => {
        let msg = data
        if (this.decode) {
            msg = this.decode(msg)
        }

        this.processMessage(msg)
    }

    private handleKick = (data) => {
        data = JSON.parse(strdecode(data))
        this.emit('onKick', data)
    }

    private socket: WebSocket | null = null
    private buffer = new OctetsStream(8)
    private useCrypto = false
    private encode
    private decode
    private initCallback
    private requestIdGenerator = 0

    private reconnectUrl = ""
    private reconnect = false
    private reconnectTimer: any
    private reconnectAttempts = 0
    private reconnectionDelay = 5000

    private handshakeBuffer = {
        'sys': {
            type: 'js-websocket',
            version: '0.0.1',
            rsa: {}
        },
        'user': {}
    };

    private pushHandlers = new Map<string, PushHandlerFunc>()
    private handlers = new Map<number, HandlerFunc>()
    private routeMap = new Map<number, any>()
    private callbacks = new Map<number, any>()

    private abbrs = {}
    private dict = {}

    private heartbeatInterval = 0
    private heartbeatTimeout = 0
    private nextHeartbeatTimeout = 0
    private heartbeatTimeoutId: any
    private heartbeatId: any
    private handshakeCallback
}