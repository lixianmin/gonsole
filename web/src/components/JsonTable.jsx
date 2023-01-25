import {defineComponent} from "vue";
import {ElTable, ElTableColumn} from "element-plus";
import {scrollMainPanelToBottom} from "../code/main_panel";

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

function onRenderHeader({column, $index}) {
    let after = " "
    switch (column.order) {
        case "ascending":
            after = "↑"
            break
        case "descending":
            after = "↓"
            break
    }

    return <div>{column.label}{after}</div>
}

export default defineComponent(
    {
        props: {
            tableData: {type: String}
        }
        , setup(props) {
            let tableData = processTableData(props.tableData)
            return () =>
                <div>
                    <ElTable data={tableData} tableLayout="auto">
                        {
                            Object.keys({tableData}.tableData[0]).map(item => {
                                return <ElTableColumn prop={item} label={item} sortable
                                                      renderHeader={onRenderHeader}/>
                            })
                        }
                    </ElTable>
                    <br/>
                </div>
        }
        , mounted() {
            setTimeout(() => {
                scrollMainPanelToBottom()
            })
        }
    })