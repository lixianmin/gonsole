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
            const headIndex = [' ', i + 1]
            results[i] = new Map([headIndex, ...Object.entries(results[i])])
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
            const mapList = processTableData(props.tableData)

            const headData = Array.from(mapList[0].keys()).map((item, index) => {
                    const handler = `sortTableByHead.call(this, ${index})`
                    return <th onclick={handler}>{item}</th>
                }
            )

            const bodyData = mapList.map(row => {
                const rowHtml = Array.from(row.values()).map(item => <td>{item}</td>)
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
            this.$nextTick(() => {
                scrollMainPanelToBottom()
            })
        }
    })