import { afterEach, beforeEach, describe, expect, test } from "bun:test"
import { newSession } from "@src/code/road/session"
import { newOctetsStream, SeekOrigin } from "@src/code/iox/octets_stream"
import { newOctetsWriter } from "@src/code/iox/octets_writer"
import { encode } from "@src/code/road/packet_tools"
import { newJsonSerde } from "@src/code/road/json_serde"
import { PacketKind } from "@src/code/road/consts"

const serde = newJsonSerde()

// 假的WebSocket, 用来在测试中驱动client与server之间的报文收发
class FakeWebSocket {
    static instances = []

    constructor(url) {
        this.url = url
        this.readyState = 0
        this.sent = []
        FakeWebSocket.instances.push(this)
    }

    send(bytes) {
        this.sent.push(new Uint8Array(bytes))
    }

    close() {
        this.readyState = 3
        this.onclose?.({code: 1000, reason: "normal"})
    }

    // 测试辅助: 模拟server推送一个报文
    receive(bytes) {
        this.onmessage?.({data: bytes.buffer})
    }
}

function encodePacket(pack) {
    const writer = newOctetsWriter(newOctetsStream())
    writer.stream.reset()
    encode(writer, pack)
    writer.stream.seek(0, SeekOrigin.Begin)
    return writer.stream.bytes
}

async function deflateRawBase64(text) {
    const compressed = new Blob([new TextEncoder().encode(text)])
        .stream()
        .pipeThrough(new CompressionStream("deflate-raw"))
    const bytes = new Uint8Array(await new Response(compressed).arrayBuffer())
    return btoa(String.fromCharCode(...bytes))
}

function tick() {
    return new Promise(resolve => setTimeout(resolve, 10))
}

// 完成一次握手, 使session建立起kind与route的映射关系
async function handshake(socket, routes) {
    const data = serde.serialize({
        nonce: 12345,
        routes: await deflateRawBase64(routes),
        heartbeat: 3600,
    })
    socket.receive(encodePacket({kind: PacketKind.Handshake, requestId: 0, data}))
    await tick()
}

// 模拟server向client推送一个route报文
function push(socket, kind, payload) {
    socket.receive(encodePacket({kind, requestId: 0, data: serde.serialize(payload)}))
}

let documentStub
let originalWebSocket
let originalDocument

beforeEach(() => {
    FakeWebSocket.instances = []

    documentStub = {
        visibilityState: "visible",
        listeners: {},
        addEventListener(type, handler) {
            (this.listeners[type] ??= []).push(handler)
        },
    }
    originalDocument = globalThis.document
    originalWebSocket = globalThis.WebSocket
    globalThis.document = documentStub
    globalThis.WebSocket = FakeWebSocket
})

afterEach(() => {
    globalThis.document = originalDocument
    globalThis.WebSocket = originalWebSocket
})

describe("route处理器", () => {
    test("握手后收到route报文时应当调用对应处理器", async () => {
        const session = newSession()
        const received = []
        session.on("console.test", (response) => received.push(response))

        session.connect("ws://localhost:8888/ws")
        const socket = FakeWebSocket.instances[0]
        await handshake(socket, "console.test")

        push(socket, PacketKind.UserBase, {text: "hello"})

        expect(received).toEqual([{text: "hello"}])
    })

    test("断线重连后route处理器仍然有效", async () => {
        const session = newSession()
        const received = []
        session.on("console.test", (response) => received.push(response))

        session.connect("ws://localhost:8888/ws")
        const first = FakeWebSocket.instances[0]
        await handshake(first, "console.test")

        // 模拟断线, session会自动重连
        first.close()
        expect(FakeWebSocket.instances.length).toBe(2)
        const second = FakeWebSocket.instances[1]
        await handshake(second, "console.test")

        push(second, PacketKind.UserBase, {text: "after reconnect"})

        expect(received).toEqual([{text: "after reconnect"}])
    })
})

describe("断线通知", () => {
    test("连接断开时应当通知disconnect处理器", async () => {
        const session = newSession()
        let disconnected = 0
        session.on("disconnect", () => disconnected++)

        session.connect("ws://localhost:8888/ws")
        const socket = FakeWebSocket.instances[0]
        await handshake(socket, "console.test")

        socket.close()

        expect(disconnected).toBe(1)
    })
})

describe("request", () => {
    test("未知route时不应该发送报文", async () => {
        const session = newSession()
        session.connect("ws://localhost:8888/ws")
        const socket = FakeWebSocket.instances[0]
        await handshake(socket, "console.test")

        const before = socket.sent.length
        session.request("console.not-exist", {command: "help"})

        expect(socket.sent.length).toBe(before)
    })

    test("重连后旧的requestId不再有对应的处理器", async () => {
        const session = newSession()
        const received = []
        session.connect("ws://localhost:8888/ws")
        const first = FakeWebSocket.instances[0]
        await handshake(first, "console.test")

        session.request("console.test", {command: "help"}, (response) => received.push(response))
        const requestId = 1

        // 断线重连, 服务端此前未响应的请求应当被丢弃
        first.close()
        const second = FakeWebSocket.instances[1]
        await handshake(second, "console.test")

        // 服务端迟到的响应
        second.receive(encodePacket({
            kind: PacketKind.UserBase,
            requestId,
            data: serde.serialize({text: "late"}),
        }))

        expect(received).toEqual([])
    })
})

describe("visibilitychange", () => {
    test("尚未连接时触发visibilitychange不应该抛异常", () => {
        const session = newSession()

        expect(() => {
            for (const handler of documentStub.listeners["visibilitychange"] ?? []) {
                handler()
            }
        }).not.toThrow()
    })
})
