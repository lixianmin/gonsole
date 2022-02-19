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
                this.currentIndex = history.length; // 初始大小
            }
        }

        // 在unload时将history存储到localStorage中
        window.onunload = evt => {
            const data = this.list.slice(-100)
            localStorage.setItem(key, JSON.stringify(data))
        }
    }

    public add(command): void {
        const history = this.list
        const size = history.length;
        // 如果history中存储的最后一条与command不一样，则将command加入到history列表。否则将historyIndex调整到最后
        if (size === 0 || history[size - 1] !== command) {
            this.currentIndex = history.push(command)
        } else { // addHistory()都是在输入命令时才调用的，这时万一historyIndex处于history数组的中间位置，将其调整到最后
            this.currentIndex = history.length;
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
            if (step < 0 && nextIndex >= 0
                || step > 0 && nextIndex < history.length) {
                this.currentIndex = nextIndex
                const text = this.list[nextIndex]
                return text
            }
        }

        return ""
    }

    private currentIndex = -1
    private readonly list: string[] = []
}