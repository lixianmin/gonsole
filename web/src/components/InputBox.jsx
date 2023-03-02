/********************************************************************
 created:    2023-02-28
 author:     lixianmin

 Copyright (C) - All Rights Reserved
 *********************************************************************/
import {createSignal, onMount} from "solid-js";
import {createDelayed, useKeyDown} from "../code/tools";
import {useHistoryStore} from "../code/use_history_store";

export default function InputBox(props) {
    const [value, setValue] = createSignal('')
    const historyStore = useHistoryStore()

    let inputBox
    onMount(() => {
        inputBox = document.getElementById(props.id)
        inputBox.focus()

        const delayedSetCursor = createDelayed((position)=>{
            inputBox.setSelectionRange(position, position)
        })

        useKeyDown(inputBox, ['ArrowUp', 'ArrowDown'], evt => {
            const step = evt.key === 'ArrowUp' ? -1 : 1
            const nextText = historyStore.move(step)

            // 按bash中history的操作习惯, 如果是arrow down的话, 最后一个应该是""
            if (nextText !== '' || step === 1) {
                inputBox.value = nextText
                delayedSetCursor(nextText.length)
            }

            evt.preventDefault()
        })
    })

    function onInput(evt) {
        setValue(evt.target.value)
    }

    // props.ref是ref转发，但也因此导致没法定义一个inputBox直接使用了，没办法，这里使用document.getElementById()自己找一个
    return <>
        <input id={props.id} ref={props.ref} value={value()} onInput={onInput}
               placeholder="Tab补全命令, Enter执行命令"/>
    </>
}