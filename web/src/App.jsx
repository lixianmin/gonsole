'use strict'
/********************************************************************
 created:    2023-02-27
 author:     lixianmin

 Copyright (C) - All Rights Reserved
 *********************************************************************/

import './App.module.css';

import {StartX} from "./code/starx";
import {createWebConfig} from "./code/web_config.js";
import {createLogin} from "./code/login";
import {useHistoryStore} from "./code/use_history_store.js";
import moment from "moment";
import {longestCommonPrefix, useKeyDown} from "./code/tools";
import History from "./widgets/History";
import {render} from "solid-js/web";
import JsonTable from "./widgets/JsonTable";
import {onMount} from "solid-js";
import InputBox from "./widgets/InputBox";
import MainPanel, {printHtml, println, printWithTimestamp} from "./widgets/MainPanel";
import LogList from "./widgets/LogList";
import {newSession} from "@src/code/network/session";

// todo 修改从golang的template传参到js的逻辑, 不再使用title
// todo disconnected from server的时候, 写一个online time
/**
 * todo 需要在readme中加入npm的开发和使用流程
 * todo 各种js中的any需要调整一下
 * todo 打包后生成的assets的根目录是否需要修改
 */

const App = () => {
    let inputBox
    let username = ''
    let isAuthorizing = false

    const config = createWebConfig()
    const historyStore = useHistoryStore()

    const session = newSession()
    const rootUrl = config.getRootUrl()

    // 开放sendCommand方法, 使client端写js代码的时候用websocket跟server交互
    window.sendCommand = sendCommand

    let login = createLogin((cmd, username, digestOrToken, fingerprint) => {
        printWithTimestamp("<b>client请求：</b>")
        printHtml(`${cmd} ${username} [digest | token] fingerprint`)
        println()

        const bean = {command: `${cmd} ${username} ${digestOrToken} ${fingerprint}`}
        return new Promise(resolve => {
            // 把callback改为promise
            session.request("console.command", bean, obj => {
                const cloned = {...obj.data}  // shadow clone
                resolve(obj.data)

                delete cloned.token
                const text = JSON.stringify(cloned)
                printWithTimestamp("<b>server响应：</b>" + text)
                println()
            })
        })
    })

    session.connect(config.getWebsocketUrl(), (nonce) => {
        console.log("websocket connected")
        printHtml(config.body)
        println()
        login.tryAutoLogin(nonce).then()
    })

    const uptime = new Date()
    session.on("disconnect", (response, err) => {
        const onlineTime = moment.duration(new Date().getTime() - uptime.getTime(), "milliseconds").humanize()
        printWithTimestamp(`<b> disconnected from server after ${onlineTime} </b>`)
    })

    session.on("console.html", onHtml)
    session.on("console.default", onDefault)

    function onHtml(response, err) {
        printWithTimestamp("<b>server响应：</b>" + response)
        println()
    }

    function onTable(response, err) {
        printHtml(() => <JsonTable tableData={response}/>)
    }

    function onDefault(response, err) {
        const operation = response
        const text = JSON.stringify(operation)
        printWithTimestamp("<b>server响应：</b>" + text)
        println()
    }

    function sendBean(route, bean, callback) {
        const json = JSON.stringify(bean)
        printWithTimestamp("<b>client请求：</b>")
        printHtml(json)
        println()
        session.request(route, bean, callback)
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
                printHtml(() => <LogList logFiles={obj.data.logFiles} rootUrl={rootUrl}/>)
                break
            case "history":
                printHtml(() => <History/>)
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

    onMount(() => {
        useKeyDown(window, 'Enter', evt => {
            const control = document.activeElement;
            if (control !== inputBox) {
                inputBox.focus()
                // return false的意思是：这个按键事件当前代码处理了，不再bubble上传这个事件。
                // 默认情况下会继续传播按键事件，Enter会导致页面refresh
                return false
            }
        })

        useKeyDown(inputBox, 'Enter', evt => {
            let command = inputBox.value
            if (command !== "") {
                inputBox.value = ""

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
                    evt.target.type = "password"
                    printWithTimestamp(command + "<br/> <h3>请输入密码：</h3><br/>")
                    historyStore.add(command)
                } else if (isAuthorizing && textsLength >= 1) {
                    isAuthorizing = false
                    evt.target.type = "text"
                    login.login(username, name).then()
                } else {
                    sendCommand(texts.join(' '))
                    historyStore.add(command)
                }
            } else {
                printWithTimestamp('')
            }
        })

        useKeyDown(inputBox, 'Tab', evt => {
            const text = inputBox.value
            if (text.length > 0) {
                const bean = {
                    head: text,
                }

                session.request("console.hint", bean, (list) => {
                    const size = list.length
                    if (size > 0) {
                        const names = list.map(v => v.Name)
                        inputBox.value = longestCommonPrefix(names)
                        if (size > 1) {
                            onTable(JSON.stringify(list))
                        }
                    }
                })
            }

            evt.preventDefault()
        })
    })

    return <>
        <MainPanel/>
        <div id="inputBoxDiv">
            <InputBox id='inputBox' ref={inputBox}/>
        </div>
    </>
}

const app = document.getElementById('app')
render(() => <App/>, app)