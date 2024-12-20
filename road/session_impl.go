package road

import (
	"context"
	"net"
	"reflect"
	"sync"
	"sync/atomic"

	"github.com/lixianmin/gonsole/ifs"
	"github.com/lixianmin/gonsole/road/intern"
	"github.com/lixianmin/gonsole/road/serde"
	"github.com/lixianmin/got/iox"
	"github.com/lixianmin/got/loom"
	"github.com/lixianmin/logo"
)

/********************************************************************
created:    2020-08-28
author:     lixianmin

Copyright (C) - All Rights Reserved
*********************************************************************/

var (
	globalIdGenerator atomic.Int64
	echoIdGenerator   atomic.Int32
)

type sessionWrapper struct {
	*sessionImpl
}

type sessionImpl struct {
	manager    *Manager
	writer     *iox.OctetsWriter
	writeLock  sync.Mutex
	id         int64
	link       intern.Link
	ctxValue   reflect.Value
	attachment *AttachmentImpl
	wc         loom.WaitClose
	serde      serde.Serde
	routeKinds map[string]int32

	handlerLock          sync.Mutex
	onHandShakenHandlers []func()
	onClosedHandlers     []func()
	echoHandlers         map[int32]func()
}

func newSession(manager *Manager, link intern.Link) Session {
	var id = globalIdGenerator.Add(1)
	var routeKinds = manager.CloneRouteKinds()
	var my = &sessionWrapper{&sessionImpl{
		manager:      manager,
		writer:       iox.NewOctetsWriter(&iox.OctetsStream{}),
		id:           id,
		link:         link,
		attachment:   &AttachmentImpl{},
		routeKinds:   routeKinds,
		echoHandlers: map[int32]func(){},
	}}

	// 线上有大量的非法请求, 感觉是攻击, 先用Debug输出吧, 否则会生成大量无效日志
	logo.Info("create session(%d), addr=%s", my.id, my.link.RemoteAddr())
	var ctx = context.WithValue(context.Background(), keySession, my.sessionImpl)
	my.attachment.Set(ifs.KeyContext, ctx)

	my.ctxValue = reflect.ValueOf(ctx)
	my.startGoLoop()

	// 这个设计, 可以保证finalizer被调用到, 但极大延长了对象在内存中存活的时间, 导致内存上涨很快. 外网扫描器很多, 有可能导致内存OOM
	//// 参考: https://zhuanlan.zhihu.com/p/76504936
	//runtime.SetFinalizer(my, func(w *sessionWrapper) {
	//	_ = w.Close()
	//})

	return my
}

// Close 可以被多次调用，但只触发一次OnClosed事件
func (my *sessionImpl) Close() error {
	return my.wc.Close(func() error {
		var err = my.link.Close()
		my.attachment.dispose()
		my.onEventClosed()
		return err
	})
}

func (my *sessionImpl) onEventClosed() {
	my.handlerLock.Lock()
	defer my.handlerLock.Unlock()
	{
		for _, handler := range my.onClosedHandlers {
			handler()
		}
		my.onClosedHandlers = nil
	}
}

func (my *sessionImpl) OnHandShaken(handler func()) {
	if handler != nil {
		my.handlerLock.Lock()
		defer my.handlerLock.Unlock()

		my.onHandShakenHandlers = append(my.onHandShakenHandlers, handler)
	}
}

// OnClosed 需要保证OnClosed事件在任何情况下都会有且仅有一次触发：无论是主动断开，还是意外断开链接；无论client端有没有因为网络问题收到回复消息
func (my *sessionImpl) OnClosed(handler func()) {
	if handler != nil {
		my.handlerLock.Lock()
		defer my.handlerLock.Unlock()

		my.onClosedHandlers = append(my.onClosedHandlers, handler)
	}
}

// 在session中加入SendCallback()的相关权衡？
// 为什么要删除？ 因为使用SendCallback()的往往都是异步IO，然而异步IO往往都会卡session，所以别用
//
// 正面：
// 1. 异步转同步
// 2. 分帧削峰
// 3. player类不再需要独立的goroutine，至少节约2~8KB的goroutine初始内存，这比TaskQueue占用的内存要多得多
//
// 负面：
// 1. 这个方法只对业务有可能有用，但对网络库本身并没有意义；
// 2. 必须谨慎使用，过长的处理时间会影响后续网络消息处理，可能导致链接超时（当然你可以选择不用）
//func (my *sessionImpl) SendCallback(handler taskx.Handler) loom.ITask {
//	return my.tasks.SendCallback(handler)
//}
//
//// 延迟任务
//func (my *sessionImpl) SendDelayed(delayed time.Duration, handler taskx.Handler) {
//	my.tasks.SendDelayed(delayed, handler)
//}

// Id 全局唯一id
func (my *sessionImpl) Id() int64 {
	return my.id
}

func (my *sessionImpl) RemoteAddr() net.Addr {
	return my.link.RemoteAddr()
}

func (my *sessionImpl) Attachment() Attachment {
	return my.attachment
}

func (my *sessionImpl) Nonce() int32 {
	return my.attachment.Int32(keyNonce)
}
