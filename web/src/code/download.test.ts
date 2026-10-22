import { describe, expect, test } from "bun:test"
import { fetchFileWithToken, HeaderAccessToken } from "@src/code/download"

function stubFetch(impl: (url: string, options: any) => Promise<any>) {
    const original = globalThis.fetch
    globalThis.fetch = impl as any
    return () => {
        globalThis.fetch = original
    }
}

describe("fetchFileWithToken", () => {
    test("令牌放在请求头里, 不能出现在url中", async () => {
        let requestedUrl = ""
        let requestedHeaders: Record<string, string> = {}
        const restore = stubFetch(async (url, options) => {
            requestedUrl = url
            requestedHeaders = options?.headers ?? {}
            return {ok: true, status: 200, blob: async () => "blob"}
        })

        try {
            await fetchFileWithToken("http://localhost:8888/logs/app.log", "my-token")
        } finally {
            restore()
        }

        expect(requestedUrl).toBe("http://localhost:8888/logs/app.log")
        expect(requestedHeaders[HeaderAccessToken]).toBe("my-token")
        expect(requestedUrl).not.toContain("my-token")
    })

    test("服务端返回非2xx时抛出异常", async () => {
        const restore = stubFetch(async () => ({ok: false, status: 401, blob: async () => ""}))

        try {
            await expect(fetchFileWithToken("http://localhost:8888/logs/app.log", "expired-token"))
                .rejects.toThrow(/401/)
        } finally {
            restore()
        }
    })

    test("服务端返回2xx时返回文件内容", async () => {
        const restore = stubFetch(async () => ({ok: true, status: 200, blob: async () => "file-content" as any}))

        try {
            const result: any = await fetchFileWithToken("http://localhost:8888/logs/app.log", "token")
            expect(result).toBe("file-content")
        } finally {
            restore()
        }
    })
})
