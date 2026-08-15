package intern

import (
	"net"
	"sync/atomic"
	"time"
)

/********************************************************************
created:    2022-11-27
author:     lixianmin

Copyright (C) - All Rights Reserved
*********************************************************************/

type commonLink struct {
	conn     net.Conn
	isClosed int32
}

func (my *commonLink) resetReadDeadline(kickInterval time.Duration) {
	// 不知道为什么经常报:  close session(27) by err="read tcp 192.168.31.96:8888->192.168.31.96:59562: i/o timeout"
	// 1. 改为SetDeadline(),即同时改写read/write的deadline也是无效的
	// 2. 时间已经设置为3倍的heartbeat时间了
	//
	// i/o timeout有可能是chrome的throttling mechanism导致的, 当chrome tab在background的时候, setInterval()的调用间隔可能会放大到1min, 就很容易超时了
	// 因为玩家可能切游戏到后台很久去做其它的事情, 因此这个值必须要大一些, 太短很容易被服务器踢的
	_ = my.conn.SetReadDeadline(time.Now().Add(kickInterval))
}

// Close closes the connection.
// Any blocked Read or Write operations will be unblocked and return errors.
func (my *commonLink) Close() error {
	atomic.StoreInt32(&my.isClosed, 1)

	// bugfix: 重构引入commonLink后，Close()只设置isClosed标志、没有关闭底层conn，导致：
	//  1. app.go goLoop中扫描器连接调用Close()后socket仍然敞开，且这些连接不会启动GoLoop，
	//     fd永远不会被回收（扫描器大量连入会耗尽fd）
	//  2. Kick()/session.Close()后，receive goroutine仍阻塞在Read()上，直到kickInterval超时
	//     （默认1分钟）才由GoLoop的defer真正关闭，session不是立即关闭的
	// 必须关闭底层conn才能unblock阻塞中的Read。GoLoop的defer中也会再次Close()，
	// net.Conn.Close()重复调用是安全的（返回error但无副作用）
	return my.conn.Close()
}

// RemoteAddr returns the remote address.
func (my *commonLink) RemoteAddr() net.Addr {
	return my.conn.RemoteAddr()
}
