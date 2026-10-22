import { describe, expect, test } from "bun:test"
import { parseConsoleInput, ParseContext, ParseResult } from "@src/code/console_input"

// 构造一个历史记录存储器, 行为与use_history_store一致
function createHistory(list: string[]) {
    return {
        getHistory: (index: number) => (index >= 0 && index < list.length ? list[index] : ""),
        getHistoryCount: () => list.length,
    }
}

const history = createHistory(["help", "auth alice", "tail /var/log/app.log"])

const ctx: ParseContext = {
    isAuthorizing: false,
    getHistory: history.getHistory,
    historyCount: history.getHistoryCount(),
}

describe("parseConsoleInput", () => {
    // 表驱动: [名称, 输入, 上下文, 期望结果]
    const cases: [string, string, ParseContext, ParseResult][] = [
        ["空输入", "", ctx, {type: "empty"}],
        ["只有空格", "   ", ctx, {type: "empty"}],
        ["普通命令", "tail /var/log/app.log", ctx, {type: "command", command: "tail /var/log/app.log"}],
        ["普通命令合并连续空格", "tail    /var/log/app.log", ctx, {type: "command", command: "tail /var/log/app.log"}],
        ["help", "help", ctx, {type: "help"}],
        ["sub", "sub topic1", ctx, {type: "sub", topic: "topic1"}],
        ["unsub", "unsub topic1", ctx, {type: "unsub", topic: "topic1"}],
        ["auth", "auth alice", ctx, {type: "auth", username: "alice"}],
        ["request", 'request console.command {"command":"help"}', ctx,
            {type: "roadRequest", route: "console.command", data: '{"command":"help"}'}],
        ["只有一个参数的sub按普通命令处理", "sub", ctx, {type: "command", command: "sub"}],
        ["只有一个参数的auth按普通命令处理", "auth", ctx, {type: "command", command: "auth"}],
        ["只有一个参数的request按普通命令处理", "request", ctx, {type: "command", command: "request"}],
    ]

    for (const [name, input, context, expected] of cases) {
        test(name, () => {
            expect(parseConsoleInput(input, context)).toEqual(expected)
        })
    }
})

describe("历史命令引用", () => {
    const cases: [string, string, ParseResult][] = [
        ["!1引用第一条", "!1", {type: "help"}],
        ["!2引用第二条", "!2", {type: "auth", username: "alice"}],
        ["!3引用第三条", "!3", {type: "command", command: "tail /var/log/app.log"}],
        ["!4越界", "!4", {type: "unknownHistory", reference: "4"}],
        ["!0非法", "!0", {type: "unknownHistory", reference: "0"}],
        ["!999越界", "!999", {type: "unknownHistory", reference: "999"}],
        ["!1.5非法", "!1.5", {type: "unknownHistory", reference: "1.5"}],
        ["!abc不是引用按普通命令处理", "!abc", {type: "command", command: "!abc"}],
        ["! 2 带空格", "! 2", {type: "auth", username: "alice"}],
    ]

    for (const [name, input, expected] of cases) {
        test(name, () => {
            expect(parseConsoleInput(input, ctx)).toEqual(expected)
        })
    }

    test("历史记录本身以!开头时不会无限递归", () => {
        const loopHistory = createHistory(["!1", "!1"])
        const loopCtx: ParseContext = {
            isAuthorizing: false,
            getHistory: loopHistory.getHistory,
            historyCount: loopHistory.getHistoryCount(),
        }

        expect(parseConsoleInput("!1", loopCtx)).toEqual({type: "command", command: "!1"})
    })
})

describe("输入密码", () => {
    const authCtx: ParseContext = {
        isAuthorizing: true,
        getHistory: ctx.getHistory,
        historyCount: ctx.historyCount,
    }

    // 表驱动: [名称, 输入, 期望密码]
    const cases: [string, string, string][] = [
        ["单个字符", "a", "a"],
        ["多个字符", "secret", "secret"],
        ["带空格的密码", "my secret password", "my secret password"],
        ["首尾空格被去掉", "  secret  ", "secret"],
        ["密码是help", "help", "help"],
        ["密码是auth", "auth bob", "auth bob"],
    ]

    for (const [name, input, expected] of cases) {
        test(name, () => {
            expect(parseConsoleInput(input, authCtx)).toEqual({type: "password", password: expected})
        })
    }
})
