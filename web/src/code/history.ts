/********************************************************************
 created:    2022-01-19
 author:     lixianmin

 Copyright (C) - All Rights Reserved
 *********************************************************************/

export class History {
    public constructor() {
        const key = "list"
        const item = localStorage.getItem(key)
        if (item) {
            const json = JSON.parse(item)
            if (json) {
                this.list = json
                this.currentIndex = this.list.length // 初始大小
            }
        }

        // 在unload时将history存储到localStorage中
        window.onunload = evt => {
            const data = this.list.slice(-100)
            localStorage.setItem(key, JSON.stringify(data))
        }
    }

    public add(command: string): void {
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

    public getHistories(): string[] {
        return this.list
    }

    public getHistory(index: number): string {
        if (index >= 0 && index < this.list.length) {
            return this.list[index]
        }

        return ""
    }

    public move(step: number): string {
        if (step != 0) {
            let nextIndex = this.currentIndex + step
            if ( nextIndex >= 0 && nextIndex < this.list.length) {
                this.currentIndex = nextIndex
                const text = this.list[nextIndex]
                // console.log(this.toString())
                return text
            }
        }

        return ""
    }

    public toString() :string{
        return `currentIndex=${this.currentIndex}, list=[${this.list}]`
    }

    private currentIndex = -1
    private readonly list: string[] = []
}