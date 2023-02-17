'use strict'

/********************************************************************
 created:    2022-01-19
 author:     lixianmin

 Copyright (C) - All Rights Reserved
 *********************************************************************/

export function createWebConfig() {
    // document.title的默认值 ，本地debug的时候使用localhost:8888/ws
    let _host = "localhost:8888"
    let _directory = "ws"
    let _websocketPath = ""

    let _body = "<h2>fake body</h2>"
    if (document.title !== "{{.Data}}") {
        let data = JSON.parse(document.title)
        // console.log("data:", data)

        _host = window.location.host
        _directory = data.directory
        _websocketPath = data.websocketPath

        document.title = data.title
        _body = data.body
    }

    return {
        get body() {
            return _body
        },

        getRootUrl() {
            let url = `${document.location.protocol}//${_host}/${_directory}`
            if (url.endsWith("/")) {
                url = url.substring(0, url.length - 1)
            }

            return url
        },

        getWebsocketUrl() {
            const isHttps = "https:" === document.location.protocol
            const protocol = isHttps ? "wss:" : "ws:"
            if (_directory !== "") {
                return `${protocol}//${_host}/${_directory}/${_websocketPath}`
            } else {
                return `${protocol}//${_host}/${_websocketPath}`
            }
        },

        toString() {
            return `host=${_host}, directory=${_directory}, websocketPath=${_websocketPath}, autoLoginLimit=${_autoLoginLimit}, body=${_body}`
        }
    }
}