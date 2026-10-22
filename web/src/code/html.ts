// 转义HTML特殊字符, 用于innerHTML拼接的场景
// 修正: 之前JsonTable的表头、LogList的日志路径与样本都是直接拼进innerHTML的,
// 服务端返回的内容一旦包含尖括号或引号就会被当成HTML解析
export function escapeHtml(text: string): string {
    return String(text ?? "")
        .replace(/&/g, "&amp;")
        .replace(/</g, "&lt;")
        .replace(/>/g, "&gt;")
        .replace(/"/g, "&quot;")
        .replace(/'/g, "&#39;")
}

// 转义URL路径中的每一段, 保留目录分隔符
export function encodeUrlPath(path: string): string {
    return String(path ?? "")
        .split("/")
        .map(segment => encodeURIComponent(segment))
        .join("/")
}

// 把纯文本转换成HTML, 换行转换为<br>
export function textToHtml(text: string): string {
    return escapeHtml(text).replace(/\r\n|\r|\n/g, "<br>")
}
