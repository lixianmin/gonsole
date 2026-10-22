'use strict'
/********************************************************************
 created:    2023-03-02
 author:     lixianmin

 Copyright (C) - All Rights Reserved
 *********************************************************************/
import {For} from "solid-js";
import {getHumanReadableSize} from "../code/tools";
import {escapeHtml, encodeUrlPath} from "../code/html";

export default function LogList(props) {
    // https://blog.ninja-squad.com/2021/09/30/script-setup-syntax-in-vue-3/
    const logFiles = props.logFiles
    const count = logFiles.length
    const totalSize = getHumanReadableSize(logFiles.map(fi => fi.size).reduce((last, current) => last + current, 0))

    function fetchNameHtml(fi) {
        // 修正: 之前路径与样本直接拼进innerHTML, 路径里的引号可以逃逸出onclick的js字符串
        const url = escapeHtml(`${props.rootUrl}/${encodeUrlPath(fi.path)}?access_token=${encodeUrlPath(fi.access_token)}`)
        const path = escapeHtml(fi.path)
        const sample = escapeHtml(fi.sample)
        // JSON.stringify生成带双引号的js字符串字面量, 再用escapeHtml转义双引号, 避免提前结束html属性
        const jsPath = escapeHtml(JSON.stringify(fi.path))

        return `<a href="${url}">${path}</a>
          <span class="tips_text">${sample}</span>
          <input type="button" class="copy_button" onclick="copyToClipboard(${jsPath})" value="复制"/>`
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
                        <div innerHTML={fetchNameHtml(fi)} className='tips'></div>
                    </td>
                    <td>{fi.mod_time}</td>
                </tr>
            }</For>
            </tbody>
        </table>
        <br/>
    </>
}