/********************************************************************
 created:    2022-01-19
 author:     lixianmin

 Copyright (C) - All Rights Reserved
 *********************************************************************/

export class WebConfig {
    public constructor() {
        if (document.title != "{{.Data}}") {
            let data = JSON.parse(document.title)
            // console.log("data:", data)

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

            this.body = "<h2>fake body</h2>"
            // console.log(`config=[${this}]`)
        }
    }

    public getRootUrl(): string {
        let url = `${document.location.protocol}//${this.host}/${this.directory}`
        if (url.endsWith("/")) {
            url = url.substring(0, url.length - 1)
        }

        return url
    }

    public getWebsocketUrl(): string {
        const isHttps = "https:" === document.location.protocol
        const protocol = isHttps ? "wss:" : "ws:"
        if (this.directory != "") {
            return `${protocol}//${this.host}/${this.directory}/${this.websocketPath}`
        } else {
            return `${protocol}//${this.host}/${this.websocketPath}`
        }
    }

    public toString(): string {
        return `host=${this.host}, directory=${this.directory}, websocketPath=${this.websocketPath}, autoLoginLimit=${this.autoLoginLimit}, body=${this.body}`
    }

    public readonly host: string
    public readonly directory: string
    public readonly websocketPath: string
    public readonly autoLoginLimit

    public readonly body: string
}