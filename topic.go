package gonsole

import (
	"github.com/lixianmin/gonsole/road"
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

	sessions struct {
		sync.RWMutex
		d map[road.Session]struct{}
	}
}

func (topic *Topic) start() {
	if topic.Interval <= 0 || topic.BuildResponse == nil {
		logo.Error("topic.Interval <= 0 || topic.BuildResponse == nil")
		return
	}

	topic.sessions.d = make(map[road.Session]struct{})

	go func() {
		time.Sleep(randx.Duration(0, topic.Interval))

		for {
			topic.sessions.RLock()
			var count = len(topic.sessions.d)
			if count > 0 {
				var response = topic.BuildResponse()
				var route = "console." + response.Operation
				for session := range topic.sessions.d {
					if err := session.Send(route, response); err != nil {
						logo.JsonW("route", route, "err", err)
					}
				}
			}
			topic.sessions.RUnlock()
			time.Sleep(topic.Interval)
		}
	}()
}

func (topic *Topic) addClient(session road.Session) {
	if session != nil {
		topic.sessions.Lock()
		topic.sessions.d[session] = struct{}{}
		topic.sessions.Unlock()
	}
}

func (topic *Topic) removeClient(session road.Session) {
	if session != nil {
		topic.sessions.Lock()
		delete(topic.sessions.d, session)
		topic.sessions.Unlock()
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
