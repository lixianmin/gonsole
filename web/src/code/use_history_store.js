'use strict'

/********************************************************************
 created:    2022-02-25
 author:     lixianmin

 Copyright (C) - All Rights Reserved
 *********************************************************************/

// 原先使用pinia的defineStore+useLocalStorage, 其实也是利用了localStorage去实现跨页面的存储, 实现复杂度高一些
// 现在改为每次add()的时候自动写入到localStorage中, 流程应该是一样的
function createHistoryStore() {
    const storageKey = "this.is.history.store"
    let _list = JSON.parse(localStorage.getItem(storageKey)) ?? []
    let _currentIndex = _list.length

    return {
        add(command) {
            if (typeof command === 'string' && command !== "") {
                const size = _list.length

                // 如果history中存储的最后一条与command不一样，则将command加入到history列表。否则将historyIndex调整到最后
                if (size === 0 || _list[size - 1] !== command) {
                    _currentIndex = _list.push(command)
                } else { // add()都是在输入命令时才调用的，这时万一historyIndex处于history数组的中间位置，将其调整到最后
                    _currentIndex = _list.length
                }
            }

            localStorage.setItem(storageKey, JSON.stringify(_list))
        },
        move(step) {
            if (typeof step === 'number' && step !== 0) {
                let nextIndex = _currentIndex + step
                if (nextIndex >= 0 && nextIndex < _list.length) {
                    _currentIndex = nextIndex
                    const text = _list.at(nextIndex)
                    // console.log(this.toString())
                    return text
                }
            }

            return ''
        },
        getHistoryList() {
            return _list
        },
        getHistory(index) {
            return typeof index === 'number' ? _list.at(index) : ''
        },
        getHistoryCount() {
            return _list.length
        }
    }
}

let store = createHistoryStore()

// 只所以导出useHistoryStore(), 是希望在所有的地方使用的都是同一个对象和存储, 如此而已
export const useHistoryStore = () => store

// import {defineStore} from "pinia";
// import {useLocalStorage} from "@vueuse/core";
//
// interface HistoryStore {
//     currentIndex: number,
//     list: string[],
// }
//
// export const useHistoryStore = defineStore({
//     id: "historyStore", // id is required so pinia can connect to the devtools
//     state: () => (useLocalStorage("this.is.history.store", {
//         currentIndex: 0,
//         list: [],
//     } as HistoryStore, {
//         serializer: {
//             read: (v: string) => {
//                 if (v) {
//                     let json = JSON.parse(v)
//                     if (json) {
//                         return {list: json, currentIndex: json.length} as HistoryStore
//                     }
//                 }
//             },
//             write: (v: any) => JSON.stringify(v.list.slice(-100)),
//         }
//     })),
//     getters: {
//         histories: state => state.list
//         , count: state => state.list.length
//     },
//     actions: {
//         add(command: string): void {
//             if (command != null && command != "") {
//                 const list = this.list
//                 const size = list.length
//
//                 // 如果history中存储的最后一条与command不一样，则将command加入到history列表。否则将historyIndex调整到最后
//                 if (size == 0 || list[size - 1] !== command) {
//                     this.currentIndex = list.push(command)
//                 } else { // add()都是在输入命令时才调用的，这时万一historyIndex处于history数组的中间位置，将其调整到最后
//                     this.currentIndex = list.length
//                 }
//             }
//         }
//         , getHistory(index: number): string {
//             if (index >= 0 && index < this.list.length) {
//                 return this.list[index]
//             }
//
//             return ""
//         }
//         , move(step: number): string {
//             if (step != 0) {
//                 let nextIndex = this.currentIndex + step
//                 if (nextIndex >= 0 && nextIndex < this.list.length) {
//                     this.currentIndex = nextIndex
//                     const text = this.list[nextIndex]
//                     // console.log(this.toString())
//                     return text
//                 }
//             }
//
//             return ""
//         }
//     }
// })
