package road

import (
	"github.com/lixianmin/gonsole/road/epoll"
	"github.com/lixianmin/got/loom"
	"github.com/lixianmin/logo"
	"golang.org/x/time/rate"
	"net"
	"runtime"
	"sync/atomic"
	"time"
)

/********************************************************************
created:    2022-04-07
author:     lixianmin

Copyright (C) - All Rights Reserved
*********************************************************************/

type Session interface {
	Push(route string, v interface{}) error
	Kick() error

	OnHandShaken(handler func())
	OnClosed(handler func())

	Id() int64
	RemoteAddr() net.Addr
	Attachment() *Attachment
}

type sessionWrapper struct {
	*sessionImpl
}

type sessionImpl struct {
	app         *App
	id          int64
	conn        epoll.IConn
	attachment  *Attachment
	rateLimiter *rate.Limiter
	wc          loom.WaitClose

	onHandShaken delegate
	onClosed     delegate
}

func NewSession(app *App, conn epoll.IConn) Session {
	var id = atomic.AddInt64(&globalIdGenerator, 1)
	var my = &sessionWrapper{&sessionImpl{
		app:         app,
		id:          id,
		conn:        conn,
		attachment:  &Attachment{},
		rateLimiter: rate.NewLimiter(rate.Every(time.Second), app.rateLimitBySecond),
	}}

	logo.Info("create session(%d)", my.id)
	my.initialize()
	loom.Go(my.goSessionLoop)

	// 参考: https://zhuanlan.zhihu.com/p/76504936
	runtime.SetFinalizer(my, func(w *sessionWrapper) {
		_ = w.Close()
	})

	return my
}
