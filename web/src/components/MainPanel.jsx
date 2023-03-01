/********************************************************************
 created:    2023-03-01
 author:     lixianmin

 Copyright (C) - All Rights Reserved
 *********************************************************************/
import {createStore, produce} from "solid-js/store";
import {createEffect, For, Match, onMount, Switch} from "solid-js";
import moment from "moment/moment";

let _mainPanel = null
const [items, setItems] = createStore([])

export function scrollMainPanelToBottom() {
    const mainPanel = _mainPanel
    if (mainPanel) {
        const targetPosition = mainPanel.scrollHeight - mainPanel.clientHeight - 1
        if (mainPanel.scrollTop < targetPosition) {
            mainPanel.scrollTop = targetPosition
        }
    }
}

export function printHtml(html) {
    if (typeof html === 'string' || typeof html === 'function') {
        setItems(produce((state) => {
            state.push({html})
            // todo 这里是每次加数据都调用一次,这个可能不是我想要的
            setTimeout(() => scrollMainPanelToBottom())
        }))
    }
}

export function println() {
    printHtml("<br/>")
}

export function printWithTimestamp(html) {
    const time = moment(new Date()).format("HH:mm:ss.S")
    printHtml(`[${time}] ${html}`)
}

export default function MainPanel(props) {
    onMount(() => {
        _mainPanel = document.getElementById(props.id)
    })

    createEffect(() => {
        // todo 这里好像只执行了一次, 我们需要一个机制统一处理好这个调用
        scrollMainPanelToBottom()
    })

    return <>
        <div id={props.id} ref={props.ref}>
            <For each={items}>{item =>
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