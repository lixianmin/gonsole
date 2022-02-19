/********************************************************************
 created:    2022-01-19
 author:     lixianmin

 Copyright (C) - All Rights Reserved
 *********************************************************************/

export class WebConfig {
    public loadData(data) {
        console.log(data)
        this.autoLoginLimit = data.autoLoginLimit
        this.websocketPath = data.websocketPath
        this.urlRoot = data.urlRoot

        this.title = data.title
        this.body = data.body
    }

    public getAutoLoginLimit(): number {
        return this.autoLoginLimit
    }

    public getWebsocketUrl(host): string {
        const isHttps = "https:" === document.location.protocol
        const protocol = isHttps ? "wss:" : "ws:"
        const url = `${protocol}//${host}/${this.websocketPath}`
        return url
    }

    public getUrlRoot(): string {
        return this.urlRoot
    }

    public getTitle(): string {
        return this.title
    }

    public getBody(): string {
        return this.body
    }

    private autoLoginLimit: number = 0
    private websocketPath: string = ""
    private urlRoot: string = ""

    private title: string = ""
    private body: string = ""
}