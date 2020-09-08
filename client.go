package gonsole

import (
	"github.com/lixianmin/road"
)

/********************************************************************
created:    2019-11-16
author:     lixianmin

Copyright (C) - All Rights Reserved
*********************************************************************/

type Client struct {
	session *road.Session
	server  *Server
	topics  map[string]struct{}
}

// newClient 创建一个新的client对象
func newClient(server *Server, session *road.Session) *Client {
	var client = &Client{
		session: session,
		server:  server,
		topics:  make(map[string]struct{}),
	}

	return client
}

//func (client *Client) PushHtml(html string) {
//	var bean = beans.HtmlResponse{
//		Html: html,
//	}
//
//	var jsonBytes, err = tools.MarshalUnescape(bean)
//	if err == nil {
//		err = client.session.Push("console.html", jsonBytes)
//		if err != nil {
//			logger.Info("err=%q", err)
//		}
//	} else {
//		logger.Warn("Can not marshal bean=%v, err=%s", bean, err)
//	}
//}

func (client *Client) PushDefault(v interface{}) {
	const key = "console.default" // 参考console.html中的 onDefault()方法
	_ = client.session.Push(key, v)
}

func (client *Client) OnClosed(callback func()) {
	client.session.OnClosed(callback)
}

func (client *Client) Session() *road.Session {
	return client.session
}

func (client *Client) Attachment() *road.Attachment {
	return client.session.Attachment()
}
