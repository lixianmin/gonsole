<template>
  <b>历史命令列表：</b> <br/> count: &nbsp; {{ count }}
  <ol id="history-with-index">
    <li v-for="history in histories" :key="history">
      {{ history }}
    </li>
  </ol>
  <br>
</template>

<script setup>
import {onMounted, toRaw} from "vue"
import {useHistoryStore} from "@/code/use_history_store"
import {scrollMainPanelToBottom} from "@/code/main_panel"

const historyStore = useHistoryStore()
// 把所有的数据都复制一份, 这是为了防止history的数据变化的时候, 这边再收到通知
const histories = toRaw(historyStore.getHistoryList())
const count = historyStore.getHistoryCount()

onMounted(() => {
  scrollMainPanelToBottom()
})

</script>