'use strict'

/********************************************************************
 created:    2023-06-08
 author:     lixianmin

 Copyright (C) - All Rights Reserved
 *********************************************************************/

export function newOctetsWriter(stream) {
    const textEncoder = new TextEncoder()

    function writeByte(b) {
        stream.writeByte(b)
    }

    function writeInt32(d) {
        stream.writeInt32(d)
    }

    // write7BitEncodedInt 开启整数压缩
    function write7BitEncodedInt(d) {
        let num = d
        while (num > 127) {
            stream.writeByte(num | 0xFFFFFF80)
            num >>= 7
        }

        stream.writeByte(num)
    }

    function writeBytes(data) {
        const size = data?.length ?? 0
        write7BitEncodedInt(size)
        stream.write(data, 0, size)
    }

    function writeString(s) {
        const data = textEncoder.encode(s)
        writeBytes(data)
    }

    return {
        writeByte: writeByte,
        writeInt32: writeInt32,
        write7BitEncodedInt: write7BitEncodedInt,
        writeBytes: writeBytes,
        writeString: writeString,
        get stream() {
            return stream
        }
    }
}