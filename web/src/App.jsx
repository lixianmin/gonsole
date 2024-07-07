'use strict'
/********************************************************************
 created:    2023-02-27
 author:     lixianmin

 Copyright (C) - All Rights Reserved
 *********************************************************************/

import './App.module.css';
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
import MainPanel, {changeWidget, printHtml, println, printWithTimestamp} from "./widgets/MainPanel";
import LogList from "./widgets/LogList";
import {newSession} from "@src/code/road/session";

// todo 修改从golang的template传参到js的逻辑, 不再使用title
// todo disconnected from server的时候, 写一个online time
/**
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
            session.request("console.command", bean, response => {
                const cloned = {...response.data}  // shadow clone
                resolve(response.data)

                delete cloned.token
                const text = JSON.stringify(cloned)
                printWithTimestamp("<b>server响应：</b>" + text)
                println()
            })
        })
    })

    session.connect(config.getWebsocketUrl(), (nonce) => {
        const time = moment(new Date()).format("HH:mm:ss.S")
        console.log(`[${time}] websocket connected`)

        printHtml(config.body)
        println()
        login.tryAutoLogin(nonce).then()
    })

    const uptime = new Date()
    session.on("disconnect", (response, err) => {
        const onlineTime = moment.duration(new Date().getTime() - uptime.getTime(), "milliseconds").humanize()
        printWithTimestamp(`<b> disconnected from server after ${onlineTime} </b>`)
    })

    session.on("console.html", onHtmlHandler)
    session.on("console.default", onDefaultHandler)
    session.on("console.stream", onStreamHandler)

    function onHtmlHandler(response, err) {
        printWithTimestamp("<b>server响应：</b>" + response.data)
        println()
    }

    function onTableData(data) {
        printHtml(() => <JsonTable tableData={data}/>)
    }

    function onDefaultHandler(response, err) {
        const input = err ?? response
        const isString = typeof input === 'string'
        // 如果输入本身就是string的话, 则不调用stringify(), 解决序列的json串全是\\的问题
        const text = isString ? input : JSON.stringify(input)

        printWithTimestamp("<b>server响应：</b>" + text)
        println()
    }

    let streamWidget = undefined

    function onReceivedStreamText(text) {
        text = text ?? ""
        text = text.replace(/\r\n|\r|\n/g, '<br>')

        streamWidget.html += text
        changeWidget(streamWidget)
    }

    function onStreamHandler(response, err) {
        if (err) {
            printWithTimestamp("<b>server响应：</b>" + err)
            println()

            streamWidget = undefined
            return
        }

        const item = JSON.parse(response)
        if (!streamWidget) {
            printWithTimestamp("<b>server响应：</b>")
            streamWidget = printHtml('')
            onReceivedStreamText(item.text)
            return
        }

        if (item.done) {
            streamWidget = undefined
            return
        }

        onReceivedStreamText(item.text)
        // console.log(`response=${response}, html=${streamWidget.html}`)
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

    function sendRoadRequest(texts) {
        if (texts.length < 2) {
            console.log("InvalidCommandFormat", texts)
            return
        }

        const route = texts[1];

        let data = ""
        if (texts.length > 2) {
            data = texts.slice(2).join(' ')
        }

        // console.log(`route="${route}", data=${data}`)
        const bean = JSON.parse(data);
        sendBean(route, bean, onCommand)
    }

    function onCommand(response, err) {
        if (response && response.op) {
            switch (response.op) {
                case "log.list":
                    printHtml(() => <LogList logFiles={response.data.logFiles} rootUrl={rootUrl}/>)
                    break
                case "history":
                    printHtml(() => <History/>)
                    break
                case "html":
                    onHtmlHandler(response)
                    break
                case "table":
                    onTableData(response.data)
                    break
                case "empty":
                    break
                default:
                    onDefaultHandler(response, err)
            }
        } else {
            onDefaultHandler(response, err)
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
                } else if (textsLength >= 2 && name === 'request') {
                    sendRoadRequest(texts)
                    historyStore.add(command)
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
                            onTableData(JSON.stringify(list))
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