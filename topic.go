package gonsole

import (
	"github.com/lixianmin/gonsole/tools"
	"sync"
	"time"
)

/********************************************************************
created:    2020-06-05
author:     lixianmin

Copyright (C) - All Rights Reserved
*********************************************************************/

type Topic struct {
	Name        string             // 名称
	Note        string             // 描述
	Interval    time.Duration      // 推送周期
	PrepareData func() interface{} // 处理方法

	lock    sync.Mutex
	clients map[*Client]struct{}
}

func (topic *Topic) start() {
	go func() {
		tools.RandomSleep(0, topic.Interval)
		for {
			func() {
				var data = topic.PrepareData()
				topic.lock.Lock()
				defer topic.lock.Unlock()

				for client := range topic.clients {
					client.SendBean(data)
				}
			}()
			time.Sleep(topic.Interval)
		}
	}()
}

func (topic *Topic) addClient(client *Client) {
	if client != nil {
		topic.lock.Lock()
		defer topic.lock.Unlock()

		topic.clients[client] = struct{}{}
	}
}

func (topic *Topic) removeClient(client *Client) {
	if client != nil {
		topic.lock.Lock()
		defer topic.lock.Unlock()

		delete(topic.clients, client)
	}
}
