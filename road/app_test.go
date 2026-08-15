package road

import (
	"fmt"
	"net"
	"testing"
	"time"

	"github.com/lixianmin/gonsole/road/epoll"
)

/********************************************************************
created:    2026-08-10
author:     xmli

bugfix回归：扫描器连接必须被真正关闭（fd泄漏）。
Red: 旧实现 app.go goLoop中 IsScanner(ip)==true 时仅 continue，不关闭连接；
     且 commonLink.Close() 只置标志不关底层conn → 扫描器连接永不关闭，fd泄漏
Green: 扫描器路径显式 conn.Close()，且 Close() 真正关闭底层conn → 客户端读到EOF
Copyright (C) - All Rights Reserved
*********************************************************************/

func TestAppScannerConnectionClosed(t *testing.T) {
	var address = fmt.Sprintf("127.0.0.1:%d", 0) // 端口0由listener自动分配？TcpAcceptor用net.Listen(address)…
	// TcpAcceptor不支持端口0自动分配（address直接传给net.Listen，0会随机分配但GetLinkChan不可知），
	// 因此这里监听一个随机端口
	var listener, err = net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	var realPort = listener.Addr().(*net.TCPAddr).Port
	_ = listener.Close() // 释放端口，让acceptor重新监听

	address = fmt.Sprintf("127.0.0.1:%d", realPort)
	var acceptor = epoll.NewTcpAcceptor(address)
	var app = NewApp(acceptor)
	app.Listen()
	defer acceptor.Close()

	// 建立 10 个连接且不发握手包：第10个连接触发 ScanDefender 的 scanner 判定
	var conns = make([]net.Conn, 0, 10)
	defer func() {
		for _, conn := range conns {
			_ = conn.Close()
		}
	}()

	// TcpAcceptor.Listen()是异步起listener，dial需要重试等待bind完成
	var dialConn func() (net.Conn, error)
	dialConn = func() (net.Conn, error) {
		var conn, err2 = net.Dial("tcp", address)
		if err2 != nil {
			return nil, err2
		}
		return conn, nil
	}

	for i := 0; i < 10; i++ {
		var conn net.Conn
		var err2 error
		for attempt := 0; attempt < 40; attempt++ {
			conn, err2 = dialConn()
			if err2 == nil {
				break
			}
			time.Sleep(50 * time.Millisecond)
		}
		if err2 != nil {
			t.Fatalf("dial #%d: %v", i, err2)
		}
		conns = append(conns, conn)
	}

	// 等待服务端处理完（轮询，最多3秒）：第10个连接应被关闭（读返回EOF/错误）
	// 注意：Read超时（i/o timeout）不代表连接被关闭，必须区分
	var last = conns[9]
	var deadline = time.Now().Add(3 * time.Second)
	var buf = make([]byte, 32)
	for time.Now().Before(deadline) {
		_ = last.SetReadDeadline(time.Now().Add(200 * time.Millisecond))
		var _, err3 = last.Read(buf)
		if err3 != nil {
			if netErr, ok := err3.(net.Error); ok && netErr.Timeout() {
				continue // 读超时：连接仍打开，继续轮询
			}
			// 连接被服务端关闭（EOF / use of closed network connection）
			return // PASS
		}
		// 可能先读到服务端的heartbeat探测包，继续读直到EOF
	}

	t.Fatal("scanner connection not closed within 3s (bug: fd泄漏, 扫描器连接永不关闭)")
}
