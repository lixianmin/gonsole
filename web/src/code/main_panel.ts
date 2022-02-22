/********************************************************************
 created:    2022-01-08
 author:     lixianmin

 Copyright (C) - All Rights Reserved
 *********************************************************************/

import moment from 'moment'

function appendLog(item: Node) {
    const mainPanel = document.getElementById("mainPanel")
    if (mainPanel) {
        let doScroll = mainPanel.scrollTop > mainPanel.scrollHeight - mainPanel.clientHeight - 1
        mainPanel.appendChild(item)

        if (doScroll) {
            mainPanel.scrollTop = mainPanel.scrollHeight - mainPanel.clientHeight
        }
    }
}

export function printHtml(html: string) {
    const item = document.createElement("div")
    item.innerHTML = html
    appendLog(item)
}

export function println() {
    printHtml("<br>");
}

export function printWithTimestamp(html: string) {
    printHtml("[" + moment(new Date()).format("HH:mm:ss.S") + "] " + html);
}