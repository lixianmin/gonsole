'use strict'

/********************************************************************
 created:    2022-01-08
 author:     lixianmin

 Copyright (C) - All Rights Reserved
 *********************************************************************/

import moment from 'moment'

let _mainPanel = null

function getMainPanel() {
    if (_mainPanel == null) {
        _mainPanel = document.getElementById("mainPanel")
    }

    return _mainPanel
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

export function printHtml(html) {
    if (typeof html !== 'string') {
        return ''
    }

    const item = document.createElement("div")
    item.innerHTML = html

    const mainPanel = getMainPanel()
    if (mainPanel) {
        mainPanel.appendChild(item)
        scrollMainPanelToBottom()
    }

    return item
}

// todo 这个方法也许可以优化, 不应该每次都生成一个<div>吧?
export function println() {
    printHtml("<br>")
}

export function printWithTimestamp(html) {
    const time = moment(new Date()).format("HH:mm:ss.S")
    printHtml(`[${time}] ${html}`)
}