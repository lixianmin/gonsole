package gonsole

import (
	"net"
	"testing"

	"github.com/lixianmin/gonsole/road"
)

/********************************************************************
created:    2026-08-10
author:     xmli

bugfix回归：Topic.addClient 对 nil map 赋值会panic。
Red: 旧实现 start() 才初始化 sessions.d；未start的Topic（或热更注册时序）
     直接 addClient → nil map assignment panic
Green: addClient 懒初始化 + start() 持锁初始化
Copyright (C) - All Rights Reserved
*********************************************************************/

type fakeSession struct {
	attachment road.Attachment
}

func (s *fakeSession) Handshake() error               { return nil }
func (s *fakeSession) Kick(reason string) error       { return nil }
func (s *fakeSession) Send(route string, v any) error { return nil }
func (s *fakeSession) Echo(handler func())            {}
func (s *fakeSession) OnHandShaken(handler func())    {}
func (s *fakeSession) OnClosed(handler func())        {}
func (s *fakeSession) Id() int64                      { return 0 }
func (s *fakeSession) RemoteAddr() net.Addr {
	return &net.TCPAddr{IP: net.IPv4(127, 0, 0, 1), Port: 8888}
}
func (s *fakeSession) Attachment() road.Attachment { return s.attachment }
func (s *fakeSession) Nonce() int32                { return 0 }

// TestTopicAddClient_NilMapNoPanic
// Red: 未start的Topic（sessions.d==nil），addClient直接赋值 → panic
// Green: 懒初始化，不panic；nil session安全跳过；removeClient安全
func TestTopicAddClient_NilMapNoPanic(t *testing.T) {
	var topic = &Topic{Name: "top"} // 不经过TopicManager.Register，start()未执行 → d==nil
	var session = &fakeSession{attachment: &road.AttachmentImpl{}}

	topic.addClient(session) // 旧实现这里panic: assignment to entry in nil map
	topic.addClient(nil)     // nil session应安全跳过

	// 验证订阅生效
	topic.sessions.RLock()
	var _, ok = topic.sessions.d[session]
	topic.sessions.RUnlock()
	if !ok {
		t.Fatal("expected session to be subscribed")
	}

	topic.removeClient(session) // nil map上delete是安全的
	topic.removeClient(nil)

	topic.sessions.RLock()
	_, ok = topic.sessions.d[session]
	topic.sessions.RUnlock()
	if ok {
		t.Fatal("expected session to be unsubscribed after removeClient")
	}
}
