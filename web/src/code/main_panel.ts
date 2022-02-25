/********************************************************************
 created:    2022-01-08
 author:     lixianmin

 Copyright (C) - All Rights Reserved
 *********************************************************************/

import moment from 'moment'

function appendLog(item: Node) {
    const mainPanel = document.getElementById("mainPanel")
    if (mainPanel) {
        let needScroll = mainPanel.scrollTop < mainPanel.scrollHeight - mainPanel.clientHeight - 1
        mainPanel.appendChild(item)

        if (needScroll) {
            mainPanel.scrollTop = mainPanel.scrollHeight - mainPanel.clientHeight - 1
        }
    }
}

export function printHtml(html: string): HTMLDivElement {
    const item = document.createElement("div")
    item.innerHTML = html
    appendLog(item)

    return item
}

// todo 这个方法也许可以优化, 不应该每次都生成一个<div>吧?
export function println() {
    printHtml("<br>")
}

export function printWithTimestamp(html: string) {
    printHtml("[" + moment(new Date()).format("HH:mm:ss.S") + "] " + html);
}