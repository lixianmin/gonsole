'use strict'
/********************************************************************
 created:    2023-03-02
 author:     lixianmin

 Copyright (C) - All Rights Reserved
 *********************************************************************/
import {For} from "solid-js";
import {getHumanReadableSize} from "../code/tools";

export default function LogList(props) {
    // https://blog.ninja-squad.com/2021/09/30/script-setup-syntax-in-vue-3/
    const logFiles = props.logFiles
    const count = logFiles.length
    const totalSize = getHumanReadableSize(logFiles.map(fi => fi.size).reduce((last, current) => last + current, 0))

    function fetchNameHtml(fi) {
        return `<a href="${props.rootUrl}/${fi.path}?access_token=${fi.access_token}">${fi.path}</a>
          <span class="tips_text">${fi.sample}</span>
          <input type="button" class="copy_button" onclick="copyToClipboard('${fi.path}')" value="复制"/>`
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