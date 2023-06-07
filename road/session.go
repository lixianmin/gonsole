package road

import (
	"context"
	"github.com/lixianmin/gonsole/ifs"
	"github.com/lixianmin/gonsole/road/epoll"
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
	ShakeHand(heartbeat float32) error // 心跳间隔. 单位: 秒
	Push(route string, v interface{}) error
	Kick() error

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
	conn       epoll.IConn
	ctxValue   reflect.Value
	attachment *Attachment
	wc         loom.WaitClose
	onClosed   delegate
}

func NewSession(manager *NetManager, conn epoll.IConn) Session {
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
