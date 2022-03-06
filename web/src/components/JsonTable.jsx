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
            // 如果是struct, 则转为array
            if ( tableData[0] === undefined) {
                tableData = [tableData]
            }

            return () =>
                <div>
                    <ElTable data={tableData} tableLayout="auto">
                        {
                            Object.keys({tableData}.tableData[0]).map(item => {
                                return <ElTableColumn prop={item} label={item} sortable/>
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