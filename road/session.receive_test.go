package road

import (
	"errors"
	"net"
	"sync"
	"testing"
	"time"

	"github.com/lixianmin/gonsole/road/component"
	"github.com/lixianmin/gonsole/road/intern"
	"github.com/lixianmin/gonsole/road/serde"
)

/********************************************************************
created:    2026-08-10
author:     xmli

bugfix回归：
1. TestOnReceivedUserdata_InterceptorPlainError — interceptor返回普通error时不能panic
   （原实现 err1.(*Error) 无comma-ok断言，普通error直接panic）
2. TestSessionConcurrentSendAndSetSerde — Send()与setSerde()并发时的data race
   （原实现 my.serde 字段无锁读写；my.routeKinds 在Send中无锁读取）
Copyright (C) - All Rights Reserved
*********************************************************************/

type fakeLink struct{}

func (f *fakeLink) GoLoop(kickInterval time.Duration, onReadHandler intern.OnReadHandler) {}
func (f *fakeLink) Write(data []byte) (int, error)                                        { return len(data), nil }
func (f *fakeLink) Close() error                                                          { return nil }
func (f *fakeLink) RemoteAddr() net.Addr                                                  { return &net.TCPAddr{IP: net.IPv4(127, 0, 0, 1), Port: 8888} }

func newTestSession(t *testing.T, my *Manager) *sessionImpl {
	t.Helper()
	var session = newSession(my, &fakeLink{}, my.idGenerator.Add(1))
	return session.(*sessionWrapper).sessionImpl
}

// TestOnReceivedUserdata_InterceptorPlainError
// Red: 旧实现 `err1.(*Error)` 对普通error直接panic，session被panic杀死
// Green: comma-ok断言，普通error按正常错误路径返回
func TestOnReceivedUserdata_InterceptorPlainError(t *testing.T) {
	var my = newManager(3*time.Second, time.Minute)
	my.AddHandler("plain.err", &component.Handler{})
	my.RebuildHandlerKinds()
	my.AddInterceptor(func(session Session, route string) error {
		return errors.New("plain error, not *road.Error")
	})

	var session = newTestSession(t, my)
	session.setSerde(&serde.JsonSerde{})

	var kind = my.CloneRouteKinds()["plain.err"]
	var err = session.onReceivedUserdata(serde.Packet{Kind: kind})
	if err == nil {
		t.Fatal("expected interceptor error to be returned")
	}
	if err.Error() != "plain error, not *road.Error" {
		t.Fatalf("unexpected error: %v", err)
	}
}

// TestSessionConcurrentSendAndSetSerde
// Red: my.serde（receive线程写，其它goroutine读）与 my.routeKinds（sendRouteKind持锁替换，
//
//	Send无锁读取）均为未同步的字段访问，-race必然报race
//
// Green: serde通过getSerde/setSerde持writeLock访问，Send的读在writeLock内
func TestSessionConcurrentSendAndSetSerde(t *testing.T) {
	var my = newManager(3*time.Second, time.Minute)
	var session = newTestSession(t, my)
	// 预置一个serde，避免Send全部提前返回ErrNilSerde（提前返回也要读mySerde，仍会触发race）
	session.setSerde(&serde.JsonSerde{})

	var stop = make(chan struct{})
	var wg sync.WaitGroup

	// goroutine A：模拟receive线程反复写入serde（握手完成后一般只写一次，这里放大竞争窗口）
	wg.Add(1)
	go func() {
		defer wg.Done()
		for {
			select {
			case <-stop:
				return
			default:
				session.setSerde(&serde.JsonSerde{})
			}
		}
	}()

	// goroutine B/C：模拟topic推送/业务goroutine并发Send（读serde + 读routeKinds + 动态注册）
	for i := 0; i < 2; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for {
				select {
				case <-stop:
					return
				default:
					_ = session.Send("console.stream", "hello")
				}
			}
		}()
	}

	time.Sleep(200 * time.Millisecond)
	close(stop)
	wg.Wait()
}
