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
	Example       string           // 举例
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

	// 持锁初始化sessions.d：热更时RegisterTopic可能在已有session订阅（addClient持锁写d）
	// 之后调用，无锁赋值会与addClient/removeClient产生data race；
	// 且若addClient先于start()执行，d为nil map，赋值直接panic
	topic.sessions.Lock()
	if topic.sessions.d == nil {
		topic.sessions.d = make(map[road.Session]struct{})
	}
	topic.sessions.Unlock()

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
		// 懒初始化：防止start()尚未执行（或Topic未start）时对nil map赋值panic
		if topic.sessions.d == nil {
			topic.sessions.d = make(map[road.Session]struct{})
		}
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

func (topic *Topic) GetExample() string {
	return topic.Example
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
