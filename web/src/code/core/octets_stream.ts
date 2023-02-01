/********************************************************************
 created:    2022-01-17
 author:     lixianmin

 Copyright (C) - All Rights Reserved
 *********************************************************************/
import {Buffers} from "./buffers";

export enum SeekOrigin {
    Begin,
    Current,
    End
}

export class OctetsStream {
    public constructor(capacity: number) {
        if (capacity < 0) {
            throw new Error(`capacity=${capacity}`)
        }

        this.capacity = capacity
        this.buffer = new Uint8Array(capacity)
    }

    public setCapacity(capacity: number) {
        if (capacity == this.capacity) {
            return
        }

        if (capacity != this.buffer.length) {
            if (capacity != 0) {
                const array = new Uint8Array(capacity)
                for (let i = 0; i < this.length; i++) {
                    array[i] = this.buffer[i]
                }

                this.buffer = array
            }

            this.dirtyBytes = 0
            this.capacity = capacity
        }
    }

    public getCapacity(): number {
        return this.capacity - this.initialIndex
    }

    public setPosition(position: number) {
        if (position < 0) {
            throw new Error(`position=${position}`)
        }

        this.position = this.initialIndex + position
    }

    public getPosition(): number {
        return this.position - this.initialIndex
    }

    public setLength(length: number) {
        if (length > this.capacity) {
            throw new Error("length can not be greater than capacity")
        }

        if (length < 0 || length + this.initialIndex > OctetsStream.maxCapacity) {
            throw new Error("out of range")
        }

        let num = length + this.initialIndex
        if (num > this.length) {
            this.expand(num)
        } else if (num < this.length) {
            this.dirtyBytes += this.length - num
        }

        this.length = num

        if (this.position > this.length) {
            this.position = this.length
        }
    }

    public getLength(): number {
        return this.length - this.initialIndex
    }

    private expand(newSize: number) {
        if (newSize > this.capacity) {
            let num = newSize
            if (num < 32) {
                num = 32
            } else if (num < this.capacity << 1) {
                num = this.capacity << 1
            }

            this.setCapacity(num)
        } else if (this.dirtyBytes > 0) {
            for (let i = 0; i < this.dirtyBytes; i++) {
                let index = i + this.length;
                this.buffer[index] = 0
            }

            this.dirtyBytes = 0
        }
    }

    public readByte(): number {
        if (this.position >= this.length) {
            return -1
        }

        return this.buffer[this.position++]
    }

    public read(buffer: Uint8Array, offset: number, count: number): number {
        if (offset < 0 || count < 0) {
            throw new Error(`offset=${offset}, count=${count}`)
        }

        if (this.buffer.length - offset < count) {
            throw new Error("the size of buffer is less than offset + count")
        }

        if (this.position >= this.length || count == 0) {
            return 0
        }

        if (this.position >= this.length - count) {
            count = this.length - this.position
        }

        Buffers.blockCopy(this.buffer, this.position, buffer, offset, count)
        this.position += count
        return count
    }

    public write(buffer: Uint8Array, offset: number, count: number) {
        if (offset < 0 || count < 0) {
            throw new Error(`offset=${offset}, count=${count}`)
        }

        if (buffer.byteLength - offset < count) {
            throw new Error("the size of the buffer is less than offset + count")
        }

        if (this.position > this.length - count) {
            this.expand(this.position + count)
        }

        Buffers.blockCopy(buffer, offset, this.buffer, this.position, count)
        this.position += count

        if (this.position >= this.length) {
            this.length = this.position
        }
    }

    public seek(offset: number, location: SeekOrigin) {
        if (offset > OctetsStream.maxCapacity) {
            throw new Error("offset is out of range")
        }

        let num: number
        switch (location) {
            case SeekOrigin.Begin:
                if (offset < 0) {
                    throw new Error("attempted to seek before start of OctetsSteam")
                }
                num = this.initialIndex
                break
            case SeekOrigin.Current:
                num = this.position
                break
            case SeekOrigin.End:
                num = this.length
                break
            default:
                throw new Error("invalid SeekOrigin")
        }

        num += offset
        if (num < this.initialIndex) {
            throw new Error("attempted to seek before start of OctetsStream")
        }

        this.position = num
        return this.position
    }

    public tidy() {
        const count = this.length - this.position
        Buffers.blockCopy(this.buffer, this.position, this.buffer, 0, count)

        this.setPosition(0)
        this.setLength(count)
    }

    public toString(): string {
        return `dirtyBytes=${this.dirtyBytes}, position=${this.position}, length=${this.length}, capacity=${this.capacity}, buffer=${this.buffer}`
    }

    private static maxCapacity = 2147483647

    private initialIndex = 0
    private dirtyBytes = 0
    private position = 0        // read|write from position
    private length = 0          // real data length
    private capacity = 0        // buffer size
    private buffer: Uint8Array
}