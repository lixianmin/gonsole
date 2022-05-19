import {sha256} from "js-sha256";

/********************************************************************
 created:    2022-01-20
 author:     lixianmin

 Copyright (C) - All Rights Reserved
 *********************************************************************/

export class Login {
    public constructor(sendLogin) {
        this.sendLogin = sendLogin
    }

    public login(username: string, password: string, autoLoginLimit: number) {
        this.doLogin(username, password)
        this.save(username, password, autoLoginLimit)
    }

    // 自动登录
    public tryAutoLogin() {
        const data = localStorage.getItem(this.key)
        if (data) {
            const item = JSON.parse(data)
            if (item && new Date().getTime() < item.expireTime) {
                this.doLogin(item.username, item.password)
            }
        }
    }

    private doLogin(username: string, password: string) {
        const key = "hey pet!"
        const digest = sha256.hmac(key, password)
        this.sendLogin("auth", username, digest)
    }

    private save(username: string, password: string, autoLoginLimit: number) {
        const item = {
            username: username,
            password: password,
            expireTime: new Date().getTime() + autoLoginLimit,
        }

        const data = JSON.stringify(item)
        localStorage.setItem(this.key, data)
    }

    private readonly sendLogin: any
    private readonly key = "autoLoginUser"
}