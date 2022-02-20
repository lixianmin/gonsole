<script setup lang="ts">
import axios from 'axios';
import {StartX} from "./code/starx";
import {printHtml, println, printWithTimestamp} from "./code/main_panel";
import {sha256} from "js-sha256";
import {History} from "./code/history";
import {WebConfig} from "./code/web_config";

// todo 把auth验证的逻辑提取出来, 并改成安全的逻辑
// todo 修改从golang的template传参到js的逻辑, 不再使用title
// todo dist目录每次npm run build都会被删除重新生成一遍, 怎么解决vendor目录中把dist中的资源包含进来的问题
/**
 * todo 修改传入urlRoot, 从 /ws改为ws, 这样更符合格式化时的习惯
 *
 * todo 列表:
 * 1. 需要在readme中加入npm的开发和使用流程
 * 2. 需要清理旧的代码: 包括各种旧的go与js代码
 * 3. 能够通过npm run dev与真正跑代码
 * 4. 实际应用到一个项目中: build.naked后面的cp逻辑需要把res改到dist
 * 5. 把evt.target.value等逻辑修改为vue应该使用的逻辑
 * 6. 各种js中的any需要调整一下
 * 7. 确认在家里无法修改vendor目录下代码进行调试的原因
 * 8. 打包后生成的assets的根目录是否需要修改
 */

let myHost = `${document.location.host}${document.title}`
let rootUrl = `${document.location.protocol}//${myHost}`

let text = ""
let username = ""
let isAuthorizing = false

let config = new WebConfig()
let history = new History()
let star = new StartX()

axios.get(rootUrl + "/web_config").then((response) => {
  config.loadData(response.data)

  let url = config.getWebsocketUrl(myHost)
  star.connect({url: url}, () => {
    console.log("websocket connected")
  })

  star.on("disconnect", () => {
    printWithTimestamp("<b> disconnected from server </b>")
  })

  document.title = config.getTitle()
  printHtml(config.getBody())
  println()

  star.on("console.html", onHtml)
  star.on("console.default", onDefault)
})

window.onload = () => {
  const inputBox = document.getElementById("inputBox")
  if (!inputBox) {
    return
  }

  inputBox.focus()
  document.onkeydown = function (evt: KeyboardEvent) {
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

function onHtml(obj) {
  printWithTimestamp("<b>server响应：</b>" + obj.data)
  println()
}

function onDefault(obj) {
  const text = JSON.stringify(obj)
  printWithTimestamp("<b>server响应：</b>" + text)
  println()
}

function sendBean(route, msg, callback) {
  const json = JSON.stringify(msg);
  printWithTimestamp("<b>client请求：</b>")
  printHtml(json)
  println()
  star.request(route, msg, callback)
}

function onCommand(obj) {
  switch (obj.op) {
    case "log.list":
      onLogList(obj.data)
      break;
    case "history":
      onHistory(obj.data)
      break;
    case "html":
      onHtml(obj)
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
      const index = parseInt(command.substring(1)) - 1;
      if (!isNaN(index)) {
        command = history.getHistory(index)
      }
    }

    let texts = command.split(/\s+/);  // 支持连续多个空格
    let textsLength = texts.length;
    const name = texts[0];

    if (name === "help") {
      const bean = {
        command: name + " " + rootUrl,
      };

      sendBean("console.command", bean, onCommand)
      history.add(command)
    } else if (textsLength >= 2 && (name === "sub" || name === "unsub")) {
      const bean = {
        topic: texts[1],
      };

      const route = "console." + name;
      sendBean(route, bean, onCommand);
      history.add(command);
    } else if (textsLength >= 2 && name === "auth") {
      username = texts[1];
      isAuthorizing = true
      // $el.type = "password"
      evt.target.type = "password"
      printWithTimestamp(command + "<br/> <h3>请输入密码：</h3><br/>");
      history.add(command);
    } else if (isAuthorizing && textsLength >= 1) {
      isAuthorizing = false
      // this.$el.type = "text"
      evt.target.type = "text"

      const password = name;
      login(username, password);

      if (localStorage) {
        const key = "autoLoginUser";
        const item = {
          username: username,
          password: password,
          expireTime: new Date().getTime() + config.getAutoLoginLimit(),
        }

        const data = JSON.stringify(item)
        localStorage.setItem(key, data)
      }
    } else {
      const bean = {
        command: texts.join(' '),
      }

      sendBean("console.command", bean, onCommand)
      history.add(command)
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
  const key = "hey pet!"
  const digest = sha256.hmac(key, password)

  const bean = {
    command: "auth " + username + " " + digest,
  };

  sendBean("console.command", bean, onCommand)
}

function onHistory(obj) {
  const list = history.getHistories()
  const count = list.length
  const items = new Array(count)

  for (let i = 0; i < count; i++) {
    items[i] = "<li>" + list[i] + "</li>"
  }

  let result = "<b>历史命令列表：</b> <br/> count:&nbsp;" + count + "<br/><ol>" + items.join("") + "</ol>"
  printWithTimestamp(result)
  println()
}

function on_tab(evt) {
  const text = evt.target.value.trim()
  if (text.length > 0) {
    const bean = {
      head: text,
    };

    star.request("console.hint", bean, (obj) => {
      const names = obj.names
      const notes = obj.notes
      const count = names.length
      if (count > 0) {
        evt.target.value = longestCommonPrefix(names)
        if (count > 1) {
          const items = new Array(count)
          for (let i = 0; i < count; i++) {
            items[i] = `<tr> <td>${i + 1}</td> <td>${names[i]}</td> <td>${notes[i]}</td> </tr>`
          }

          const header = "<table> <tr> <th></th> <th>Name</th> <th>Note</th> </tr>"
          const result = header + items.join("") + "</table>"
          printWithTimestamp(result)
          println()
        }
      }
    })
  }
}

function on_up_down(evt) {
  const isArrowUp = evt.key === 'ArrowUp'
  const step = evt.key == 'ArrowUp' ? -1 : 1
  const nextText = history.move(step)
  if (nextText != '') {
    let target = evt.target
    target.value = nextText
    setTimeout(() => {
      let position = nextText.length
      target.setSelectionRange(position, position)
      target.focus()
    }, 0)
  }
}

function longestCommonPrefix(list: string[]): string {
  if (list.length < 2) {
    return list.join()
  }

  let str = list[0];
  for (let i = 1; i < list.length; i++) {
    for (let j = str.length; j > 0; j--) {
      if (str !== list[i].substring(0, j)) str = str.substring(0, j - 1)
      else break
    }
  }

  return str
}

function onLogList(data) {
  const logFiles = data.logFiles
  const fileCount = logFiles.length
  const links = new Array(fileCount)
  let totalSize = 0;
  for (let i = 0; i < fileCount; i++) {
    const fi = logFiles[i];
    totalSize += fi.size;
    let sizeText = getHumanReadableSize(fi.size);
    links[i] = `<tr> <td>${i + 1}</td> <td>${sizeText}</td> <td> <div class="tips"><a href="${rootUrl}/${fi.path}">${fi.path}</a> <span class="tips_text">${fi.sample}</span>
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
    return size + "B"
  }

  if (size < 1048576) {
    return (size / 1024.0).toFixed(1) + "K"
  }

  return (size / 1048576.0).toFixed(1) + "M"
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
