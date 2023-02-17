import ls from 'localstorage-slim'
import sha256 from 'crypto-js/sha256'
import Base64 from 'crypto-js/enc-base64'

/********************************************************************
 created:    2022-01-20
 author:     lixianmin

 Copyright (C) - All Rights Reserved
 *********************************************************************/

export function createLogin(sendLogin: Function) {
    const tokenKey = "auto.login.user"

    async function doLogin(username: string, digestOrToken: string) {
        const response = await sendLogin("auth", username, digestOrToken)

        // 如果返回了token, 说明是使用digest登录的, 说明client需要缓存jwt
        switch (response.code) {
            case 'ok':
                if (response.token) {
                    const item = {
                        username: username,
                        token: response.token,
                    }

                    ls.set(tokenKey, item)
                    // console.log('response', response)
                }
                break
            case 'token_expired':
                ls.remove(tokenKey)
                break
        }
    }

    return {
        // 自动登录
        async tryAutoLogin() {
            const item = ls.get(tokenKey)
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