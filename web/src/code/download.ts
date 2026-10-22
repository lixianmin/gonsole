// 访问令牌通过请求头传递, 不放在url的query里,
// 避免令牌进入浏览器历史与Referer
export const HeaderAccessToken = "X-Access-Token"

// 下载受令牌保护的文件, 令牌放在X-Access-Token请求头中
export async function fetchFileWithToken(url: string, token: string): Promise<Blob> {
    const response = await fetch(url, {headers: {[HeaderAccessToken]: token}})
    if (!response.ok) {
        throw new Error(`下载失败, status=${response.status}, 令牌可能已过期, 请重新执行log.list命令`)
    }

    return await response.blob()
}
