'use strict'
/********************************************************************
 created:    2023-02-01
 author:     lixianmin

 Copyright (C) - All Rights Reserved
 *********************************************************************/
import {For} from "solid-js";
import {escapeHtml} from "@src/code/html";

function processTableData(tableData) {
    let results = JSON.parse(tableData)
    const isStruct = results[0] === undefined
    // 如果是struct, 则转为array
    if (isStruct) {
        results = [results]
    } else {
        // 如果list, 则在首列加入行号
        for (let i = 0; i < results.length; i++) {
            const headIndex = [' ', i + 1]
            results[i] = new Map([headIndex, ...Object.entries(results[i])])
        }
    }

    return results
}

export default function JsonTable(props) {
    const mapList = processTableData(props.tableData)

    // 下面这种用法, 能把color:red放进去, 但onClick放不进去, 有些奇怪
    // const headData = Array.from(mapList[0].keys()).map((item, index) => {
    //     const attributes = {onClick: `sortTableByHead.call(this, ${index})`, style: 'color:red'}
    //     return <tr {...attributes}>{item}</tr>
    // })

    // 因为我们产生的代码里有字符串, 因此只能用innerHTML内嵌的方式
    // 修正: 表头内容之前直接拼进innerHTML, 服务端返回的内容含尖括号时会被当成html解析
    const headData = Array.from(mapList[0].keys()).map((item, index) => `<th onClick= "sortTableByHead.call(this, ${index})">${escapeHtml(item)}</th>`).join('')

    // const bodyData = mapList.map(row => {
    //     const rowHtml = Array.from(row.values()).map(item => <td>{item}</td>)
    //     return <tr>{rowHtml}</tr>
    // })

    return <>
        <table>
            <thead>
            <tr innerHTML={headData}/>
            </thead>
            <tbody>
            <For each={mapList}>{row =>
                <tr>
                    <For each={Array.from(row.values())}>{item =>
                        <td>{item}</td>
                    }</For>
                </tr>
            }</For>
            </tbody>
        </table>
        <br/>
    </>
}