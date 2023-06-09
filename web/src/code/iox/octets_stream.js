/********************************************************************
 created:    2023-06-08
 author:     lixianmin

 Copyright (C) - All Rights Reserved
 *********************************************************************/
import {Buffers} from "@src/code/iox/buffers";

export const SeekOrigin = {
    Begin: 0,
    Current: 1,
    End: 2
}

export function newOctetsStream(capacity = 8) {
    if (capacity < 0) {
        throw new Error(`invalid capacity=${capacity}`)
    }

    const maxCapacity = 2147483647
    let _buffer = new Uint8Array(capacity)  // _buffer.length is the capacity of the buffer
    let _length = 0   // _length points to the array part that contains data, this is the same as the len() of golang
    let _position = 0 // _position is both relevant with reading data and writing data, this is different from golang version

    // function setPosition(position) {
    //     if (position < 0) {
    //         throw new Error(`position=${position}`)
    //     }
    //
    //     _position = position
    // }
    //
    // function setLength(length) {
    //     if (length > _buffer.length) {
    //         throw new Error("length can not be greater than capacity")
    //     }
    //
    //     if (length < 0 || length > maxCapacity) {
    //         throw new Error("out of range")
    //     }
    //
    //     const num = length
    //     if (num > _length) {
    //         expand(num)
    //     }
    //
    //     _length = num
    //     if (_position > _length) {
    //         _position = _length
    //     }
    // }

    function expand(newSize) {
        const capacity = _buffer.length
        if (newSize > capacity) {
            let num = newSize
            if (num < 32) {
                num = 32
            } else if (num < capacity << 1) {
                num = capacity << 1
            }

            if (_length > 0) {
                const array = new Uint8Array(num)
                for (let i = 0; i < _length; i++) {
                    array[i] = _buffer[i]
                }

                _buffer = array
            }
        }
    }

    function readByte() {
        if (_position >= _length) {
            throw new Error(`NotEnoughData: _position=${_position}, _length=${_length}`)
        }

        return _buffer[_position++]
    }

    function readInt32() {
        const readSize = 4
        if (_position + readSize > _length) {
            throw new Error(`NotEnoughData: _position=${_position}, _length=${_length}`)
        }

        const b = _buffer
        const p = _position
        let num = b[p] | (b[p + 1] << 8) | (b[p + 2] << 16) || (b[p + 3] << 24)

        _position += readSize
        return num
    }

    function read(buffer, offset, count) {
        if (offset < 0 || count < 0) {
            throw new Error(`InvalidArgument: offset=${offset}, count=${count}`)
        }

        if (_buffer.length - offset < count) {
            throw new Error("the size of buffer is less than offset + count")
        }

        if (_position >= _length || count === 0) {
            return 0
        }

        if (_position >= _length - count) {
            count = _length - _position
        }

        Buffers.blockCopy(_buffer, _position, buffer, offset, count)
        _position += count
        return count
    }

    function writeByte(b) {
        const nextLength = _position + 1
        if (nextLength > _length) {
            expand(nextLength);
            _length = nextLength;
        }

        _buffer[_position++] = b;
    }

    function writeInt32(d) {
        const nextLength = _position + 4
        if (nextLength > _length) {
            expand(nextLength)
            _length = nextLength
        }

        _buffer[_position++] = d & 0xff
        _buffer[_position++] = (d >> 8) & 0xff
        _buffer[_position++] = (d >> 16) & 0xff
        _buffer[_position++] = (d >> 24) & 0xff
    }

    function write(buffer, offset, count) {
        if (offset < 0 || count < 0) {
            throw new Error(`offset=${offset}, count=${count}`)
        }

        if (buffer.length - offset < count) {
            throw new Error("the size of the buffer is less than offset + count")
        }

        if (_position > _length - count) {
            expand(_position + count)
        }

        Buffers.blockCopy(buffer, offset, _buffer, _position, count)
        _position += count

        if (_position >= _length) {
            _length = _position
        }
    }

    function seek(offset, location) {
        if (offset > maxCapacity) {
            throw new Error("offset is out of range")
        }

        let num
        switch (location) {
            case SeekOrigin.Begin:
                if (offset < 0) {
                    throw new Error("attempted to seek before start of OctetsSteam")
                }
                num = 0
                break
            case SeekOrigin.Current:
                num = _position
                break
            case SeekOrigin.End:
                num = _length
                break
            default:
                throw new Error("invalid SeekOrigin")
        }

        num += offset
        if (num < 0) {
            throw new Error("attempted to seek before start of OctetsStream")
        }

        _position = num
        return _position
    }

    function tidy() {
        const count = _length - _position
        Buffers.blockCopy(_buffer, _position, _buffer, 0, count)

        _position = 0
        _length = count
    }

    function reset() {
        _position = 0
        _length = 0
    }

    return {
        readByte: readByte,
        readInt32: readInt32,
        read: read,
        writeByte: writeByte,
        writeInt32: writeInt32,
        write: write,
        seek: seek,
        tidy: tidy,
        reset: reset,
        get buffer() {
            return _buffer
        }
    }
}