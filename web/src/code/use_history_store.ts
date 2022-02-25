/********************************************************************
 created:    2022-02-25
 author:     lixianmin

 Copyright (C) - All Rights Reserved
 *********************************************************************/
import {defineStore} from "pinia";
import {useLocalStorage} from "@vueuse/core";
// todo 这个如果超过100个, 就不应该全部分存储到localstorage中了
interface HistoryStore {
    currentIndex: number,
    list: string[],
}

export const useHistoryStore = defineStore({
    id: "historyStore", // id is required so pinia can connect to the devtools
    state: () => (useLocalStorage("this.is.history.store", {
        currentIndex: 0,
        list: [],
    } as HistoryStore)),
    getters: {
        histories: state => state.list
        , count: state => state.list.length
    },
    actions: {
        add(command: string): void {
            if (command != null && command != "") {
                const list = this.list
                const size = list.length

                // 如果history中存储的最后一条与command不一样，则将command加入到history列表。否则将historyIndex调整到最后
                if (size == 0 || list[size - 1] !== command) {
                    this.currentIndex = list.push(command)
                } else { // add()都是在输入命令时才调用的，这时万一historyIndex处于history数组的中间位置，将其调整到最后
                    this.currentIndex = list.length
                }
            }
        }
        , getHistory(index: number): string {
            if (index >= 0 && index < this.list.length) {
                return this.list[index]
            }

            return ""
        }
        , move(step: number): string {
            if (step != 0) {
                let nextIndex = this.currentIndex + step
                if (nextIndex >= 0 && nextIndex < this.list.length) {
                    this.currentIndex = nextIndex
                    const text = this.list[nextIndex]
                    // console.log(this.toString())
                    return text
                }
            }

            return ""
        }
    }
})
