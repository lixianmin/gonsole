import {useHistoryStore} from "../code/use_history_store";
import {For} from "solid-js";

export default function History() {
    const historyStore = useHistoryStore()
    return <>
        <b>历史命令列表：</b> <br/> count: &nbsp; {historyStore.getHistoryCount()}
        <ol>
            <For each={historyStore.getHistoryList()}>{
                history =>
                    <li>{history}</li>
            }</For>
        </ol>
        <br/>
    </>
}