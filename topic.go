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

	clients sync.Map
}

func (topic *Topic) start() {
	if topic.Interval <= 0 || topic.BuildData == nil {
		logger.Error("topic.Interval <= 0 || topic.BuildData == nil")
		return
	}

	go func() {
		var d = randx.Duration(0, topic.Interval)
		time.Sleep(d)

		for {
			func() {
				var data = topic.BuildData()
				topic.clients.Range(func(key, value interface{}) bool {
					var client = key.(*Client)
					client.SendBean(data)
					return true
				})
			}()
			time.Sleep(topic.Interval)
		}
	}()
}

func (topic *Topic) addClient(client *Client) {
	if client != nil {
		topic.clients.Store(client, nil)
	}
}

func (topic *Topic) removeClient(client *Client) {
	if client != nil {
		topic.clients.Delete(client)
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
