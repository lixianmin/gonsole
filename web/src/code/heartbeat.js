/********************************************************************
 created:    2022-11-27
 author:     lixianmin

 Copyright (C) - All Rights Reserved
 *********************************************************************/

export function createHeartbeat() {
    let interval = 10 * 1000 // 心跳间隔, 单位(ms)
    let timeoutId = 0
    return {
        get interval() {
            return interval
        },
        set interval(v) {
            interval = v
        },
        setTimeout(callback) {
            if (typeof callback === 'function') {
                timeoutId = window.setTimeout(callback, interval * 3)
            }
        },
        clearTimeout() {
            if (timeoutId > 0) {
                window.clearTimeout(timeoutId)
                timeoutId = 0
            }
        }
    }
}