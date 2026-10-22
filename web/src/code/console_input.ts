export type ParseContext = {
    // 当前是否正在等待用户输入密码
    isAuthorizing: boolean
    // 按下标(从0开始)取历史命令, 越界时返回空字符串
    getHistory: (index: number) => string
    // 历史命令的总数
    historyCount: number
}

export type ParseResult =
    | {type: "empty"}
    | {type: "unknownHistory", reference: string}
    | {type: "password", password: string}
    | {type: "help"}
    | {type: "sub", topic: string}
    | {type: "unsub", topic: string}
    | {type: "auth", username: string}
    | {type: "roadRequest", route: string, data: string}
    | {type: "command", command: string}

// 解析用户在控制台输入的一行文本, 返回应当执行的动作
// 修正: 之前这段逻辑写在App.jsx的事件回调里, 无法单元测试, 且存在两个缺陷:
//   1. "!N"越界时historyStore.getHistory()返回undefined, 后续调用split()会抛出TypeError
//   2. 输入密码时只取了split之后的第一个词, 带空格的密码会被截断
export function parseConsoleInput(text: string, ctx: ParseContext): ParseResult {
    return parse(text, ctx, true)
}

function parse(text: string, ctx: ParseContext, allowHistory: boolean): ParseResult {
    const trimmed = (text ?? "").trim()
    if (trimmed === "") {
        return {type: "empty"}
    }

    // 正在等待输入密码时, 整行输入都是密码, 不能按命令解析
    if (ctx.isAuthorizing) {
        return {type: "password", password: trimmed}
    }

    if (allowHistory && trimmed.startsWith("!")) {
        const reference = trimmed.substring(1).trim()
        const index = Number(reference)
        if (reference !== "" && !Number.isNaN(index)) {
            if (!Number.isInteger(index) || index < 1 || index > ctx.historyCount) {
                return {type: "unknownHistory", reference}
            }

            // 历史命令本身可能以"!"开头, 因此不再对取出的内容重复做历史解析
            return parse(ctx.getHistory(index - 1), ctx, false)
        }
    }

    const texts = trimmed.split(/\s+/)
    const name = texts[0]
    const length = texts.length

    if (name === "help") {
        return {type: "help"}
    }

    if (length >= 2 && (name === "sub" || name === "unsub")) {
        return {type: name, topic: texts[1]}
    }

    if (length >= 2 && name === "auth") {
        return {type: "auth", username: texts[1]}
    }

    if (length >= 2 && name === "request") {
        return {type: "roadRequest", route: texts[1], data: texts.slice(2).join(" ")}
    }

    return {type: "command", command: texts.join(" ")}
}
