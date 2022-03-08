<template>
  <b>日志文件列表：</b><br>
  count: &nbsp; {{ count }} <br>
  total: &nbsp; {{ totalSize }} <br>
  <br>
  <table>
    <tr>
      <th></th>
      <th>Size</th>
      <th>Name</th>
      <th>Modified Time</th>
    </tr>
    <tbody>
    <tr v-for="(fi, index) in logFiles" :key="fi.path">
      <td>{{ index + 1 }}</td>
      <td>{{ getHumanReadableSize(fi.size) }}</td>
      <td>
        <div v-html="fetchNameHtml(fi)" class="tips"></div>
      </td>
      <td>{{ fi.mod_time }}</td>
    </tr>
    </tbody>
  </table>
  <br>
</template>

<script setup lang="ts">

import {computed, nextTick} from "vue";
import {scrollMainPanelToBottom} from "../code/main_panel";
import {getHumanReadableSize} from "../code/tools";

interface FileInfo {
  size: number
  path: string
  mod_time: string
  sample: string
}

interface Props {
  logFiles: FileInfo[]
  rootUrl: string
}

// https://blog.ninja-squad.com/2021/09/30/script-setup-syntax-in-vue-3/
const props = defineProps<Props>()
const logFiles = props.logFiles

const count = computed(() => logFiles.length)
const totalSize = computed(() => getHumanReadableSize(logFiles.map(fi => fi.size).reduce((last, current) => last + current)))

function fetchNameHtml(fi: FileInfo): string {
  return `<a href="${props.rootUrl}/${fi.path}">${fi.path}</a>
          <span class="tips_text">${fi.sample}</span>
          <input type="button" class="copy_button" onclick="copyToClipboard('${fi.path}')" value="复制"/>`
}

nextTick(() => scrollMainPanelToBottom())
</script>