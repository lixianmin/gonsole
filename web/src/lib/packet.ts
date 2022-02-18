/********************************************************************
 created:    2022-01-09
 author:     lixianmin

 Copyright (C) - All Rights Reserved
 *********************************************************************/
import {OctetsStream, SeekOrigin} from "./octets_stream";
import {BufferTools} from "./buffer_tools";

export class Packet {
    /**
     * Package protocol encode.
     *
     * Pomelo package format:
     * +------+-------------+------------------+
     * | type | body length |       body       |
     * +------+-------------+------------------+
     *
     * Head: 4bytes
     *   0: package type,
     *      1 - handshake,
     *      2 - handshake ack,
     *      3 - heartbeat,
     *      4 - data
     *      5 - kick
     *   1 - 3: big-endian body length
     * Body: body length bytes
     *
     * @param  {Number}    type   package type
     * @param  {Uint8Array} body  body content in bytes
     * @return {Uint8Array}       new byte array that contains encode result
     */
    public static encode(type: number, body?: Uint8Array): Uint8Array {
        const length = body ? body.length : 0
        const headSize = 4
        const buffer = new Uint8Array(headSize + length)

        let index = 0;
        buffer[index++] = type & 0xff
        buffer[index++] = (length >> 16) & 0xff
        buffer[index++] = (length >> 8) & 0xff
        buffer[index++] = length & 0xff

        if (body) {
            BufferTools.blockCopy(body, 0, buffer, index, length)
        }

        return buffer
    }

    /**
     * Package protocol decode.
     * See encode for package format.
     *
     * @param  {OctetsStream} stream
     * @return {Packet}
     */
    public static decode(stream: OctetsStream): Packet[] {
        const list: Packet[] = []
        const headSize = 4

        while (stream.getLength() - stream.getPosition() >= headSize) {
            const type = stream.readByte()
            const size = stream.readByte() << 16 | stream.readByte() << 8 | stream.readByte() >>> 0
            if (size < 0) {
                throw new Error(`type=${type}, length = ${size}, stream=${stream.toString()}`)
            }

            // 剩下的stream长度不够了, 则把刚刚读的4个字节吐出来
            if (stream.getLength() < size) {
                stream.seek(-4, SeekOrigin.Current)
                break
            }

            const body = new Uint8Array(size)
            if (size > 0) {
                stream.read(body, 0, size)
            }

            let pack = new Packet(type, body)
            list.push(pack)
        }

        stream.tidy()
        return list
    }

    private constructor(type: number, body: Uint8Array) {
        this.type = type
        this.body = body
    }

    public readonly type: number
    public readonly body: Uint8Array
}
