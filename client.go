package gonsole

import (
	"github.com/lixianmin/bugfly"
	"github.com/lixianmin/gonsole/beans"
	"github.com/lixianmin/gonsole/logger"
	"github.com/lixianmin/gonsole/tools"
)

/********************************************************************
created:    2019-11-16
author:     lixianmin

Copyright (C) - All Rights Reserved
*********************************************************************/

type Client struct {
	session *bugfly.Session
	server  *Server
	topics  map[string]struct{}
}

// newClient 创建一个新的client对象
func newClient(server *Server, session *bugfly.Session) *Client {
	const chanSize = 8

	var client = &Client{
		session: session,
		server:  server,
		topics:  make(map[string]struct{}),
	}

	return client
}

func loopClientSubscribe(client *Client, bean *beans.Subscribe) {
	//var topicId = bean.TopicId
	//var topic = client.server.getTopic(topicId)
	//if topic == nil || !(topic.IsPublic || client.isAuthorized) {
	//	client.Push(beans.NewBadRequestRe(bean.RequestId, InvalidTopic, "尝试订阅非法topic"))
	//	return
	//}
	//
	//if _, ok := client.topics[topicId]; ok {
	//	client.Push(beans.NewBadRequestRe(bean.RequestId, InvalidOperation, "重复订阅同一个主题"))
	//	return
	//}
	//
	//topic.addClient(client)
	//client.topics[topicId] = struct{}{}
	//client.Push(beans.NewSubscribeRe(bean.RequestId, topicId))
	//client.Push(topic.BuildData())
}

func loopClientUnsubscribe(client *Client, bean *beans.Unsubscribe) {
	//var topicId = bean.TopicId

	//var topic = client.server.getTopic(topicId)
	//if topic == nil {
	//	client.Push(beans.NewBadRequestRe(bean.RequestId, InvalidTopic, "尝试取消非法topic"))
	//	return
	//}
	//
	//if _, ok := client.topics[topicId]; !ok {
	//	client.Push(beans.NewBadRequestRe(bean.RequestId, InvalidOperation, "尝试取消未订阅主题"))
	//	return
	//}
	//
	//topic.removeClient(client)
	//delete(client.topics, topicId)
	//client.Push(beans.NewUnsubscribeRe(bean.RequestId, topicId))
}

func (client *Client) PushHtml(html string) {
	var bean = beans.HtmlResponse{
		Html: html,
	}

	var jsonBytes, err = tools.MarshalUnescape(bean)
	if err == nil {
		err = client.session.Push("console.html", jsonBytes)
		if err != nil {
			logger.Info("err=%q", err)
		}
	} else {
		logger.Warn("Can not marshal bean=%v, err=%s", bean, err)
	}
}

func (client *Client) Push(route string, v interface{}) {
	err := client.session.Push(route, v)
	if err != nil {
		logger.Info("err=%q", err)
	}
}

func (client *Client) OnClosed(callback func()) {
	client.session.OnClosed(callback)
}

func (client *Client) Session() *bugfly.Session {
	return client.session
}

func (client *Client) Attachment() *bugfly.Attachment {
	return client.session.Attachment()
}
