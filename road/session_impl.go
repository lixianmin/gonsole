package road

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
created:    2020-08-28
author:     lixianmin

Copyright (C) - All Rights Reserved
*********************************************************************/

var (
	globalIdGenerator int64 = 0
)

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
	attachment *AttachmentImpl
	wc         loom.WaitClose
	serde      serde.Serde

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
		attachment: &AttachmentImpl{},
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

// Close 可以被多次调用，但只触发一次OnClosed事件
func (my *sessionImpl) Close() error {
	return my.wc.Close(func() error {
		var err = my.link.Close()
		my.attachment.dispose()

		var handler = my.onClosedHandler
		if handler != nil {
			handler()
		}

		return err
	})
}

func (my *sessionImpl) OnReceivingPacket(handler func(pack serde.Packet) error) {
	my.onReceivingPacketHandler = handler
}

// OnClosed 需要保证OnClosed事件在任何情况下都会有且仅有一次触发：无论是主动断开，还是意外断开链接；无论client端有没有因为网络问题收到回复消息
func (my *sessionImpl) OnClosed(handler func()) {
	my.onClosedHandler = handler
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
