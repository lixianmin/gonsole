/********************************************************************
 created:    2022-01-19
 author:     lixianmin

 Copyright (C) - All Rights Reserved
 *********************************************************************/

export function createWebConfig() {
    // document.title的默认值 ，本地debug的时候使用localhost:8888/ws
    let host = "localhost:8888"
    let directory = "ws"
    let autoLoginLimit = 86400000
    let websocketPath = ""

    let body = "<h2>fake body</h2>"
    if (document.title !== "{{.Data}}") {
        let data = JSON.parse(document.title)
        // console.log("data:", data)

        host = window.location.host
        directory = data.directory
        autoLoginLimit = data.autoLoginLimit
        websocketPath = data.websocketPath

        document.title = data.title
        body = data.body
    }

    return {
        get autoLoginLimit() {
            return autoLoginLimit
        },

        get body() {
            return body
        },

        getRootUrl() {
            let url = `${document.location.protocol}//${host}/${directory}`
            if (url.endsWith("/")) {
                url = url.substring(0, url.length - 1)
            }

            return url
        },

        getWebsocketUrl() {
            const isHttps = "https:" === document.location.protocol
            const protocol = isHttps ? "wss:" : "ws:"
            if (directory !== "") {
                return `${protocol}//${host}/${directory}/${websocketPath}`
            } else {
                return `${protocol}//${host}/${websocketPath}`
            }
        },

        toString() {
            return `host=${host}, directory=${directory}, websocketPath=${websocketPath}, autoLoginLimit=${autoLoginLimit}, body=${body}`
        }
    }
}