import { afterEach, beforeEach, describe, expect, test } from "bun:test"
import { createWebConfig } from "@src/code/web_config"

let originalDocument
let originalWindow

function setDocument(title) {
    globalThis.document = {
        title,
        location: {protocol: "http:", host: "localhost:8888"},
    }
    globalThis.window = {
        location: {protocol: "http:", host: "localhost:8888"},
    }
}

beforeEach(() => {
    originalDocument = globalThis.document
    originalWindow = globalThis.window
})

afterEach(() => {
    globalThis.document = originalDocument
    globalThis.window = originalWindow
})

describe("createWebConfig", () => {
    test("toString不应该抛出异常", () => {
        setDocument(JSON.stringify({
            directory: "ws",
            websocketPath: "",
            title: "Gonsole",
            body: "<h2>body</h2>",
        }))

        const config = createWebConfig()
        expect(() => config.toString()).not.toThrow()
    })

    test("toString的内容包含host与directory", () => {
        setDocument(JSON.stringify({
            directory: "ws",
            websocketPath: "",
            title: "Gonsole",
            body: "<h2>body</h2>",
        }))

        const config = createWebConfig()
        const text = config.toString()
        expect(text).toContain("directory=ws")
        expect(text).toContain("body=<h2>body</h2>")
    })

    test("标题不是模板时使用标题中的数据", () => {
        setDocument(JSON.stringify({
            directory: "console",
            websocketPath: "ws",
            title: "Gonsole",
            body: "<h2>body</h2>",
        }))

        const config = createWebConfig()
        expect(config.body).toBe("<h2>body</h2>")
        expect(config.getWebsocketUrl()).toBe("ws://localhost:8888/console/ws")
    })
})
