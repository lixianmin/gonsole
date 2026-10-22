import { describe, expect, test } from "bun:test"
import { useKeyDown, longestCommonPrefix, getHumanReadableSize } from "@src/code/tools"

// 假的DOM节点, 记录注册的事件监听器
function createTargetStub() {
    const listeners: Record<string, ((evt: any) => void)[]> = {}
    return {
        listeners,
        addEventListener(type: string, handler: (evt: any) => void) {
            (listeners[type] ??= []).push(handler)
        },
        dispatch(type: string, evt: any) {
            for (const handler of listeners[type] ?? []) {
                handler(evt)
            }
        },
    }
}

describe("useKeyDown", () => {
    test("按键匹配时应当调用handler", () => {
        const target = createTargetStub()
        let called = 0
        useKeyDown(target as unknown as HTMLElement, "Enter", () => called++)

        target.dispatch("keydown", {key: "Enter"})

        expect(called).toBe(1)
    })

    test("按键不匹配时不应该调用handler", () => {
        const target = createTargetStub()
        let called = 0
        useKeyDown(target as unknown as HTMLElement, "Enter", () => called++)

        target.dispatch("keydown", {key: "Escape"})

        expect(called).toBe(0)
    })

    test("支持多个按键", () => {
        const target = createTargetStub()
        const called: string[] = []
        useKeyDown(target as unknown as HTMLElement, ["ArrowUp", "ArrowDown"], (evt: any) => called.push(evt.key))

        target.dispatch("keydown", {key: "ArrowUp"})
        target.dispatch("keydown", {key: "ArrowDown"})
        target.dispatch("keydown", {key: "Enter"})

        expect(called).toEqual(["ArrowUp", "ArrowDown"])
    })

    test("handler能够阻止默认行为", () => {
        const target = createTargetStub()
        let prevented = false
        useKeyDown(target as unknown as HTMLElement, "Enter", (evt: any) => evt.preventDefault())

        target.dispatch("keydown", {key: "Enter", preventDefault: () => prevented = true})

        expect(prevented).toBe(true)
    })
})

describe("longestCommonPrefix", () => {
    // 表驱动: [名称, 输入, 期望输出]
    const cases: [string, string[], string][] = [
        ["空数组", [], ""],
        ["单个元素", ["abc"], "abc"],
        ["相同前缀", ["abc", "abd"], "ab"],
        ["没有公共前缀", ["abc", "xyz"], ""],
        ["前缀关系", ["abc", "abcdef"], "abc"],
    ]

    for (const [name, input, expected] of cases) {
        test(name, () => {
            expect(longestCommonPrefix(input)).toBe(expected)
        })
    }
})

describe("getHumanReadableSize", () => {
    // 表驱动: [名称, 输入, 期望输出]
    const cases: [string, number, string][] = [
        ["0", 0, "0B"],
        ["小于1K", 1023, "1023B"],
        ["1K", 1024, "1.0K"],
        ["1M", 1048576, "1.0M"],
    ]

    for (const [name, input, expected] of cases) {
        test(name, () => {
            expect(getHumanReadableSize(input)).toBe(expected)
        })
    }
})
