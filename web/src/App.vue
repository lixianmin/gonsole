<script setup lang="ts">
import axios from 'axios'
import StartX from "./code/starx";
import {printHtml, println, printWithTimestamp} from "./code/main_panel";
import {sha256} from "js-sha256";

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
let history :any[]= []
let starx = new StartX()

axios.get(url).then((response) => {
  const config = new WebConfig(response.data)
  let url = config.getWebsocketUrl()
  starx.connect({url: url}, () => {
    console.log("websocket connected")
  })

  starx.on("disconnect", () => {
    printWithTimestamp("<b> disconnected from server </b>")
  })

  printHtml(response.data.body)
  println()

  starx.on("console.html", onHtml)
  starx.on("console.default", onDefault)
})

window.onload = function () {
  const inputBox = document.getElementById("inputBox")
  document.onkeydown = function (evt) {
    if (evt.key === 'Enter') {
      let control = document.activeElement;
      if (control !== inputBox && inputBox) {
        inputBox.focus()
        // return false的意思是：这个按键事件本js处理了，不再传播这个事件。
        // 默认情况下会继续传播按键事件，Enter会导致页面refresh
        return false
      }
    }
  }
}

if (localStorage) {
  const key = "history"
  const item = localStorage.getItem(key)
  if (item){
    const json = JSON.parse(item)
    if (json) {
      history = json
      historyIndex = history.length; // 初始大小
    }
  }

  // 在unload时将history存储到localStorage中
  window.onunload = evt =>{
    const key = "history"
    localStorage.setItem(key, JSON.stringify(history.slice(-100)))
  }
}

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
  // console.log("onCommand -->", obj.op)
  switch (obj.op) {
    case "log.list":
      onLogList(obj.data)
      break;
      case "history":
        onHistory(obj.data);
        break;
      case "html":
        onHtml(obj);
        break;
      case "empty":
        break;
      default:
        onDefault(obj)
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
      const host = document.location.protocol + "//" + myHost
      const bean = {
        command: name + " " + host,
      };

      sendBean("console.command", bean, onCommand)
      addHistory(command)
    } else if (textsLength >= 2 && (name === "sub" || name === "unsub")) {
      const bean = {
        topic: texts[1],
      };

      const route = "console." + name;
      sendBean(route, bean, onCommand);
      addHistory(command);
    } else if (textsLength >= 2 && name === "auth") {
      username = texts[1];
      isAuthorizing = true;
      // $el.type = "password";
      printWithTimestamp(command + "<br/> <h3>请输入密码：</h3><br/>");
      addHistory(command);
    } else if (isAuthorizing && textsLength >= 1) {
      // this.isAuthorizing = false;
      // this.$el.type = "text";
      //
      const password = name;
      login(username, password);

      if (localStorage) {
        const key = "autoLoginUser";
        const item = {
          username: username,
          password: password,
          expireTime: new Date().getTime() //+ {{.AutoLoginLimit}},
      }

        const data = JSON.stringify(item);
        localStorage.setItem(key, data);
      }
    } else {
      const bean = {
        command: texts.join(' '),
      };

      sendBean("console.command", bean, onCommand)
      addHistory(command)
    }
  } else {
    printWithTimestamp('')
  }

  const mainPanel = document.getElementById("mainPanel")
  if (mainPanel) {
    mainPanel.scrollTop = mainPanel.scrollHeight - mainPanel.clientHeight // 其实在shell中只要有输入就会滚屏
  }
}

function login(username, password) {
  const key = "hey pet!";
  const digest = sha256.hmac(key, password);

  const bean = {
    command: "auth " + username + " " + digest,
  };

  sendBean("console.command", bean, onCommand);
}

function addHistory(command) {
  const size = history.length;
  // 如果history中存储的最后一条与command不一样，则将command加入到history列表。否则将historyIndex调整到最后
  if (size === 0 || history[size - 1] !== command) {
    historyIndex = history.push(command)
  } else { // addHistory()都是在输入命令时才调用的，这时万一historyIndex处于history数组的中间位置，将其调整到最后
    historyIndex = history.length;
  }
}

function onHistory(obj) {
  const count = history.length;
  const items = new Array(count);
  for (let i = 0; i < count; i++) {
    items[i] = "<li>" + history[i] + "</li>";
  }

  let result = "<b>历史命令列表：</b> <br/> count:&nbsp;" + count + "<br/><ol>" + items.join("") + "</ol>"
  printWithTimestamp(result);
  println();
}

function on_tab(evt) {
  const text = evt.target.value.trim();
  if (text.length > 0) {
    const bean = {
      head: text,
    };

    starx.request("console.hint", bean, function (obj) {
      const names = obj.names;
      const notes = obj.notes;
      const count = names.length;
      if (count > 0) {
        evt.target.value = longestCommonPrefix(names);
        if (count > 1) {
          const items = new Array(count);
          for (let i = 0; i < count; i++) {
            items[i] = `<tr> <td>${i + 1}</td> <td>${names[i]}</td> <td>${notes[i]}</td> </tr>`;
          }

          const header = "<table> <tr> <th></th> <th>Name</th> <th>Note</th> </tr>";
          const result = header + items.join("") + "</table>";
          printWithTimestamp(result);
          println();
        }
      }
    })
  }
}

function on_up_down(evt) {
  const isArrowUp = evt.key === 'ArrowUp'
  let isChanged = false
  let index = historyIndex
  if (isArrowUp && index > 0) {
    index -= 1
    isChanged = true
  } else if (!isArrowUp && index+1 <history.length) {
    index += 1
    isChanged = true
  }

  if (isChanged) {
    historyIndex = index
    text = index < history.length ? history[index] : ''
    setTimeout(function () {
      let position = text.length
      // that.$el.setSelectionRange(position, position)
      // that.$el.focus()
    }, 0)
  }
}

function longestCommonPrefix(list) {
  if (list.length < 2) {
    return list.join()
  }

  let str = list[0];
  for (let i = 1; i < list.length; i++) {
    for (let j = str.length; j > 0; j--) {
      if (str !== list[i].substring(0, j)) str = str.substring(0, j - 1);
      else break
    }
  }

  return str
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
