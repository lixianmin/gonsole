'use strict'

/********************************************************************
 created:    2022-11-27
 author:     lixianmin

 Copyright (C) - All Rights Reserved
 *********************************************************************/

/**
 * @typedef {Object} Heartbeat
 * @property {function(callback)} setTimeout --函数签名如果写成{function(callback : function)}, 则会被认为是2个参数, 然后导致编译不过
 * @property {function()} clearTimeout
 * @property {number} interval
 * @returns {Heartbeat}
 */
export function createHeartbeat() {
    let _interval = 10 * 1000 // 心跳间隔, 单位(ms)
    let _timeoutId = 0

    return {
        get interval() {
            return _interval
        },
        set interval(v) {
            _interval = v
        },
        setTimeout(callback) {
            if (typeof callback === 'function') {
                _timeoutId = window.setTimeout(callback, _interval * 3)
            }
        },
        clearTimeout() {
            if (_timeoutId > 0) {
                window.clearTimeout(_timeoutId)
                _timeoutId = 0
            }
        }
    }
}