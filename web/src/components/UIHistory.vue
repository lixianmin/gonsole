<template>
  <b>历史命令列表：</b> <br/> count: &nbsp; {{ count }}
  <ol id="history-with-index">
    <li v-for="history in histories" :key="history">
      {{ history }}
    </li>
  </ol>
  <br>
</template>

<script setup lang="ts">
import {onMounted} from "vue";
import {useHistoryStore} from "../code/use_history_store";
import {scrollMainPanelToBottom} from "../code/main_panel";

const store = useHistoryStore()
// 把所有的数据都复制一份, 这是为了防止history的数据变化的时候, 这边再收到通知
const histories = store.histories.map(item => item)
const count = store.count

onMounted(() => {
  scrollMainPanelToBottom()
})

</script>