/********************************************************************
 created:    2023-02-27
 author:     lixianmin

 Copyright (C) - All Rights Reserved
 *********************************************************************/

import './App.module.css';

import {StartX} from "./code/starx";
import {printHtml, println, printWithTimestamp} from "./code/main_panel.js";
import {createWebConfig} from "./code/web_config.js";
import {createLogin} from "./code/login";
import {useHistoryStore} from "./code/use_history_store.js";
import moment from "moment";
import {longestCommonPrefix} from "./code/tools";
import History from "./components/History";
import {render} from "solid-js/web";
import JsonTable from "./components/JsonTable";
import LogList from "./components/LogList";

// todo 修改从golang的template传参到js的逻辑, 不再使用title
// todo disconnected from server的时候, 写一个online time
/**
 * todo 需要在readme中加入npm的开发和使用流程
 * todo 把evt.target.value等逻辑修改为vue应该使用的逻辑
 * todo 各种js中的any需要调整一下
 * todo 打包后生成的assets的根目录是否需要修改
 */

const App = () => {

    let inputText
    let username = ""
    let isAuthorizing = false

    let config = createWebConfig()
    const historyStore = useHistoryStore()

    let star = new StartX()
    let rootUrl = config.getRootUrl()

    // 开放sendCommand方法, 使client端写js代码的时候用websocket跟server交互
    window.sendCommand = sendCommand

    let login = createLogin((cmd, username, digestOrToken, fingerprint) => {
        printWithTimestamp("<b>client请求：</b>")
        printHtml(`${cmd} ${username} [digest | token] fingerprint`)
        println()

        const bean = {command: `${cmd} ${username} ${digestOrToken} ${fingerprint}`}
        return new Promise(resolve => {
            // 把callback改为promise
            // @ts-ignore
            star.request("console.command", bean, obj => {
                const cloned = {...obj.data}  // shadow clone
                resolve(obj.data)

                delete cloned.token
                const text = JSON.stringify(cloned)
                printWithTimestamp("<b>server响应：</b>" + text)
                println()
            })
        })
    })

    star.connect({url: config.getWebsocketUrl()}, (nonce) => {
        console.log("websocket connected")
        printHtml(config.body)
        println()
        login.tryAutoLogin(nonce).then()
    })

    const uptime = new Date()
    star.on("disconnect", () => {
        const onlineTime = moment.duration(new Date().getTime() - uptime.getTime(), "milliseconds").humanize()
        printWithTimestamp(`<b> disconnected from server after ${onlineTime} </b>`)
    })


    star.on("console.html", onHtml)
    star.on("console.default", onDefault)

    window.onload = () => {
        const inputBox = document.getElementById("inputBox")
        if (!inputBox) {
            return
        }

        inputBox.focus()
        document.onKeyDown = function (evt) {
            if (evt.key === 'Enter') {
                let control = document.activeElement;
                if (control !== inputBox && inputBox) {
                    inputBox.focus()
                    // return false的意思是：这个按键事件当前代码处理了，不再bubble上传这个事件。
                    // 默认情况下会继续传播按键事件，Enter会导致页面refresh
                    return false
                }
            }
        }
    }

    function onHtml(data) {
        printWithTimestamp("<b>server响应：</b>" + data)
        println()
    }

    function onTable(data) {
        render(() => <JsonTable tableData={data}/>, printHtml(''))
    }

    function onDefault(operation) {
        const text = JSON.stringify(operation)
        printWithTimestamp("<b>server响应：</b>" + text)
        println()
    }

    function sendBean(route, bean, callback) {
        const json = JSON.stringify(bean)
        printWithTimestamp("<b>client请求：</b>")
        printHtml(json)
        println()
        star.request(route, bean, callback)
    }

    // args是可变参数列表
    function sendCommand(cmd, ...args) {
        let bean = {command: cmd}
        if (args.length > 0) {
            bean.command = cmd + " " + args.join(" ")
        }

        sendBean("console.command", bean, onCommand)
    }

    function onCommand(obj) {
        switch (obj.op) {
            case "log.list":
                render(() => <LogList logFiles={obj.data.logFiles} rootUrl={rootUrl}/>, printHtml(''))
                break
            case "history":
                render(() => <History/>, printHtml(''))
                break
            case "html":
                onHtml(obj.data)
                break
            case "table":
                onTable(obj.data)
                break
            case "empty":
                break
            default:
                onDefault(obj)
        }
    }

    function onEnter(evt) {
        let command = inputText.value
        if (command !== "") {
            inputText.value = ""

            // 检查是不是调用history命令
            if (command.startsWith("!")) {
                const index = Number(command.substring(1)) - 1
                // console.log("index:", index)
                if (!Number.isNaN(index)) {
                    command = historyStore.getHistory(index)
                    command = historyStore.getHistory(index)
                }
            }

            let texts = command.split(/\s+/)  // 支持连续多个空格
            let textsLength = texts.length
            const name = texts[0]

            if (name === "help") {
                sendCommand(name, rootUrl)
                historyStore.add(command)
            } else if (textsLength >= 2 && (name === "sub" || name === "unsub")) {
                const bean = {
                    topic: texts[1],
                };

                const route = "console." + name
                sendBean(route, bean, onCommand)
                historyStore.add(command)
            } else if (textsLength >= 2 && name === "auth") {
                username = texts[1]
                isAuthorizing = true
                // $el.type = "password"
                evt.target.type = "password"
                printWithTimestamp(command + "<br/> <h3>请输入密码：</h3><br/>")
                historyStore.add(command)
            } else if (isAuthorizing && textsLength >= 1) {
                isAuthorizing = false
                // this.$el.type = "text"
                evt.target.type = "text"
                login.login(username, name).then()
            } else {
                sendCommand(texts.join(' '))
                historyStore.add(command)
            }
        } else {
            printWithTimestamp('')
        }
    }

    function onTab(evt) {
        const text = inputText.value
        if (text.length > 0) {
            const bean = {
                head: text,
            }

            star.request("console.hint", bean, (list) => {
                const size = list.length
                if (size > 0) {
                    const names = list.map(v => v.Name)
                    inputText.value = longestCommonPrefix(names)
                    if (size > 1) {
                        // todo 这个可以化简
                        onTable(JSON.stringify(list))
                    }
                }
            })
        }
    }

    function onUpDown(evt) {
        const step = evt.key === 'ArrowUp' ? -1 : 1
        const nextText = historyStore.move(step)

        // 按bash中history的操作习惯, 如果是arrow down的话, 最后一个应该是""
        if (nextText !== '' || step === 1) {
            inputText.value = nextText

            setTimeout(() => {
                let position = nextText.length
                evt.target.setSelectionRange(position, position)
                evt.target.focus()
            })
        }
    }

    function onKeyDown(evt) {
        const key = evt.key
        let eaten = false
        if (key === 'Enter') {
            onEnter(evt)
            eaten = true
        } else if (key === 'Tab') {
            onTab(evt)
            eaten = true
        } else if (key === 'ArrowUp' || key === 'ArrowDown') {
            onUpDown(evt)
            eaten = true
        }

        if (eaten) {
            evt.preventDefault()
        }

        return true
    }

    return (
        <>
            <div id="mainPanel"></div>
            <div id="inputBoxDiv">
                <input id="inputBox" ref={inputText} placeholder="Tab补全命令, Enter执行命令" onKeyDown={onKeyDown}/>
            </div>
        </>
    );
};

render(() => <App/>, document.getElementById('app'))