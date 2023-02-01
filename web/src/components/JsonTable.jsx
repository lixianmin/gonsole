/********************************************************************
 created:    2023-02-01
 author:     lixianmin

 Copyright (C) - All Rights Reserved
 *********************************************************************/
import {defineComponent} from "vue"
import {scrollMainPanelToBottom} from "@/code/main_panel"

function processTableData(tableData) {
    let results = JSON.parse(tableData)
    const isStruct = results[0] === undefined
    // 如果是struct, 则转为array
    if (isStruct) {
        results = [results]
    } else {
        // 如果list, 则在首列加入行号
        for (let i = 0; i < results.length; i++) {
            let items = {" ": i + 1}
            for (const [key, value] of Object.entries(results[i])) {
                items[key] = value
            }
            results[i] = items
        }
    }

    return results
}

export default defineComponent(
    {
        props: {
            tableData: {type: String}
        }
        , setup(props) {
            const tableData = processTableData(props.tableData)

            const headData = Object.keys(tableData[0]).map((item, index) => {
                    const handler = `sortTableByHead.call(this, ${index})`
                    return <th onclick={handler}>{item}</th>
                }
            )

            const bodyData = tableData.map(row => {
                const rowHtml = Object.values(row).map(item => <td>{item}</td>)
                return <tr>{rowHtml}</tr>
            })

            return () =>
                <div>
                    <table>
                        <thead>
                        <tr>
                            {headData}
                        </tr>
                        </thead>
                        <tbody>
                        {bodyData}
                        </tbody>
                    </table>
                    <br/>
                </div>
        }
        , mounted() {
            setTimeout(() => {
                scrollMainPanelToBottom()
            })
        }
    })