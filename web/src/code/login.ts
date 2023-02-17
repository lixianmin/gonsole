import ls from 'localstorage-slim'
import sha256 from 'crypto-js/sha256'
import Base64 from 'crypto-js/enc-base64'
import FingerprintJS from '@fingerprintjs/fingerprintjs'

/********************************************************************
 created:    2022-01-20
 author:     lixianmin

 Copyright (C) - All Rights Reserved
 *********************************************************************/

export function createLogin(sendLogin: Function) {
    const loginKey = "auto.login.user"

    async function fetchFingerprint() {
        const fp = await FingerprintJS.load()
        const result = await fp.get()
        return result.visitorId
    }

    async function doLogin(username: string, digestOrToken: string) {
        const fingerprint = await fetchFingerprint()
        const response = await sendLogin("auth", username, digestOrToken, fingerprint)

        // 如果返回了token, 说明是使用digest登录的, 说明client需要缓存jwt
        const code = response.code
        if (code === 'ok') {
            const token = response.token
            if (typeof token === 'string' && token.length > 0) {
                ls.set(loginKey, {username, token})
            }
        } else {
            ls.remove(loginKey)
        }
    }

    return {
        // 自动登录
        async tryAutoLogin() {
            const item = ls.get(loginKey)
            if (item) {
                // @ts-ignore
                await doLogin(item.username, item.token)
            }
        },

        async login(username: string, password: string) {
            // 这个digest的固定长度为44
            const digest = Base64.stringify(sha256(password))
            // console.log(`password=${password}, digest=${digest}, length=${digest.length}`)

            await doLogin(username, digest)
        },
    }
}