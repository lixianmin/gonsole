package gonsole

//
//import (
//	"github.com/lixianmin/gonsole/road"
//)
//
///********************************************************************
//created:    2019-11-16
//author:     lixianmin
//
//Copyright (C) - All Rights Reserved
//*********************************************************************/
//
//type Client struct {
//	session road.Session
//	topics  map[string]struct{}
//}
//
//// newClient 创建一个新的client对象
//func newClient(session road.Session) *Client {
//	var client = &Client{
//		session: session,
//		topics:  make(map[string]struct{}),
//	}
//
//	return client
//}
//
////func (client *Client) PushHtml(html string) {
////	var bean = beans.HtmlResponse{
////		Html: html,
////	}
////
////	var jsonBytes, err = tools.MarshalUnescape(bean)
////	if err == nil {
////		err = client.session.Push("console.html", jsonBytes)
////		if err != nil {
////			logo.Info("err=%q", err)
////		}
////	} else {
////		logo.Warn("Can not marshal bean=%v, err=%s", bean, err)
////	}
////}
//
//func (client *Client) PushDefault(v any) {
//	_ = client.session.Send("console.default", v)
//}
//
//func (client *Client) OnClosed(callback func()) {
//	client.session.OnClosed(callback)
//}
//
//func (client *Client) Session() road.Session {
//	return client.session
//}
//
//func (client *Client) Attachment() road.Attachment {
//	return client.session.Attachment()
//}
