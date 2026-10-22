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

export function changeWidget(widget) {
    setStorePanel(produce((state) => {
        state.widgets[widget.index] = {...widget}
        // 修正: 之前在这里直接调用模块级的滚动函数, 而该函数是MainPanel渲染时才赋值的全局可变变量,
        // 违反了"无全局可变状态"的约定. 现在改由MainPanel中的effect负责滚动
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

    const scrollMainPanelToBottom = function () {
        const targetPosition = mainPanel.scrollHeight - mainPanel.clientHeight - 1
        if (mainPanel.scrollTop < targetPosition) {
            mainPanel.scrollTop = targetPosition
        }
    }

    const delayedScrollMainPanelToBottom = createDelayed(() => {
        scrollMainPanelToBottom()
    })

    // 监控widgets的内容, 无论是新增还是修改(changeWidget)都会滚动到底部
    // 修正: 之前滚动函数是模块级的可变变量, 由MainPanel渲染时赋值, 违反了"无全局可变状态"的约定
    createEffect(() => {
        // 读取每个widget的html, 建立effect与widgets内容之间的依赖
        for (const widget of storePanel.widgets) {
            void widget.html
        }

        delayedScrollMainPanelToBottom()
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