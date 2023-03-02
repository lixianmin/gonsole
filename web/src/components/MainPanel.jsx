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
    items: []
})

export function printHtml(html) {
    if (typeof html === 'string' || typeof html === 'function') {
        setStorePanel(produce((state) => {
            state.items.push({html})
        }))
    } else {
        console.warn(`invalid html type, html=${html}`)
    }
}

export function println() {
    printHtml("<br/>")
}

export function printWithTimestamp(html) {
    const time = moment(new Date()).format("HH:mm:ss.S")
    printHtml(`[${time}] ${html}`)
}

export default function MainPanel() {
    let mainPanel

    function scrollMainPanelToBottom() {
        const targetPosition = mainPanel.scrollHeight - mainPanel.clientHeight - 1
        if (mainPanel.scrollTop < targetPosition) {
            mainPanel.scrollTop = targetPosition
        }
    }

    const delayedScrollMainPanelToBottom = createDelayed(() => {
        scrollMainPanelToBottom()
    })

    // 如果监控items, 则只执行一次; 如果监控items.length, 则可以每次在push后都执行
    createEffect(() => {
        delayedScrollMainPanelToBottom(storePanel.items.length)
    })

    return <>
        <div id='mainPanel' ref={mainPanel}>
            <For each={storePanel.items}>{item =>
                <Switch>
                    <Match when={typeof item.html === 'string'}>
                        <div innerHTML={item.html}/>
                    </Match>
                    <Match when={typeof item.html === 'function'}>
                        <div>
                            {item.html()}
                        </div>
                    </Match>
                </Switch>
            }</For>
        </div>
    </>
}