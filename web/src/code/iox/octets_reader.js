'use strict'

/********************************************************************
 created:    2023-06-09
 author:     lixianmin

 Copyright (C) - All Rights Reserved
 *********************************************************************/

export function newOctetsReader(stream) {
    const textDecoder = new TextDecoder()

    function readByte() {
        return stream.readByte()
    }

    function readInt32() {
        return stream.readInt32()
    }

    function read7BitEncodedInt() {
        let num = 0
        for (let i = 0; i < 28; i += 7) {
            const b = stream.readByte()
            num |= (b & 0x7F) << i
            if (b <= 127) {
                return num
            }
        }

        const b = stream.readByte()
        if (b > 15) {
            throw new Error('ErrBad7BitInt')
        }

        return num | (b << 28)
    }

    function readBytes() {
        const size = read7BitEncodedInt()
        if (size < 0) {
            throw new Error('ErrNegativeSize')
        }

        if (size === 0) {
            return undefined
        }

        const data = new Uint8Array(size)
        const num = stream.read(data, 0, size)
        if (num !== size) {
            throw new Error('ErrNotEnoughData')
        }

        return data
    }

    function readString() {
        const data = readBytes()
        return textDecoder.decode(data)
    }

    return {
        readByte: readByte,
        readInt32: readInt32,
        read7BitEncodedInt: read7BitEncodedInt,
        readBytes: readBytes,
        readString: readString,
        get stream() {
            return stream
        }
    }
}