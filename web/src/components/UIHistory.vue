<template>
  <b>历史命令列表：</b> <br/> count: &nbsp; {{ histories.count }}
  <ol id="history-with-index">
    <li v-for="history in histories.getHistories()" :key="history">
      {{ history }}
    </li>
  </ol>
  <br>
</template>

<script setup lang="ts">
import {History} from "../code/history";
import {onMounted, ref} from "vue";

let histories = ref(new History())

onMounted(() => {
  const mainPanel = document.getElementById("mainPanel")
  if (mainPanel) {
    mainPanel.scrollTop = mainPanel.scrollHeight - mainPanel.clientHeight
  }
})

</script>