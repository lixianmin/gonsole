import {sha256} from "js-sha256";

/********************************************************************
 created:    2022-01-20
 author:     lixianmin

 Copyright (C) - All Rights Reserved
 *********************************************************************/

export function createLogin(sendLogin: Function) {
    const key = "autoLoginUser"

    function save(username: string, password: string, autoLoginLimit: number) {
        const item = {
            username: username,
            password: password,
            expireTime: new Date().getTime() + autoLoginLimit,
        }

        const data = JSON.stringify(item)
        localStorage.setItem(key, data)
    }

    function doLogin(username: string, password: string) {
        const key = "hey pet!"
        const digest = sha256.hmac(key, password)
        sendLogin("auth", username, digest)
    }

    return {
        // 自动登录
        tryAutoLogin() {
            const data = localStorage.getItem(key)
            if (data) {
                const item = JSON.parse(data)
                if (item && new Date().getTime() < item.expireTime) {
                    doLogin(item.username, item.password)
                }
            }
        },
        
        login(username: string, password: string, autoLoginLimit: number) {
            doLogin(username, password)
            save(username, password, autoLoginLimit)
        },
    }
}