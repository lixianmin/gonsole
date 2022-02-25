<template>
  <b>历史命令列表：</b> <br/> count: &nbsp; {{ history.count }}
  <ol id="history-with-index">
    <li v-for="history in history.histories" :key="history">
      {{ history }}
    </li>
  </ol>
  <br>
</template>

<script setup lang="ts">
import {onMounted, ref} from "vue";
import {useHistoryStore} from "../code/use_history_store";

let history = useHistoryStore()

onMounted(() => {
  const mainPanel = document.getElementById("mainPanel")
  if (mainPanel) {
    mainPanel.scrollTop = mainPanel.scrollHeight - mainPanel.clientHeight
  }
})

</script>