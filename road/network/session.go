package network

import (
	"context"
	"github.com/lixianmin/gonsole/ifs"
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
	Handshake() error                       // server主动向client发送服务器的配置信息
	Kick() error                            // server主动踢client
	Push(route string, v interface{}) error // 推送数据到对端

	OnClosed(handler func())

	Id() int64
	RemoteAddr() net.Addr
	Attachment() *Attachment
}

type sessionWrapper struct {
	*sessionImpl
}

type sessionImpl struct {
	manger     *NetManager
	writer     *iox.OctetsWriter
	writeLock  sync.Mutex
	id         int64
	conn       Connection
	ctxValue   reflect.Value
	attachment *Attachment
	wc         loom.WaitClose
	onClosed   delegate
}

func newSession(manager *NetManager, conn Connection) Session {
	var id = atomic.AddInt64(&globalIdGenerator, 1)
	var my = &sessionWrapper{&sessionImpl{
		manger:     manager,
		writer:     iox.NewOctetsWriter(&iox.OctetsStream{}),
		id:         id,
		conn:       conn,
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
