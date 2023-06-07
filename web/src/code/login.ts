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

    // 自定义指纹, 防止token被窃用, 不使用fingerprintJS的原因是后者太容易变了, 影响自动登录
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

    let _nonce = 0
    return {
        // 自动登录
        async tryAutoLogin(nonce: number) {
            _nonce = nonce

            const item = ls.get(loginKey)
            if (item) {
                // @ts-ignore
                await doLogin(item.username, item.token)
            }
        },

        async login(username: string, password: string) {
            // 通常不应该使用固定盐
            const salt = "Hey Nurse!!"

            // 取hash值, 避免后端拿到原始password, 避免后端存储原始password或泄露到日志中
            const substitute = sha256(password + salt)

            // 使用nonce弱加密, 加密强度不重要, 重要的是避免replay attack
            // 这里返回的pw.words是array[8], 每一位是int32;
            // 在golang里计算返回的是array[32], 每一位是byte, 因此是 byte[32]
            substitute.words = substitute.words.map((item: number) => item ^ _nonce)

            // 这个digest的固定长度为44
            const digest = Base64.stringify(substitute)
            // console.log(`password=${password}, digest=${digest}, length=${digest.length}`)

            await doLogin(username, digest)
        },
    }
}