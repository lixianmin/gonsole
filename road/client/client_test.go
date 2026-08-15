package client

import (
	"net"
	"sync"
	"testing"
	"time"

	"github.com/lixianmin/gonsole/road"
	"github.com/lixianmin/gonsole/road/serde"
)

/********************************************************************
created:    2026-08-10
author:     xmli

bugfix回归：Client的routeKinds/kindRoutes/requestHandlers/registeredHandlers/nonce
被两个goroutine并发访问（goLoop goroutine写入、用户goroutine读写），
原实现完全无锁 → -race必然报race；修复后全部受mapLock保护。
Copyright (C) - All Rights Reserved
*********************************************************************/

// fakeConn 只覆盖 Write/Close，其它方法由嵌入的 net.Conn 提供（不会被调用）
type fakeConn struct {
	net.Conn
}

func (f *fakeConn) Write(b []byte) (int, error) { return len(b), nil }
func (f *fakeConn) Close() error                { return nil }

// TestClientConcurrentAccess
// Red: 原实现无mapLock，map并发读写-race必报
// Green: mapLock保护后无race
func TestClientConcurrentAccess(t *testing.T) {
	var my = NewClient()
	// 预置serde与conn，让Request/Send能走到map读写（不依赖Connect的网络握手）
	my.serde = &serde.JsonSerde{}
	my.conn = &fakeConn{}

	var stop = make(chan struct{})
	var wg sync.WaitGroup

	// goroutine A：模拟goLoop goroutine处理RouteKind/userdata（写routeKinds/kindRoutes，读registeredHandlers）
	wg.Add(1)
	go func() {
		defer wg.Done()
		var index int32 = 0
		for {
			select {
			case <-stop:
				return
			default:
				index++
				var pack = serde.Packet{Kind: serde.UserBase + index%100, RequestId: 0, Data: []byte(`{"kind":10,"route":"console.stream"}`)}
				_ = my.onReceivedRouteKind(pack)
				_ = my.onReceivedUserdata(serde.Packet{Kind: serde.UserBase + index%100, RequestId: index})
			}
		}
	}()

	// goroutine B/C：模拟用户goroutine调用Request/Send/On（读写routeKinds、写requestHandlers/registeredHandlers）
	for i := 0; i < 2; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			var index int32 = 0
			for {
				select {
				case <-stop:
					return
				default:
					index++
					var response struct{ Text string }
					_ = my.Request("console.command", map[string]string{"command": "help"}, &response, func(err *road.Error) {})
					_ = my.Send("console.stream", "hello")
					_ = my.On("console.stream", &response, func(err *road.Error) {})
					_ = my.Nonce()
				}
			}
		}()
	}

	time.Sleep(200 * time.Millisecond)
	close(stop)
	wg.Wait()
}
