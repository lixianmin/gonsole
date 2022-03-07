<template>
  <b>日志文件列表：</b><br>
  count: &nbsp; {{ count }} <br>
  total: &nbsp; {{ totalSize }} <br>

  <ol id="log-list-with-index">
    <li v-for="fi in props.logFiles" :key="fi">
      {{ fi }}
    </li>
  </ol>
  <br>
</template>

<script setup lang="ts">

import {computed, onMounted} from "vue";
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
}

// https://blog.ninja-squad.com/2021/09/30/script-setup-syntax-in-vue-3/
const props = defineProps<Props>()
const count = computed(() => props.logFiles.length)
const totalSize = computed(() => getHumanReadableSize(props.logFiles.map(fi => fi.size).reduce((last, current) => last + current)))

// const logFiles = data.logFiles
// const fileCount: number = logFiles.length
// const links = new Array(fileCount)
// let totalSize = 0;
// for (let i = 0; i < fileCount; i++) {
//   const fi = logFiles[i]
//   totalSize += fi.size
//   let sizeText = getHumanReadableSize(fi.size)
//   links[i] = `<tr> <td>${i + 1}</td> <td>${sizeText}</td> <td> <div class="tips"><a href="${rootUrl}/${fi.path}">${fi.path}</a> <span class="tips_text">${fi.sample}</span>
//                                 <input type="button" class="copy_button" onclick="copyToClipboard('${fi.path}')" value="复制"/>
//                                 </div></td> <td>${fi.mod_time}</td> </tr>`
// }
//
// let result = "<b>日志文件列表：</b> <br> count:&nbsp;" + fileCount + "<br>total:&nbsp;&nbsp;" + getHumanReadableSize(totalSize) + "<br>"
// result += "<table> <tr> <th></th> <th>Size</th> <th>Name</th> <th>Modified Time</th> </tr>" + links.join("") + "</table>"
// printWithTimestamp(result)
// println()

onMounted(() => {
      scrollMainPanelToBottom()
    }
)

</script>