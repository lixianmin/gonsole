import {defineComponent} from "vue";
import {ElTable, ElTableColumn} from "element-plus";
import {scrollMainPanelToBottom} from "../code/main_panel";

export default defineComponent(
    {
        props: {
            tableData: {type: String}
        }
        , setup(props) {
            const tableData = JSON.parse(props.tableData)
            return () =>
                <div>
                    <ElTable data={tableData} style={{width: '80%'}}>
                        {
                            Object.keys({tableData}.tableData[0]).map(item => {
                                return <ElTableColumn prop={item} label={item} sortable/>
                            })
                        }
                    </ElTable>
                    <p/>
                </div>
        }
        , mounted() {
            setTimeout(() => {
                scrollMainPanelToBottom()
            })
        }
    })