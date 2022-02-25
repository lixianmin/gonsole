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
                this._list = json
                this._currentIndex = this._list.length // 初始大小
            }
        }

        // 在unload时将history存储到localStorage中
        window.onunload = evt => {
            const data = this._list.slice(-100)
            localStorage.setItem(key, JSON.stringify(data))
        }
    }

    public add(command: string): void {
        if (command != null && command != "") {
            const list = this._list
            const size = list.length

            // 如果history中存储的最后一条与command不一样，则将command加入到history列表。否则将historyIndex调整到最后
            if (size == 0 || list[size - 1] !== command) {
                this._currentIndex = list.push(command)
            } else { // add()都是在输入命令时才调用的，这时万一historyIndex处于history数组的中间位置，将其调整到最后
                this._currentIndex = list.length
            }
        }
    }

    public get histories(): string[] {
        return this._list
    }

    public getHistory(index: number): string {
        if (index >= 0 && index < this._list.length) {
            return this._list[index]
        }

        return ""
    }

    public get count(): number {
        return this._list.length
    }

    public move(step: number): string {
        if (step != 0) {
            let nextIndex = this._currentIndex + step
            if (nextIndex >= 0 && nextIndex < this._list.length) {
                this._currentIndex = nextIndex
                const text = this._list[nextIndex]
                // console.log(this.toString())
                return text
            }
        }

        return ""
    }

    public toString(): string {
        return `currentIndex=${this._currentIndex}, list=[${this._list}]`
    }

    private _currentIndex = -1
    private readonly _list: string[] = []
}