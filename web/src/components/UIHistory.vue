<template>
  <b>历史命令列表：</b> <br/> count: &nbsp; {{ historyStore.count }}
  <ol id="history-with-index">
    <li v-for="history in historyStore.histories" :key="history">
      {{ history }}
    </li>
  </ol>
  <br>
</template>

<script setup lang="ts">
import {onMounted} from "vue";
import {useHistoryStore} from "../code/use_history_store";
import {scrollMainPanelToBottom} from "../code/main_panel";

const historyStore = useHistoryStore()

onMounted(() => {
  scrollMainPanelToBottom()
})

</script>