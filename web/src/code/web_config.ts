/********************************************************************
 created:    2022-01-19
 author:     lixianmin

 Copyright (C) - All Rights Reserved
 *********************************************************************/

export class WebConfig {
    public loadData(data) {
        // console.log(data)
        this.autoLoginLimit = data.autoLoginLimit
        this.websocketPath = data.websocketPath
    }

    public getWebsocketUrl(host): string {
        const isHttps = "https:" === document.location.protocol
        const protocol = isHttps ? "wss:" : "ws:"
        const url = `${protocol}//${host}/${this.websocketPath}`
        return url
    }

    public getAutoLoginLimit():number {
        return this.autoLoginLimit
    }

    private autoLoginLimit: number = 0
    private websocketPath: string = ""
}