/********************************************************************
 created:    2023-03-01
 author:     lixianmin

 Copyright (C) - All Rights Reserved
 *********************************************************************/
import {createStore, produce} from "solid-js/store";
import {createEffect, For, Match, onMount, Switch} from "solid-js";
import moment from "moment/moment";
import {createDelayed} from "../code/tools";

let _mainPanel = null
const [items, setItems] = createStore([])

function scrollMainPanelToBottom() {
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

    const delayedScrollMainPanelToBottom = createDelayed(()=>{
        scrollMainPanelToBottom()
    }, 50)

    // 如果监控items, 则只执行一次; 如果监控items.length, 则可以每次在push后都执行
    createEffect(() => {
        delayedScrollMainPanelToBottom(items.length)
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