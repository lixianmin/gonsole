'use strict'
/********************************************************************
 created:    2023-03-01
 author:     lixianmin

 Copyright (C) - All Rights Reserved
 *********************************************************************/
import {createStore, produce} from "solid-js/store";
import {createEffect, For, Match, Switch} from "solid-js";
import moment from "moment/moment";
import {createDelayed} from "../code/tools";

// 把createStore()定义在外面, 从而支持export一些方法用于操作store的数据
const [storePanel, setStorePanel] = createStore({
    widgets: [],
})

let scrollMainPanelToBottom

export function changeWidget(widget) {
    setStorePanel(produce((state) => {
        state.widgets[widget.index] = {...widget}
        // 这种修改可能需要导致窗口滚动
        scrollMainPanelToBottom()
    }))
}

export function printHtml(html) {
    if (typeof html === 'string' || typeof html === 'function') {
        const widget = {html}
        setStorePanel(produce((state) => {
            widget.index = state.widgets.length
            state.widgets.push(widget)
        }))

        return widget
    } else {
        console.warn(`invalid html type, html=${html}`)
    }
}

export function println() {
    return printHtml("<br/>")
}

export function printWithTimestamp(html) {
    const time = moment(new Date()).format("HH:mm:ss.S")
    return printHtml(`[${time}] ${html}`)
}

export default function MainPanel() {
    let mainPanel

    scrollMainPanelToBottom = function () {
        const targetPosition = mainPanel.scrollHeight - mainPanel.clientHeight - 1
        if (mainPanel.scrollTop < targetPosition) {
            mainPanel.scrollTop = targetPosition
        }
    }

    const delayedScrollMainPanelToBottom = createDelayed(() => {
        scrollMainPanelToBottom()
    })

    // 如果监控widgets, 则只执行一次; 如果监控widgets.length, 则可以每次在push后都执行
    createEffect(() => {
        delayedScrollMainPanelToBottom(storePanel.widgets.length)
    })

    return <>
        <div id='mainPanel' ref={mainPanel}>
            <For each={storePanel.widgets}>{widget =>
                <Switch>
                    <Match when={typeof widget.html === 'string'}>
                        <div innerHTML={widget.html}/>
                    </Match>
                    <Match when={typeof widget.html === 'function'}>
                        <div>
                            {widget.html()}
                        </div>
                    </Match>
                </Switch>
            }</For>
        </div>
    </>
}