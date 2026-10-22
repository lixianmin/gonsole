'use strict'
/********************************************************************
 created:    2023-03-02
 author:     lixianmin

 Copyright (C) - All Rights Reserved
 *********************************************************************/
import {For} from "solid-js";
import {getHumanReadableSize} from "../code/tools";
import {encodeUrlPath} from "../code/html";
import {fetchFileWithToken} from "../code/download";
import {printWithTimestamp, println} from "./MainPanel";

export default function LogList(props) {
    // https://blog.ninja-squad.com/2021/09/30/script-setup-syntax-in-vue-3/
    const logFiles = props.logFiles
    const count = logFiles.length
    const totalSize = getHumanReadableSize(logFiles.map(fi => fi.size).reduce((last, current) => last + current, 0))

    // 修正: 之前access_token是拼在url的query里的, 令牌会进入浏览器历史与Referer,
    // 现在改为放在X-Access-Token请求头中, 通过fetch下载
    async function downloadLogFile(fi) {
        const url = `${props.rootUrl}/${encodeUrlPath(fi.path)}`
        try {
            const blob = await fetchFileWithToken(url, fi.access_token)
            const objectUrl = URL.createObjectURL(blob)
            const link = document.createElement("a")
            link.href = objectUrl
            link.download = fi.path
            document.body.appendChild(link)
            link.click()
            document.body.removeChild(link)
            URL.revokeObjectURL(objectUrl)
        } catch (ex) {
            printWithTimestamp(`<b> 下载日志文件失败: ${ex.message} </b>`)
            println()
        }
    }

    function copyPath(fi) {
        window.copyToClipboard(fi.path)
    }

    return <>
        <b>日志文件列表：</b><br/>
        count: &nbsp; {count} <br/>
        total: &nbsp; {totalSize} <br/>
        <br/>
        <table>
            <thead>
            <tr>
                <th></th>
                <th>Size</th>
                <th>Name</th>
                <th>Modified Time</th>
            </tr>
            </thead>
            <tbody>
            <For each={logFiles}>{(fi, index) =>
                <tr>
                    <td>{index() + 1}</td>
                    <td>{getHumanReadableSize(fi.size)}</td>
                    <td>
                        {/*修正: 之前这里用innerHTML拼接, 服务端返回的路径与样本会被当成html解析, 现在改用solid的文本绑定*/}
                        <div class="tips">
                            <a href="#" onClick={evt => {
                                evt.preventDefault()
                                downloadLogFile(fi)
                            }}>{fi.path}</a>
                            <span class="tips_text">{fi.sample}</span>
                            <input type="button" class="copy_button" onClick={() => copyPath(fi)} value="复制"/>
                        </div>
                    </td>
                    <td>{fi.mod_time}</td>
                </tr>
            }</For>
            </tbody>
        </table>
        <br/>
    </>
}
