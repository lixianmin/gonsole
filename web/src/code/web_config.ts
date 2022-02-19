/********************************************************************
 created:    2022-01-19
 author:     lixianmin

 Copyright (C) - All Rights Reserved
 *********************************************************************/

export class WebConfig {
    public constructor(config) {
        // console.log(config)
        this.autoLoginLimit = config.autoLoginLimit
        this.websocketPath = config.websocketPath
    }

    public getWebsocketUrl(host): string {
        const isHttps = "https:" === document.location.protocol
        const protocol = isHttps ? "wss://" : "ws://"
        const url = `${protocol}${host}/${this.websocketPath}`
        return url
    }

    public readonly autoLoginLimit: number
    private readonly websocketPath: string
}