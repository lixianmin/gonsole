package gonsole

import (
	"encoding/json"
	"github.com/gorilla/websocket"
	"github.com/lixianmin/gonsole/logger"
	"github.com/lixianmin/got/loom"
	"time"
)

/********************************************************************
created:    2019-11-16
author:     lixianmin

Copyright (C) - All Rights Reserved
*********************************************************************/

/*
	goWritePump 负责写数据到websocket链接：
	1. 程序需要保证每个conn只有一个goroutine负责写数据
	2. 发生任何异常时（如：conn写数据超时），都记录err日志并断开链接
	3. pingTicker.Stop()是必须要调用的，否则会导致资源泄漏，gc不能处理ticker。同时，因为Read(), Write()系列方法都只能各自在一个
		goroutine中调用，因此pingTicker只能放到goWritePump()中, https://godoc.org/github.com/gorilla/websocket#hdr-Concurrency
	4.
*/
func (client *Client) goWritePump(conn *websocket.Conn, readChan chan IBean) {
	defer loom.DumpIfPanic()

	var pingTicker = time.NewTicker(pingPeriod)

	// 注意：client很特殊，它由CacheServer管理，但自己负责自己的生命周期，因为它是第一时间发现自己要死的人
	// 所以，它不仅负责清理自己的数据，同时负责通知从TopicManger与CacheServer中把对自己的引用移除
	defer func() {
		// 请求CacheServer把自己移除
		var msg = DetachClient{Client: client}
		client.server.sendMessage(msg)

		// 关闭定时器
		pingTicker.Stop()

		// 请求断开链接
		_ = conn.SetWriteDeadline(time.Now().Add(writeDeadline))
		_ = conn.WriteMessage(websocket.CloseMessage, []byte{})
		_ = conn.Close()

		logger.Info("[defer func(%q)] disconnected, len(readChan)=%d, len(writeChan)=%d",
			client.GetRemoteAddress(), len(readChan), len(client.writeChan))
	}()

	for {
		select {
		case message := <-client.writeChan:
			client.writeOneMessage(conn, message)
		case <-pingTicker.C:
			// 由于某些原因， 虽然我们定时的这个pingTicker.C是20s触发一次，但client收到的时间间隔最大可能会在190s以上，因此client主动ping server是必须的，否则可能被踢
			_ = conn.SetWriteDeadline(time.Now().Add(writeDeadline))
			if err := conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				logger.Warn("[goWritePump(%q)] write PingMessage failed, err=%q", client.GetRemoteAddress(), err)
				client.Dispose()
				return
			}
		case <-client.wd.DisposeChan:
			return
		}
	}
}

func (client *Client) writeOneBean(conn *websocket.Conn, pushBean IBean) {
	var jsonBytes, err = json.Marshal(pushBean)
	if nil != err {
		logger.Error("[writeOneBean()] Failed to Marshal pushBean=%v, err=%s", pushBean, err)
		return
	}

	client.writeOneMessage(conn, jsonBytes)
}

func (client *Client) writeOneMessage(conn *websocket.Conn, message []byte) {
	_ = conn.SetWriteDeadline(time.Now().Add(writeDeadline))
	writer, err := conn.NextWriter(websocket.TextMessage)
	if err != nil {
		logger.Info("[conn.NextWriter(%q)] err=%q", client.GetRemoteAddress(), err)
		client.Dispose()
		return
	}

	_, _ = writer.Write(message)

	if err := writer.Close(); err != nil {
		logger.Info("[writer.Close(%q)] err=%q", client.GetRemoteAddress(), err)
		client.Dispose()
	}
}
