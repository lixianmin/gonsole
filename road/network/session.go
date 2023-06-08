package network

import (
	"context"
	"github.com/lixianmin/gonsole/ifs"
	"github.com/lixianmin/gonsole/road/serde"
	"github.com/lixianmin/got/iox"
	"github.com/lixianmin/got/loom"
	"github.com/lixianmin/logo"
	"net"
	"reflect"
	"runtime"
	"sync"
	"sync/atomic"
)

/********************************************************************
created:    2022-04-07
author:     lixianmin

Copyright (C) - All Rights Reserved
*********************************************************************/

type Session interface {
	Handshake() error // server主动向client发送服务器的配置信息
	Kick() error      // server主动踢client
	PushByRoute(route string, v interface{}) error
	PushByKind(kind int32, v interface{}) error
	Close() error

	OnReceivingPacket(handler func(pack serde.Packet) error)
	OnClosed(handler func())

	Id() int64
	RemoteAddr() net.Addr
	Attachment() *Attachment
}

type sessionWrapper struct {
	*sessionImpl
}

type sessionImpl struct {
	manger     *Manager
	writer     *iox.OctetsWriter
	writeLock  sync.Mutex
	id         int64
	link       Link
	ctxValue   reflect.Value
	attachment *Attachment
	wc         loom.WaitClose

	onReceivingPacketHandler func(packet serde.Packet) error
	onClosedHandler          func()
}

func newSession(manager *Manager, link Link) Session {
	var id = atomic.AddInt64(&globalIdGenerator, 1)
	var my = &sessionWrapper{&sessionImpl{
		manger:     manager,
		writer:     iox.NewOctetsWriter(&iox.OctetsStream{}),
		id:         id,
		link:       link,
		attachment: &Attachment{},
	}}

	logo.Info("create session(%d)", my.id)
	my.ctxValue = reflect.ValueOf(context.WithValue(context.Background(), ifs.CtxKeySession, my))
	my.startGoLoop()

	// 参考: https://zhuanlan.zhihu.com/p/76504936
	runtime.SetFinalizer(my, func(w *sessionWrapper) {
		_ = w.Close()
	})

	return my
}
