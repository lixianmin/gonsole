/********************************************************************
 created:    2022-01-10
 author:     lixianmin

 Copyright (C) - All Rights Reserved
 *********************************************************************/
import {strDecode, strEncode} from "./protocol";
import {Buffers} from "./core/buffers";
import {MessageType} from "./message_type";

export class Message {
    static MSG_FLAG_BYTES = 1
    static MSG_ROUTE_CODE_BYTES = 2
    static MSG_ID_MAX_BYTES = 5
    static MSG_ROUTE_LEN_BYTES = 1

    static MSG_ROUTE_CODE_MAX = 0xffff
    static MSG_COMPRESS_ROUTE_MASK = 0x1
    static MSG_TYPE_MASK = 0x7

    /**
     * Message protocol encode.
     *
     * @param  {Number} id            message id
     * @param  {Number} type          message type
     * @param  {Number} compressRoute whether compress route
     * @param  {Number|String} route  route code or route string
     * @param  {Buffer} msg           message body bytes
     * @return {Buffer}               encode result
     */
    public static encode(id: number, type: MessageType, compressRoute: boolean, route: any, msg: Uint8Array): Uint8Array {
        // calculate message max length
        const idBytes = Message.hasId(type) ? Message.calculateMsgIdBytes(id) : 0;
        let msgLen = Message.MSG_FLAG_BYTES + idBytes;

        if (Message.hasRoute(type)) {
            if (compressRoute) {
                if (typeof route !== 'number') {
                    throw new Error('error flag for number route!')
                }
                msgLen += Message.MSG_ROUTE_CODE_BYTES
            } else {
                msgLen += Message.MSG_ROUTE_LEN_BYTES
                if (route) {
                    route = strEncode(route)
                    if (route.length > 255) {
                        throw new Error('route maxlength is overflow')
                    }
                    msgLen += route.length
                }
            }
        }

        if (msg) {
            msgLen += msg.length
        }

        const buffer = new Uint8Array(msgLen)
        let offset = 0

        // add flag
        offset = Message.encodeMsgFlag(type, compressRoute, buffer, offset)

        // add message id
        if (Message.hasId(type)) {
            offset = Message.encodeMsgId(id, buffer, offset)
        }

        // add route
        if (Message.hasRoute(type)) {
            offset = Message.encodeMsgRoute(compressRoute, route, buffer, offset)
        }

        // add body
        if (msg != null) {
            offset = Message.encodeMsgBody(msg, buffer, offset)
        }

        return buffer
    }

    /**
     * Message protocol decode.
     *
     * @param  {Buffer|Uint8Array} buffer message bytes
     * @return {Object}            message object
     */
    static decode(buffer): Message {
        const bytes = new Uint8Array(buffer)
        const bytesLen = bytes.length || bytes.byteLength
        let offset = 0
        let id = 0
        let route: string = ''

        // parse flag
        const flag = bytes[offset++];
        const compressRoute = flag & Message.MSG_COMPRESS_ROUTE_MASK
        const type = (flag >> 1) & Message.MSG_TYPE_MASK

        // parse id
        if (Message.hasId(type)) {
            let m = (bytes[offset])
            let i = 0
            do {
                m = (bytes[offset])
                id = id + ((m & 0x7f) * Math.pow(2, (7 * i)))
                offset++
                i++
            } while (m >= 128)
        }

        // parse route
        if (Message.hasRoute(type)) {
            if (compressRoute != 0) {
                route = ((bytes[offset++]) << 8 | bytes[offset++]).toString()
            } else {
                const routeLen = bytes[offset++]
                if (routeLen > 0) {
                    let buf = new Uint8Array(routeLen)
                    Buffers.blockCopy(bytes, offset, buf, 0, routeLen)
                    route = strDecode(buf)
                    // console.log("type=", type, ", compressRoute=", compressRoute, ", routeLen=", routeLen, ", route=", route)
                } else {
                    route = ''
                }
                offset += routeLen
            }
        }

        // parse body
        const bodyLen = bytesLen - offset;
        const body = new Uint8Array(bodyLen)
        Buffers.blockCopy(bytes, offset, body, 0, bodyLen)

        let result = new Message(id, type, compressRoute, route, body)
        return result
    }

    private static calculateMsgIdBytes(id: number) {
        let len = 0
        do {
            len += 1
            id >>= 7
        } while (id > 0)

        return len
    }

    private static encodeMsgFlag(type: MessageType, compressRoute: boolean, buffer, offset) {
        if (!Message.isValid(type)) {
            throw new Error('unknown message type: ' + type)
        }

        buffer[offset] = (type << 1) | (compressRoute ? 1 : 0)
        return offset + Message.MSG_FLAG_BYTES
    }

    private static encodeMsgId(id: number, buffer: Uint8Array, offset: number): number {
        do {
            let tmp = id % 128
            const next = Math.floor(id / 128)

            if (next !== 0) {
                tmp = tmp + 128
            }
            buffer[offset++] = tmp

            id = next;
        } while (id !== 0)

        return offset
    }

    private static encodeMsgRoute(compressRoute, route, buffer, offset) {
        // console.trace(`compressRoute=${compressRoute}, route=${route}, buffer=${buffer}, offset=${offset}`)
        if (compressRoute) {
            if (route > Message.MSG_ROUTE_CODE_MAX) {
                throw new Error('route number is overflow')
            }

            buffer[offset++] = (route >> 8) & 0xff
            buffer[offset++] = route & 0xff
        } else {
            if (route) {
                buffer[offset++] = route.length & 0xff
                Buffers.blockCopy(route, 0, buffer, offset, route.length)
                offset += route.length
            } else {
                buffer[offset++] = 0
            }
        }

        return offset
    }

    private static encodeMsgBody(msg: Uint8Array, buffer: Uint8Array, offset: number): number {
        Buffers.blockCopy(msg, 0, buffer, offset, msg.length)
        return offset + msg.length
    }

    private static hasId(type: MessageType): boolean {
        return type === MessageType.Request || type === MessageType.Response
    }

    private static hasRoute(type: MessageType): boolean {
        return type === MessageType.Request || type === MessageType.Notify || type === MessageType.Push
    }

    private static isValid(type: MessageType): boolean {
        return type >= MessageType.Request && type < MessageType.Count
    }

    private constructor(id: number, type: number, compressRoute: number, route: string, body: Uint8Array) {
        this.id = id
        this.type = type
        this.compressRoute = compressRoute
        this.route = route
        this.body = body
    }

    public id
    public type
    public compressRoute
    public route
    public body: Uint8Array
}