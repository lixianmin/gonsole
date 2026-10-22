import { afterEach, beforeAll, beforeEach, describe, expect, test } from "bun:test"

let createHistoryStore
let originalLocalStorage

function createLocalStorageStub(initial = {}) {
    const data = {...initial}
    return {
        data,
        getItem: key => (key in data ? data[key] : null),
        setItem: (key, value) => {
            data[key] = String(value)
        },
        removeItem: key => {
            delete data[key]
        },
        clear: () => {
            for (const key of Object.keys(data)) {
                delete data[key]
            }
        },
    }
}

beforeAll(async () => {
    originalLocalStorage = globalThis.localStorage
    globalThis.localStorage = createLocalStorageStub()

    // 模块在导入时就会创建全局store, 因此必须在准备好localStorage之后再导入
    const module = await import("@src/code/use_history_store")
    createHistoryStore = module.createHistoryStore
})

afterEach(() => {
    globalThis.localStorage = originalLocalStorage
})

describe("createHistoryStore", () => {
    beforeEach(() => {
        globalThis.localStorage = createLocalStorageStub()
    })

    test("localStorage中的内容损坏时不应该导致初始化失败", () => {
        globalThis.localStorage = createLocalStorageStub({"this.is.history.store": "{not a json"})

        const store = createHistoryStore()
        expect(store.getHistoryCount()).toBe(0)

        store.add("help")
        expect(store.getHistoryCount()).toBe(1)
    })

    test("localStorage中的内容不是数组时不应该导致初始化失败", () => {
        globalThis.localStorage = createLocalStorageStub({"this.is.history.store": "123"})

        const store = createHistoryStore()
        expect(store.getHistoryCount()).toBe(0)
    })

    test("历史命令条数应当有上限", () => {
        const store = createHistoryStore()
        for (let i = 0; i < 250; i++) {
            store.add("command-" + i)
        }

        expect(store.getHistoryCount()).toBeLessThanOrEqual(100)
        // 保留的是最近的命令
        expect(store.getHistory(store.getHistoryCount() - 1)).toBe("command-249")
    })

    test("getHistory越界时返回空字符串", () => {
        const store = createHistoryStore()
        store.add("help")

        expect(store.getHistory(0)).toBe("help")
        expect(store.getHistory(1)).toBe("")
        expect(store.getHistory(-1)).toBe("")
        expect(store.getHistory(999)).toBe("")
    })

    test("重复添加相同的命令不会产生重复记录", () => {
        const store = createHistoryStore()
        store.add("help")
        store.add("help")

        expect(store.getHistoryCount()).toBe(1)
    })
})
