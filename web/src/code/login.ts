import ls from 'localstorage-slim'
import bcrypt from 'bcryptjs'

/********************************************************************
 created:    2022-01-20
 author:     lixianmin

 Copyright (C) - All Rights Reserved
 *********************************************************************/

export function createLogin(sendLogin: Function) {
    const key = "autoLoginUser"

    function save(username: string, digest: string, autoLoginLimit: number) {
        const item = {
            username: username,
            digest: digest,
        }

        ls.set(key, item, {ttl: autoLoginLimit})
    }

    function doLogin(username: string, digest: string) {
        sendLogin("auth", username, digest)
    }

    return {
        // 自动登录
        tryAutoLogin() {
            const item = ls.get(key)
            if (item) {
                // @ts-ignore
                doLogin(item.username, item.digest)
            }
        },

        async login(username: string, password: string, autoLoginLimit: number) {
            const salt = await bcrypt.genSalt()
            const digest = await bcrypt.hash(password, salt)
            // console.log('salt', salt, 'digest', digest)

            doLogin(username, digest)
            save(username, digest, autoLoginLimit)
        },
    }
}