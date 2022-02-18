<script setup lang="ts">
// This starter template is using Vue 3 <script setup> SFCs
// Check out https://v3.vuejs.org/api/sfc-script-setup.html#sfc-script-setup
import HelloWorld from './components/HelloWorld.vue'
import axios from 'axios'
import StartX from "./lib/starx";
import {printHtml, println, printWithTimestamp} from "./lib/main_panel";

class WebConfig {
  public constructor(config) {
    // console.log(config)
    this.autoLoginLimit = config.autoLoginLimit
    this.websocketPath = config.websocketPath
  }

  public getWebsocketUrl(): string {
    const isHttps = "https:" === document.location.protocol
    const protocol = isHttps ? "wss://" : "ws://"
    const url = `${protocol}${myHost}/${this.websocketPath}`
    return url
  }

  public readonly autoLoginLimit: number
  private readonly websocketPath: string
}

let myHost = "localhost:8888/ws"
let url = `${document.location.protocol}//${myHost}/web_config`
let text = ""
let username = ""
let isAuthorizing = false
let historyIndex = -1
let history = []
let starx = new StartX()

axios.get(url).then((response) => {
  const config = new WebConfig(response.data)
  let url = config.getWebsocketUrl()
  starx.connect({url: url}, () => {
    console.log("star initialized")
  })

  starx.on("disconnect", () => {
    printWithTimestamp("<b> disconnected from server </b>")
  })

  printHtml(response.data.body)
  println()

  starx.on("console.html", onHtml)
  starx.on("console.default", onDefault)
})

function onHtml(obj) {
  printWithTimestamp("<b>server响应：</b>" + obj.data)
  println()
}

function onDefault(obj) {
  const text = JSON.stringify(obj)
  printWithTimestamp("<b>server响应：</b>" + text)
  println();
}

function sendBean(route, msg, callback) {
  const json = JSON.stringify(msg);
  printWithTimestamp("<b>client请求：</b>")
  printHtml(json)
  println()
  starx.request(route, msg, callback)
}

function onCommand(obj) {
  switch (obj.op) {
    case "log.list":
      onLogList(obj.data)
      break;
      // case "history":
      //   this.onHistory(obj.data);
      //   break;
      // case "html":
      //   this.onHtml(obj);
      //   break;
      // case "empty":
      //   break;
      // default:
      //   onDefault(obj);
  }
}

function on_enter(evt) {
  let command = evt.target.value.trim()
  if (command !== "") {
    evt.target.value = ""

    // 检查是不是调用history命令
    if (command.startsWith("!")) {
      const index = parseInt(command.substr(1)) - 1;
      if (!isNaN(index) && index >= 0 && index < history.length) {
        command = history[index];
      }
    }

    let texts = command.split(/\s+/);  // 支持连续多个空格
    let textsLength = texts.length;
    const name = texts[0];

    if (name === 'help') {
      const bean = {
        command: name,
      }

      sendBean("console.command", bean, onCommand)
      // this.addHistory(command);
    } else if (textsLength >= 2 && (name === "sub" || name === "unsub")) {
      // const bean = {
      //   topic: texts[1],
      // };
      //
      // const route = "console." + name;
      // this.sendBean(route, bean, this.onCommand);
      // this.addHistory(command);
    } else if (textsLength >= 2 && name === "auth") {
      // this.username = texts[1];
      // this.isAuthorizing = true;
      // this.$el.type = "password";
      // this.printWithTimestamp(command + "<br/> <h3>请输入密码：</h3><br/>");
      // this.addHistory(command);
    } else if (isAuthorizing && textsLength >= 1) {
      // this.isAuthorizing = false;
      // this.$el.type = "text";
      //
      // const password = name;
      // this.login(this.username, password);
      //
      // if (localStorage) {
      //   const key = "autoLoginUser";
      //   const item = {
      //     username: this.username,
      //     password: password,
      //     expireTime: new Date().getTime() + {{.AutoLoginLimit}},
      // }
      //
      //   const data = JSON.stringify(item);
      //   localStorage.setItem(key, data);
      // }
    } else {
      // const bean = {
      //   command: texts.join(' '),
      // };
      //
      // this.sendBean("console.command", bean, this.onCommand);
      // this.addHistory(command);
    }
  } else {
    printWithTimestamp('')
  }

  // const mainPanel = document.getElementById("mainPanel");
  // mainPanel.scrollTop = mainPanel.scrollHeight - mainPanel.clientHeight; // 其实在shell中只要有输入就会滚屏
}

function on_tab() {

}

function on_up_down() {
  let command = text.trim()
  console.log("command=", command, ", text=", text)
}

function onLogList(data) {
  const host = document.location.protocol + "//" + myHost;
  const logFiles = data.logFiles;
  const fileCount = logFiles.length;
  const links = new Array(fileCount);
  let totalSize = 0;
  for (let i = 0; i < fileCount; i++) {
    const fi = logFiles[i];
    totalSize += fi.size;
    let sizeText = getHumanReadableSize(fi.size);
    links[i] = `<tr> <td>${i + 1}</td> <td>${sizeText}</td> <td> <div class="tips"><a href="${host}/${fi.path}">${fi.path}</a> <span class="tips_text">${fi.sample}</span>
                                <input type="button" class="copy_button" onclick="copyToClipboard('${fi.path}')" value="复制"/>
                                </div></td> <td>${fi.mod_time}</td> </tr>`;
  }

  let result = "<b>日志文件列表：</b> <br> count:&nbsp;" + fileCount + "<br>total:&nbsp;&nbsp;" + getHumanReadableSize(totalSize) + "<br>";
  result += "<table> <tr> <th></th> <th>Size</th> <th>Name</th> <th>Modified Time</th> </tr>" + links.join("") + "</table>";
  printWithTimestamp(result)
  println()
}

function getHumanReadableSize(size) {
  if (size < 1024) {
    return size + "B";
  }

  if (size < 1048576) {
    return (size / 1024.0).toFixed(1) + "K";
  }

  return (size / 1048576.0).toFixed(1) + "M";
}


</script>

<template>
  <div id="mainPanel"></div>
  <div id="inputBoxDiv">
    <input id="inputBox" ref="mainPanel" v-model="text" placeholder="Tab补全命令, Enter执行命令"
           @keydown.enter.prevent="on_enter"
           @keydown.tab.prevent="on_tab"
           @keydown.up.down.prevent="on_up_down"
    />
  </div>
</template>

<style>
/*http://thomasf.github.io/solarized-css/*/
html {background-color: #002b36;color: #839496;margin: 1em;font-size: 1.2em;}
.copy_button { background-color: #008CBA; border: none; color: white; }

a {color: #b58900;}
a:visited {color: #cb4b16;}
a:hover {color: #cb4b16;}

table { border-width: 1px; border-color: #729ea5;border-collapse: collapse;}
th { background-color:#004949; border-width: 1px;padding: 8px;border-style: solid;border-color: #729ea5;text-align:left;}
th:hover { cursor: pointer;}
th:after { content: attr(data-text); font-size: small; margin-left: 5px;}
td { border-width: 1px;padding: 8px;border-style: solid;border-color: #729ea5;}

/*https://www.runoob.com/css/css-tooltip.html*/
.tips { position: relative; display: inline-block; border-bottom: 1px dotted black; }

.tips .tips_text {
  visibility: hidden; display: inline-block; white-space: nowrap; background: #005959; border-radius: 6px; padding: 6px 6px;
  /* 定位 */
  position: absolute; z-index: 1; top: -5px;left: 105%;
}

.tips:hover .tips_text { visibility: visible; }

#mainPanel {margin: 0;padding: 0.5em 0.5em 0.5em 0.5em;position: absolute;top: 0.5em;left: 0.5em;right: 0.5em;bottom: 3em;overflow: auto;}
#inputBoxDiv {padding: 0 0.5em 0 0.5em;margin: 0;position: absolute;bottom: 1em;left: 1px;width: 100%;overflow: hidden;}
#inputBox {width:100%;height:1.6em;font-size:1.5em; background-color: #073642; color: #859900}
</style>
