import { describe, expect, test } from "bun:test"
import { escapeHtml, encodeUrlPath, textToHtml } from "@src/code/html"

describe("escapeHtml", () => {
    // 表驱动: [名称, 输入, 期望输出]
    const cases: [string, string, string][] = [
        ["普通文本", "hello", "hello"],
        ["空字符串", "", ""],
        ["小于号", "<b>", "&lt;b&gt;"],
        ["大于号", "a>b", "a&gt;b"],
        ["与号", "a&b", "a&amp;b"],
        ["双引号", 'a"b', "a&quot;b"],
        ["单引号", "a'b", "a&#39;b"],
        ["已转义的文本不会被二次转义", "&lt;", "&amp;lt;"],
        ["script标签", "<script>alert(1)</script>", "&lt;script&gt;alert(1)&lt;/script&gt;"],
    ]

    for (const [name, input, expected] of cases) {
        test(name, () => {
            expect(escapeHtml(input)).toBe(expected)
        })
    }
})

describe("encodeUrlPath", () => {
    // 表驱动: [名称, 输入, 期望输出]
    const cases: [string, string, string][] = [
        ["普通文件名", "gonsole.log", "gonsole.log"],
        ["包含空格", "my log.log", "my%20log.log"],
        ["包含与号", "a&b.log", "a%26b.log"],
        ["保留目录分隔符", "sub/dir/a.log", "sub/dir/a.log"],
        ["目录中的空格", "sub dir/a.log", "sub%20dir/a.log"],
        ["单引号在URL中合法", "a'b.log", "a'b.log"],
    ]

    for (const [name, input, expected] of cases) {
        test(name, () => {
            expect(encodeUrlPath(input)).toBe(expected)
        })
    }
})

describe("textToHtml", () => {
    // 表驱动: [名称, 输入, 期望输出]
    const cases: [string, string, string][] = [
        ["普通文本", "hello", "hello"],
        ["换行", "a\nb", "a<br>b"],
        ["windows换行", "a\r\nb", "a<br>b"],
        ["回车", "a\rb", "a<br>b"],
        ["尖括号被转义", "<b>", "&lt;b&gt;"],
        ["尖括号与换行", "<b>\nx", "&lt;b&gt;<br>x"],
    ]

    for (const [name, input, expected] of cases) {
        test(name, () => {
            expect(textToHtml(input)).toBe(expected)
        })
    }
})
