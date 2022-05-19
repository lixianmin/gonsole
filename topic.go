package gonsole

import (
	"github.com/lixianmin/got/loom"
	"github.com/lixianmin/got/randx"
	"github.com/lixianmin/logo"
	"sync"
	"time"
)

/********************************************************************
created:    2020-06-05
author:     lixianmin

Copyright (C) - All Rights Reserved
*********************************************************************/

type Topic struct {
	loom.Flag
	Name          string           // 名称
	Note          string           // 描述
	Interval      time.Duration    // 推送周期
	BuildResponse func() *Response // 创建数据

	clients struct {
		sync.RWMutex
		d map[*Client]struct{}
	}
}

func (topic *Topic) start() {
	if topic.Interval <= 0 || topic.BuildResponse == nil {
		logo.Error("topic.Interval <= 0 || topic.BuildResponse == nil")
		return
	}

	topic.clients.d = make(map[*Client]struct{})

	go func() {
		time.Sleep(randx.Duration(0, topic.Interval))

		for {
			topic.clients.RLock()
			var count = len(topic.clients.d)
			if count > 0 {
				var response = topic.BuildResponse()
				var route = "console." + response.Operation
				for client := range topic.clients.d {
					_ = client.session.Push(route, response)
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

func (topic *Topic) IsBuiltin() bool {
	return topic.HasFlag(flagBuiltin)
}

func (topic *Topic) IsPublic() bool {
	return topic.HasFlag(FlagPublic)
}

func (topic *Topic) IsInvisible() bool {
	return topic.HasFlag(FlagInvisible)
}
