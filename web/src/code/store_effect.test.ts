import { describe, expect, test } from "bun:test"
import { createEffect, createRoot } from "solid-js"
import { createStore, produce } from "solid-js/store"

// MainPanel用effect监听widgets的内容变化来决定是否滚动到底部,
// 这里验证"读取widget.html"确实能建立依赖, widget被替换后effect会重新执行
describe("solid的effect依赖追踪", () => {
    test("widget被替换后effect会重新执行", async () => {
        let runs = 0
        let setStore: any

        createRoot(() => {
            const [store, set] = createStore({widgets: [{html: "a"}]})
            setStore = set

            createEffect(() => {
                for (const widget of store.widgets) {
                    void widget.html
                }
                runs++
            })
        })

        // solid的effect是延迟执行的
        await Promise.resolve()
        expect(runs).toBe(1)

        setStore(produce((state: any) => {
            state.widgets[0] = {html: "b"}
        }))

        await Promise.resolve()
        expect(runs).toBe(2)
    })
})
