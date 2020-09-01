package gonsole

import (
	"encoding/json"
	"github.com/lixianmin/bugfly"
	"github.com/lixianmin/gonsole/beans"
	"github.com/lixianmin/gonsole/ifs"
	"github.com/lixianmin/gonsole/logger"
	"github.com/lixianmin/gonsole/tools"
	"github.com/lixianmin/got/loom"
	"sync"
)

/********************************************************************
created:    2019-11-16
author:     lixianmin

Copyright (C) - All Rights Reserved
*********************************************************************/

type Client struct {
	wc             loom.WaitClose
	session        *bugfly.Session
	writeChan      chan []byte
	server         *Server
	topics         map[string]struct{}
	isAuthorized   bool
	onCloseHandler func()
	Attachment     sync.Map
}

// newClient 创建一个新的client对象
func newClient(server *Server, session *bugfly.Session) *Client {
	const chanSize = 8
	var readChan = make(chan ifs.Bean, chanSize)

	var client = &Client{
		session:      session,
		writeChan:    make(chan []byte, chanSize),
		server:       server,
		topics:       make(map[string]struct{}),
		isAuthorized: false,
	}

	go client.goLoop(readChan)
	return client
}

/*
	goLoop 是client的主循环。
	1. goLoop()不能与goWritePump()合并为一个。早期的确是这样设计，后来发现有deadlock:在处理订阅消息的cmd时，最终需要调用sendBean()
		发送数据到writeChan，但是由于生产者、消费者由同一个loop处理，导致在生产的过程中无法同时消费，因此导致了deadlock
	2.因为是主循环，所以相关的容器类会放到这里，比如topics
*/
func (client *Client) goLoop(readChan <-chan ifs.Bean) {
	defer loom.DumpIfPanic()

	for {
		select {
		case bean := <-readChan:
			switch bean := bean.(type) {
			case *beans.Subscribe:
				loopClientSubscribe(client, bean)
			case *beans.Unsubscribe:
				loopClientUnsubscribe(client, bean)
			case *beans.Ping:
				var pong = &beans.Pong{beans.BasicResponse{Operation: "pong"}}
				client.SendBean(pong)
			default:
				logger.Error("unexpected bean type: %T", bean)
			}
		case <-client.wc.C():
			if nil != client.onCloseHandler {
				client.onCloseHandler()
			}
			return
		}
	}
}

func loopClientSubscribe(client *Client, bean *beans.Subscribe) {
	var topicId = bean.TopicId
	var topic = client.server.getTopic(topicId)
	if topic == nil || !(topic.IsPublic || client.isAuthorized) {
		client.SendBean(beans.NewBadRequestRe(bean.RequestId, InvalidTopic, "尝试订阅非法topic"))
		return
	}

	if _, ok := client.topics[topicId]; ok {
		client.SendBean(beans.NewBadRequestRe(bean.RequestId, InvalidOperation, "重复订阅同一个主题"))
		return
	}

	topic.addClient(client)
	client.topics[topicId] = struct{}{}
	client.SendBean(beans.NewSubscribeRe(bean.RequestId, topicId))
	//client.SendBean(topic.BuildData())
}

func loopClientUnsubscribe(client *Client, bean *beans.Unsubscribe) {
	var topicId = bean.TopicId

	var topic = client.server.getTopic(topicId)
	if topic == nil {
		client.SendBean(beans.NewBadRequestRe(bean.RequestId, InvalidTopic, "尝试取消非法topic"))
		return
	}

	if _, ok := client.topics[topicId]; !ok {
		client.SendBean(beans.NewBadRequestRe(bean.RequestId, InvalidOperation, "尝试取消未订阅主题"))
		return
	}

	topic.removeClient(client)
	delete(client.topics, topicId)
	client.SendBean(beans.NewUnsubscribeRe(bean.RequestId, topicId))
}

// 这个不使用启goroutine去写client.writeChan，虽然不卡死了，但是无法保证顺序了，这就完蛋了
func (client *Client) innerSendBytes(data []byte) {
	select {
	case client.writeChan <- data:
	case <-client.wc.C():
	}
}

func (client *Client) SendHtml(html string) {
	var bean = beans.HtmlResponse{
		BasicResponse: beans.BasicResponse{
			Operation: "html",
		},
		Html: html,
	}

	var jsonBytes, err = tools.MarshalUnescape(bean)
	if err == nil {
		client.innerSendBytes(jsonBytes)
	} else {
		logger.Warn("Can not marshal bean=%v, err=%s", bean, err)
	}
}

func (client *Client) SendBean(bean interface{}) {
	if bean != nil {
		var jsonBytes, err = json.Marshal(bean)
		if err == nil {
			client.innerSendBytes(jsonBytes)
		} else {
			logger.Warn("Can not marshal bean=%v, err=%s", bean, err)
		}
	}
}

func (client *Client) SetAuthorized(b bool) {
	client.isAuthorized = b
}

func (client *Client) OnClose(handler func()) {
	client.onCloseHandler = handler
}

func (client *Client) GetRemoteAddress() string {
	return client.session.RemoteAddr().String()
}

func (client *Client) Close() {
	client.wc.Close(nil)
}
