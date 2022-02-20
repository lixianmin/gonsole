/********************************************************************
 created:    2022-01-19
 author:     lixianmin

 Copyright (C) - All Rights Reserved
 *********************************************************************/

export class WebConfig {
    public constructor() {
        if (document.title != "{{.Data}}") {
            let data = JSON.parse(document.title)
            console.log("data:", data)

            this.host = window.location.host
            this.directory = data.directory
            this.autoLoginLimit = data.autoLoginLimit
            this.websocketPath = data.websocketPath

            document.title = data.title
            this.body = data.body
        } else {
            // 如果document.title没有变, 说明是在本地debug, 所以使用localhost:8888/ws
            this.host = "localhost:8888"
            this.directory = "ws"
            this.autoLoginLimit = 86400000
            this.websocketPath = ""

            document.title = "npm test"
            this.body = "<h2>fake body</h2>"
        }
    }

    public getWebsocketUrl(): string {
        const isHttps = "https:" === document.location.protocol
        const protocol = isHttps ? "wss:" : "ws:"
        const url = `${protocol}//${this.host}/${this.directory}/${this.websocketPath}`
        return url
    }

    public readonly host: string
    public readonly directory: string
    public readonly websocketPath: string
    public readonly autoLoginLimit

    public readonly body: string
}