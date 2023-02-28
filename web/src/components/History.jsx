import {useHistoryStore} from "../code/use_history_store";
import {For, onMount} from "solid-js";
import {scrollMainPanelToBottom} from "../code/main_panel";

export default function History() {
    onMount(() => {
        scrollMainPanelToBottom()
    })

    const historyStore = useHistoryStore()
    return (
        <>
            <b>历史命令列表：</b> <br/> count: &nbsp; {historyStore.getHistoryCount()}
            <ol>
                <For each={historyStore.getHistoryList()}>{
                    history => (
                        <li>{history}</li>
                    )
                }</For>
            </ol>
            <br/>
        </>
    )
}