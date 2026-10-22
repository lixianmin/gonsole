import { describe, expect, test } from "bun:test"
import { newOctetsStream, SeekOrigin } from "@src/code/iox/octets_stream"

describe("readInt32", () => {
    // 表驱动: [名称, 写入的int32值, 期望读回的值]
    const cases = [
        ["零", 0, 0],
        ["小值", 1, 1],
        ["127", 127, 127],
        ["128", 128, 128],
        ["255", 255, 255],
        ["256", 256, 256],
        ["65535", 65535, 65535],
        ["65536", 65536, 65536],
        ["0x01020304", 0x01020304, 0x01020304],
        ["负数-1", -1, -1],
        ["负数-2", -2, -2],
        ["int32最大值", 2147483647, 2147483647],
        ["int32最小值", -2147483648, -2147483648],
    ]

    for (const [name, value, expected] of cases) {
        test(`读写往返: ${name}`, () => {
            const stream = newOctetsStream(16)
            stream.writeInt32(value)
            stream.seek(0, SeekOrigin.Begin)

            expect(stream.readInt32()).toBe(expected)
        })
    }
})

describe("read", () => {
    test("目标缓冲区小于count时应当报错", () => {
        const stream = newOctetsStream(8)
        stream.write(new Uint8Array([1, 2, 3, 4]), 0, 4)
        stream.seek(0, SeekOrigin.Begin)

        // 目标只有2字节, 却要求读4字节, 属于越界写
        const dest = new Uint8Array(2)
        expect(() => stream.read(dest, 0, 4)).toThrow()
    })

    test("offset+count超出目标缓冲区时应当报错", () => {
        const stream = newOctetsStream(8)
        stream.write(new Uint8Array([1, 2, 3, 4]), 0, 4)
        stream.seek(0, SeekOrigin.Begin)

        const dest = new Uint8Array(4)
        expect(() => stream.read(dest, 2, 3)).toThrow()
    })
})
