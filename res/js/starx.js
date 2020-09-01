(function () {
    const Protocol = window.Protocol;
    const Message = Protocol.Message;
    const Package = Protocol.Package;

    const starx = {};
    window.starx = starx;

    starx.pushHandlers = {};
    starx.on = function (key, handler) {
        starx.pushHandlers[key] = handler;
    };

    const decodeIO_encoder = null;
    const decodeIO_decoder = null;

    const DEFAULT_MAX_RECONNECT_ATTEMPTS = 10;
    let reconnectionDelay;
    let reconnectAttempts;
    let reconnectUrl;
    let useCrypto;

    const processPackage = function (msgs) {
        if (Array.isArray(msgs)) {
            for (let i = 0; i < msgs.length; i++) {
                const msg = msgs[i];
                handlers[msg.type](msg.body);
            }
        } else {
            handlers[msgs.type](msgs.body);
        }
    };

    let encode;
    let handshakeCallback;
    const callbacks = {};
    let socket;
    let reconncetTimer;
    let reconnect;

    const defaultDecode = function (data) {
        const msg = Message.decode(data);

        if (msg.id > 0) {
            msg.route = routeMap[msg.id];
            delete routeMap[msg.id];
            if (!msg.route) {
                return;
            }
        }

        msg.body = decompose(msg);
        return msg;
    };

    const reset = function () {
        reconnect = false;
        reconnectionDelay = 1000 * 5;
        reconnectAttempts = 0;
        clearTimeout(reconncetTimer);
    };

    //Initialize data used in starx client
    const initData = function (data) {
        if (!data || !data.sys) {
            return;
        }

        dict = data.sys.dict;

        // init compress dict
        if (dict) {
            abbrs = {};

            for (const route in dict) {
                abbrs[dict[route]] = route;
            }
        }
    };

    const handshakeInit = function (data) {
        if (data.sys && data.sys.heartbeat) {
            heartbeatInterval = data.sys.heartbeat * 1000;   // heartbeat interval
            heartbeatTimeout = heartbeatInterval * 2;        // max heartbeat timeout
        } else {
            heartbeatInterval = 0;
            heartbeatTimeout = 0;
        }

        initData(data);

        if (typeof handshakeCallback === 'function') {
            handshakeCallback(data.user);
        }
    };

    const decompose = function (msg) {
        let route = msg.route;

        //Decompose route from dict
        if (msg.compressRoute) {
            if (!abbrs[route]) {
                return {};
            }

            route = msg.route = abbrs[route];
        }

        if (decodeIO_decoder && decodeIO_decoder.lookup(route)) {
            return decodeIO_decoder.build(route).decode(msg.body);
        } else {
            return JSON.parse(Protocol.strdecode(msg.body));
        }

        return msg;
    };

    const processMessage = function (msg) {
        if (!msg.id) { // server push message
            const handler = starx.pushHandlers[msg.route];
            if (handler != null) {
                handler(msg.body);
            }
            return;
        }

        //if have a id then find the callback function with the request
        const cb = callbacks[msg.id];
        delete callbacks[msg.id];

        if (typeof cb !== 'function') {
            return;
        }

        cb(msg.body);
    };

    const heartbeatTimeoutCb = function () {
        const gap = nextHeartbeatTimeout - Date.now();
        if (gap > gapThreshold) {
            heartbeatTimeoutId = setTimeout(heartbeatTimeoutCb, gap);
        } else {
            console.error('server heartbeat timeout');
            starx.emit('heartbeat timeout');
            starx.disconnect();
        }
    };

    const send = function (packet) {
        socket.send(packet.buffer);
    };

    const sendMessage = function (reqId, route, msg) {
        if (useCrypto) {
            msg = JSON.stringify(msg);
            var sig = rsa.signString(msg, "sha256");
            msg = JSON.parse(msg);
            msg['__crypto__'] = sig;
        }

        if (encode) {
            msg = encode(reqId, route, msg);
        }

        const packet = Package.encode(Package.TYPE_DATA, msg);
        send(packet);
    };

    const connect = function (params, url, cb) {
        console.log('connect to: ' + url);

        params = params || {};
        const maxReconnectAttempts = params.maxReconnectAttempts || DEFAULT_MAX_RECONNECT_ATTEMPTS;
        reconnectUrl = url;

        const onopen = function (event) {
            if (!!reconnect) {
                starx.emit('reconnect');
            }
            reset();
            const obj = Package.encode(Package.TYPE_HANDSHAKE, Protocol.strencode(JSON.stringify(handshakeBuffer)));
            send(obj);
        };

        const onmessage = function (event) {
            processPackage(Package.decode(event.data), cb);

            // new package arrived, update the heartbeat timeout
            if (heartbeatTimeout) {
                nextHeartbeatTimeout = Date.now() + heartbeatTimeout;
            }
        };

        const onerror = function (event) {
            starx.emit('io-error', event);
            console.error('socket error: ', event);
        };

        const onclose = function (event) {
            starx.emit('close', event);
            starx.emit('disconnect', event);
            console.log('socket close: ', event);
            if (!!params.reconnect && reconnectAttempts < maxReconnectAttempts) {
                reconnect = true;
                reconnectAttempts++;
                reconncetTimer = setTimeout(function () {
                    connect(params, reconnectUrl, cb);
                }, reconnectionDelay);
                reconnectionDelay *= 2;
            }
        };

        socket = new WebSocket(url);
        socket.binaryType = 'arraybuffer';
        socket.onopen = onopen;
        socket.onmessage = onmessage;
        socket.onerror = onerror;
        socket.onclose = onclose;
    };

    const defaultEncode = function (reqId, route, msg) {
        const type = reqId ? Message.TYPE_REQUEST : Message.TYPE_NOTIFY;

        if (decodeIO_encoder && decodeIO_encoder.lookup(route)) {
            var Builder = decodeIO_encoder.build(route);
            msg = new Builder(msg).encodeNB();
        } else {
            msg = Protocol.strencode(JSON.stringify(msg));
        }

        var compressRoute = 0;
        if (dict && dict[route]) {
            route = dict[route];
            compressRoute = 1;
        }

        return Message.encode(reqId, type, compressRoute, route, msg);
    };

    const JS_WS_CLIENT_TYPE = 'js-websocket';
    const JS_WS_CLIENT_VERSION = '0.0.1';
    var rsa = window.rsa;

    if (typeof (window) != "undefined" && typeof (sys) != 'undefined' && sys.localStorage) {
        window.localStorage = sys.localStorage;
    }

    const RES_OK = 200;
    const RES_OLD_CLIENT = 501;

    socket = null;
    let reqId = 0;
    const handlers = {};
    //Map from request id to route
    var routeMap = {};
    var dict = {};    // route string to code
    var abbrs = {};   // code to route string

    var heartbeatInterval = 0;
    var heartbeatTimeout = 0;
    var nextHeartbeatTimeout = 0;
    var gapThreshold = 100;   // heartbeat gap threashold
    var heartbeatId = null;
    var heartbeatTimeoutId = null;
    handshakeCallback = null;

    let decode = null;
    encode = null;

    reconnect = false;
    reconncetTimer = null;
    reconnectUrl = null;
    reconnectAttempts = 0;
    reconnectionDelay = 5000;

    var handshakeBuffer = {
        'sys': {
            type: JS_WS_CLIENT_TYPE,
            version: JS_WS_CLIENT_VERSION,
            rsa: {}
        },
        'user': {}
    };

    var initCallback = null;

    starx.encode = defaultEncode;
    starx.decode = defaultDecode;

    starx.init = function (params, cb) {
        initCallback = cb;

        encode = params.encode || defaultEncode;
        decode = params.decode || defaultDecode;

        handshakeBuffer.user = params.user;
        if (params.encrypt) {
            useCrypto = true;
            rsa.generate(1024, "10001");
            handshakeBuffer.sys.rsa = {
                rsa_n: rsa.n.toString(16),
                rsa_e: rsa.e
            };
        }

        handshakeCallback = params.handshakeCallback;
        connect(params, params.url, cb);
    };

    starx.disconnect = function () {
        if (socket) {
            if (socket.disconnect) socket.disconnect();
            if (socket.close) socket.close();
            console.log('disconnect');
            socket = null;
        }

        if (heartbeatId) {
            clearTimeout(heartbeatId);
            heartbeatId = null;
        }
        if (heartbeatTimeoutId) {
            clearTimeout(heartbeatTimeoutId);
            heartbeatTimeoutId = null;
        }
    };

    starx.request = function (route, msg, cb) {
        if (arguments.length === 2 && typeof msg === 'function') {
            cb = msg;
            msg = {};
        } else {
            msg = msg || {};
        }
        route = route || msg.route;
        if (!route) {
            return;
        }

        reqId++;
        sendMessage(reqId, route, msg);

        callbacks[reqId] = cb;
        routeMap[reqId] = route;
    };

    starx.notify = function (route, msg) {
        msg = msg || {};
        sendMessage(0, route, msg);
    };

    handlers[Package.TYPE_HEARTBEAT] = function (data) {
        if (!heartbeatInterval) {
            // no heartbeat
            return;
        }

        var obj = Package.encode(Package.TYPE_HEARTBEAT);
        if (heartbeatTimeoutId) {
            clearTimeout(heartbeatTimeoutId);
            heartbeatTimeoutId = null;
        }

        if (heartbeatId) {
            // already in a heartbeat interval
            return;
        }

        heartbeatId = setTimeout(function () {
            heartbeatId = null;
            send(obj);

            nextHeartbeatTimeout = Date.now() + heartbeatTimeout;
            heartbeatTimeoutId = setTimeout(heartbeatTimeoutCb, heartbeatTimeout);
        }, heartbeatInterval);
    };

    handlers[Package.TYPE_HANDSHAKE] = function (data) {
        data = JSON.parse(Protocol.strdecode(data));
        if (data.code === RES_OLD_CLIENT) {
            starx.emit('error', 'client version not fullfill');
            return;
        }

        if (data.code !== RES_OK) {
            starx.emit('error', 'handshake fail');
            return;
        }

        handshakeInit(data);

        const obj = Package.encode(Package.TYPE_HANDSHAKE_ACK);
        send(obj);

        if (initCallback) {
            initCallback(socket);
        }
    };

    handlers[Package.TYPE_DATA] = function (data) {
        let msg = data;
        if (decode) {
            msg = decode(msg);
        }

        processMessage(msg);
    };

    handlers[Package.TYPE_KICK] = function (data) {
        data = JSON.parse(Protocol.strdecode(data));
        starx.emit('onKick', data);
    };
})();
