/********************************************************************
 created:    2022-01-08
 author:     lixianmin

 Copyright (C) - All Rights Reserved
 *********************************************************************/

import moment from 'moment'

let _mainPanel: HTMLElement | null = null

function getMainPanel(): HTMLElement | null {
    if (_mainPanel == null) {
        _mainPanel = document.getElementById("mainPanel")
    }

    return _mainPanel
}

function appendLog(item: Node) {
    const mainPanel = getMainPanel()
    if (mainPanel) {
        mainPanel.appendChild(item)
        scrollMainPanelToBottom()
    }
}

export function scrollMainPanelToBottom() {
    const mainPanel = getMainPanel()
    if (mainPanel) {
        const targetPosition = mainPanel.scrollHeight - mainPanel.clientHeight - 1
        if (mainPanel.scrollTop < targetPosition) {
            mainPanel.scrollTop = targetPosition
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