import ls from 'localstorage-slim'
import sha256 from 'crypto-js/sha256'
import Base64 from 'crypto-js/enc-base64'

/********************************************************************
 created:    2022-01-20
 author:     lixianmin

 Copyright (C) - All Rights Reserved
 *********************************************************************/

export function createLogin(sendLogin: Function) {
    const loginKey = "auto.login.user"

    function fetchFingerprint() {
        const data = JSON.stringify([
            window.navigator.userAgent,
            window.navigator.language,  // language of the browser which is set when launching the browser, such as zh-CN
            window.screen.width,        // width of the screen in pixels
            window.screen.height,
            new Date().getTimezoneOffset(), // -480
        ])

        const fingerprint = Base64.stringify(sha256(data))
        // console.log(`data=${data}, fingerprint=${fingerprint}`)
        return fingerprint
    }

    async function doLogin(username: string, digestOrToken: string) {
        const fingerprint = fetchFingerprint()
        const response = await sendLogin("auth", username, digestOrToken, fingerprint)

        // 如果返回了token, 说明是使用digest登录的, 说明client需要缓存jwt
        if (response.code === 'ok') {
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
            // 盐值是固定的, 但每一个项目应该不一样
            const salt = "Hey Nurse!!"

            // 这个digest的固定长度为44
            const digest = Base64.stringify(sha256(password + salt))
            // console.log(`password=${password}, digest=${digest}, length=${digest.length}`)

            await doLogin(username, digest)
        },
    }
}