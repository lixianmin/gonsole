package gonsole

import (
	"github.com/lixianmin/gonsole/logger"
	"github.com/lixianmin/got/randx"
	"sync"
	"time"
)

/********************************************************************
created:    2020-06-05
author:     lixianmin

Copyright (C) - All Rights Reserved
*********************************************************************/

type Topic struct {
	Name      string             // 名称
	Note      string             // 描述
	Interval  time.Duration      // 推送周期
	IsPublic  bool               // 非public方法需要登陆
	isBuiltin bool               // 是否为内置主题，排序时内置主题排在前面
	BuildData func() interface{} // 创建数据

	clients struct {
		sync.RWMutex
		d map[*Client]struct{}
	}
}

func (topic *Topic) start() {
	if topic.Interval <= 0 || topic.BuildData == nil {
		logger.Error("topic.Interval <= 0 || topic.BuildData == nil")
		return
	}

	go func() {
		time.Sleep(randx.Duration(0, topic.Interval))

		for {
			topic.clients.RLock()
			var count = len(topic.clients.d)
			if count > 0 {
				var data = topic.BuildData()
				for client := range topic.clients.d {
					client.SendBean(data)
				}
			}
			topic.clients.RUnlock()
			time.Sleep(topic.Interval)
		}
	}()
}

func (topic *Topic) addClient(client *Client) {
	if client != nil {
		topic.clients.Lock()
		topic.clients.d[client] = struct{}{}
		topic.clients.Unlock()
	}
}

func (topic *Topic) removeClient(client *Client) {
	if client != nil {
		topic.clients.Lock()
		delete(topic.clients.d, client)
		topic.clients.Unlock()
	}
}

func (topic *Topic) GetName() string {
	return topic.Name
}

func (topic *Topic) GetNote() string {
	return topic.Note
}

func (topic *Topic) CheckPublic() bool {
	return topic.IsPublic
}

func (topic *Topic) CheckBuiltin() bool {
	return topic.isBuiltin
}
