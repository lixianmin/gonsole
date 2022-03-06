import {defineComponent} from "vue";
import {ElTable, ElTableColumn} from "element-plus";
import {scrollMainPanelToBottom} from "../code/main_panel";

export default defineComponent(
    {
        props: {
            tableData: {type: String}
        }
        , setup(props) {
            let tableData = JSON.parse(props.tableData)
            const isStruct = tableData[0] === undefined
            // 如果是struct, 则转为array
            if (isStruct) {
                tableData = [tableData]
            } else {
                // 如果list, 则在首列加入行号
                for (let i = 0; i < tableData.length; i++) {
                    let results = {" ": i + 1}
                    for (const [key, value] of Object.entries(tableData[i])) {
                        results[key] = value
                    }
                    tableData[i] = results
                }
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

            return () =>
                <div>
                    <ElTable data={tableData} tableLayout="fixed">
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