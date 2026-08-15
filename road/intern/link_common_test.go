package intern

import (
	"net"
	"testing"
	"time"

	"github.com/lixianmin/got/iox"
)

/********************************************************************
created:    2026-08-10
author:     xmli

bugfix回归：commonLink.Close()必须真正关闭底层conn。
重构引入commonLink后Close()只设置isClosed标志，导致：
  1. 扫描器连接（app.go goLoop中Close后不启动GoLoop）的fd永远不会被回收
  2. Kick()/session.Close()后receive goroutine仍阻塞在Read()上，直到kickInterval超时
     （默认1分钟）才真正关闭
本测试验证：Close()后，阻塞中的Read必须立即返回（而不是等到deadline）
Copyright (C) - All Rights Reserved
*********************************************************************/

func TestCommonLinkCloseUnblocksRead(t *testing.T) {
	var listener, err = net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer listener.Close()

	var acceptChan = make(chan net.Conn, 1)
	go func() {
		var conn, err2 = listener.Accept()
		if err2 == nil {
			acceptChan <- conn
		}
	}()

	var clientConn, err3 = net.Dial("tcp", listener.Addr().String())
	if err3 != nil {
		t.Fatal(err3)
	}
	defer clientConn.Close()

	var serverConn = <-acceptChan
	defer serverConn.Close()

	var link = NewTcpLink(serverConn)

	// GoLoop会阻塞在Read上（deadline=1分钟）。Close()必须unblock这个Read
	var readReturned = make(chan struct{})
	go func() {
		link.GoLoop(time.Minute, func(reader *iox.OctetsReader, err error) {
			close(readReturned)
		})
	}()

	// 等Read真正进入阻塞
	time.Sleep(100 * time.Millisecond)

	var start = time.Now()
	_ = link.Close()

	select {
	case <-readReturned:
		var elapsed = time.Since(start)
		if elapsed > 2*time.Second {
			t.Fatalf("read unblocked after %v, want immediate (bug: Close只设置标志不关闭conn)", elapsed)
		}
	case <-time.After(3 * time.Second):
		t.Fatal("read still blocked after Close() (bug: Close只设置标志不关闭conn, fd泄漏+session不立即关闭)")
	}
}
