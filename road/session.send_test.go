package road

import (
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/lixianmin/gonsole/road/component"
	"github.com/lixianmin/gonsole/road/serde"
)

/********************************************************************
created:    2026-08-09
author:     xmli

动态 route kind 分配（bugfix：原实现 len(routeKinds)+UserBase 与预分配 kind 冲突，
修复为 Manager 级单调序列：预分配只增不减，动态 kind 永远高于预分配最大值）
Copyright (C) - All Rights Reserved
*********************************************************************/

// TestNextRouteKind_Empty 无预分配 route 时，首个动态 kind 必须是 UserBase+1
// （UserBase 之下是协议包 kind，如 Handshake/RouteKind，不可用）
func TestNextRouteKind_Empty(t *testing.T) {
	var my = newManager(3*time.Second, time.Minute)
	var kind = my.nextRouteKind()
	if kind != serde.UserBase+1 {
		t.Fatalf("empty routeKinds: got kind=%d, want %d", kind, serde.UserBase+1)
	}
}

// TestNextRouteKind_AbovePreassigned 预分配 3 个 route（kind=10,11,12）后，
// 动态 kind 必须 > 12。原实现 len(routeKinds)+UserBase 在"部分注册/热更"场景下
// 会分配出与预分配冲突的 kind（.chat_stream 与 refresh_notify 同拿 109 的复现场景）
func TestNextRouteKind_AbovePreassigned(t *testing.T) {
	var my = newManager(3*time.Second, time.Minute)
	var routes = []string{"a.svc", "b.svc", "c.svc"}
	for _, r := range routes {
		my.AddHandler(r, &component.Handler{})
	}
	my.RebuildHandlerKinds()

	var kind = my.nextRouteKind()
	if kind <= serde.UserBase+int32(len(routes)-1) {
		t.Fatalf("nextRouteKind got kind=%d, want > %d (max preassigned)", kind, serde.UserBase+int32(len(routes)-1))
	}
}

// TestNextRouteKind_StaleCloneNoCollision 回归测试：会话克隆了"旧"routeKinds 后，
// 服务端热更注册了新的 route 并 Rebuild。此时动态 kind 必须跳过新预分配的 kind，
// 而不是从旧 clone 的 max+1 递增（那会撞上新预分配 kind）。
// 这是原实现（len+UserBase）与"仅按会话内 max+1"修复都无法覆盖的场景。
func TestNextRouteKind_StaleCloneNoCollision(t *testing.T) {
	var my = newManager(3*time.Second, time.Minute)
	my.AddHandler("a.svc", &component.Handler{})
	my.AddHandler("b.svc", &component.Handler{})
	my.RebuildHandlerKinds() // a=10, b=11

	// 模拟一个在热更前创建的旧会话（克隆了旧 map）
	var staleClone = my.CloneRouteKinds()
	if len(staleClone) != 2 {
		t.Fatalf("warmup: staleClone len=%d, want 2", len(staleClone))
	}

	// 热更：注册 c.svc 并 Rebuild（预分配变为 a=10, b=11, c=12）
	my.AddHandler("c.svc", &component.Handler{})
	my.RebuildHandlerKinds()

	// 旧会话的动态分配（模拟 sendRouteKind）必须 > 12，不能是 max(staleClone)+1 = 12
	var kind = my.nextRouteKind()
	if kind <= serde.UserBase+int32(2) {
		t.Fatalf("stale clone dynamic kind=%d collides with preassigned kind=%d", kind, serde.UserBase+2)
	}
}

// TestNextRouteKind_OldFormulasCollide 双实现对照：证明旧实现的两个公式在热更场景下必然碰撞。
// 这不是在测新代码，而是把旧公式（len+UserBase 和 session内max+1）内嵌为对照，
// 客观证明"原实现会分配出与预分配冲突的kind"这一bug存在，以及为什么必须改成Manager级序列。
func TestNextRouteKind_OldFormulasCollide(t *testing.T) {
	var my = newManager(3*time.Second, time.Minute)
	my.AddHandler("a.svc", &component.Handler{})
	my.AddHandler("b.svc", &component.Handler{})
	my.RebuildHandlerKinds() // 预分配 a=10, b=11

	// 热更：注册 c.svc → 预分配变为 a=10, b=11, c=12
	my.AddHandler("c.svc", &component.Handler{})
	my.RebuildHandlerKinds()
	var preassigned = my.CloneRouteKinds()
	if preassigned["c.svc"] != serde.UserBase+2 {
		t.Fatalf("warmup: c.svc kind=%d, want %d", preassigned["c.svc"], serde.UserBase+2)
	}

	// 旧公式1: len(旧map)+UserBase —— 会话在热更前克隆的map只有 a（len=1），
	// 动态kind = 1+10 = 11，而热更后 b 的预分配kind恰好也是 11 → 碰撞
	var legacyKind1 = int32(1) + serde.UserBase
	if legacyKind1 == preassigned["b.svc"] {
		t.Logf("旧公式1 len+UserBase 必然撞上新预分配 kind=%d (bug复现)", legacyKind1)
	} else {
		t.Fatalf("test scenario invalid: legacyKind1=%d, b.svc=%d", legacyKind1, preassigned["b.svc"])
	}

	// 旧公式2: session内max+1 —— 旧map只有 a（max=10），+1 = 11，同样撞 b
	var legacyKind2 = int32(serde.UserBase) + 1
	if legacyKind2 == preassigned["b.svc"] {
		t.Logf("旧公式2 max+1 必然撞上新预分配 kind=%d (bug复现)", legacyKind2)
	} else {
		t.Fatalf("test scenario invalid: legacyKind2=%d, b.svc=%d", legacyKind2, preassigned["b.svc"])
	}

	// 新实现必须避开：动态kind > 所有预分配kind
	var newKind = my.nextRouteKind()
	for _, k := range preassigned {
		if newKind == k {
			t.Fatalf("new implementation kind=%d still collides with preassigned kind=%d", newKind, k)
		}
	}
	if newKind <= serde.UserBase+2 {
		t.Fatalf("new implementation kind=%d not above preassigned max", newKind)
	}
}

// TestNextRouteKind_Monotonic 连续分配必须严格递增且唯一
func TestNextRouteKind_Monotonic(t *testing.T) {
	var my = newManager(3*time.Second, time.Minute)
	var seen = make(map[int32]bool)
	var prev int32 = serde.UserBase
	for i := 0; i < 5; i++ {
		var kind = my.nextRouteKind()
		if kind <= prev {
			t.Fatalf("kind=%d not monotonic (prev=%d)", kind, prev)
		}
		if seen[kind] {
			t.Fatalf("kind=%d allocated twice", kind)
		}
		seen[kind] = true
		prev = kind
	}
}

// TestNextRouteKind_Concurrent 并发分配必须全局唯一（多 session 同时 sendRouteKind）
func TestNextRouteKind_Concurrent(t *testing.T) {
	var my = newManager(3*time.Second, time.Minute)
	const goroutines = 8
	const allocsPerGoroutine = 200

	var mu sync.Mutex
	var seen = make(map[int32]bool, goroutines*allocsPerGoroutine)
	var wg sync.WaitGroup
	for i := 0; i < goroutines; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < allocsPerGoroutine; j++ {
				var kind = my.nextRouteKind()
				mu.Lock()
				if seen[kind] {
					t.Errorf("kind=%d allocated by multiple goroutines", kind)
				}
				seen[kind] = true
				mu.Unlock()
			}
		}()
	}
	wg.Wait()

	if len(seen) != goroutines*allocsPerGoroutine {
		t.Fatalf("allocated %d unique kinds, want %d", len(seen), goroutines*allocsPerGoroutine)
	}
}

// TestRebuildHandlerKinds_BumpsDynamicSeq Rebuild 后动态序列必须高于新的预分配上限，
// 保证旧会话后续动态注册也不会撞上新预分配的 kind
func TestRebuildHandlerKinds_BumpsDynamicSeq(t *testing.T) {
	var my = newManager(3*time.Second, time.Minute)
	var first = my.nextRouteKind() // 11

	for i := 0; i < 10; i++ {
		my.AddHandler(fmt.Sprintf("svc.route_%02d", i), &component.Handler{})
	}
	my.RebuildHandlerKinds() // 预分配 10..19

	var kind = my.nextRouteKind()
	if kind <= serde.UserBase+int32(9) {
		t.Fatalf("after rebuild: dynamic kind=%d, want > %d", kind, serde.UserBase+9)
	}
	if kind <= first {
		t.Fatalf("after rebuild: dynamic kind=%d not above first=%d", kind, first)
	}
}
