import { afterEach, beforeEach, describe, expect, test } from "bun:test"
import { createLogin } from "@src/code/login"

let originalWindow: any

beforeEach(() => {
    originalWindow = globalThis.window
    globalThis.window = {
        navigator: {userAgent: "bun-test", language: "zh-CN"},
        screen: {width: 1920, height: 1080},
    } as any
})

afterEach(() => {
    globalThis.window = originalWindow
})

describe("login", () => {
    test("服务端返回错误时不应该抛出异常", async () => {
        // 模拟服务端返回错误, 此时sendLogin拿到的是undefined
        const login = createLogin(async () => undefined)

        await login.login("alice", "secret")
    })

    test("服务端返回错误时不应该缓存token", async () => {
        const login = createLogin(async () => undefined)
        await login.login("alice", "secret")

        // 缓存中不应该有token
        const calls: any[][] = []
        const auto = createLogin(async (...args: any[]) => {
            calls.push(args)
            return {code: "ok"}
        })
        await auto.tryAutoLogin(1)

        expect(calls.length).toBe(0)
    })

    test("登录成功时应当缓存token并在自动登录时复用", async () => {
        const login = createLogin(async () => ({code: "ok", token: "jwt-token"}))
        await login.login("alice", "secret")

        const calls: any[][] = []
        const auto = createLogin(async (...args: any[]) => {
            calls.push(args)
            return {code: "ok"}
        })
        await auto.tryAutoLogin(1)

        expect(calls.length).toBe(1)
        expect(calls[0][0]).toBe("auth")
        expect(calls[0][1]).toBe("alice")
        expect(calls[0][2]).toBe("jwt-token")
    })
})
