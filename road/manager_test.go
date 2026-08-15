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

RebuildHandlerKinds 与 NewSession 并发的半填充窗口（bugfix：动态 route kind 撞预分配）
Copyright (C) - All Rights Reserved
*********************************************************************/

// cloneIsComplete 校验 clone 的 routeKinds 是"完整"的：kind 数量与 routeHandlers 一致，
// 且 kind 空间连续无空洞（len 与 kind 值一致：kind = UserBase + index，0<=index<len）。
// 原实现 Rebuild 先赋值空 map 再填充，并发 clone 会拿到半填充 map（数量不足/空洞），
// 该 session 的 sendRouteKind(len+UserBase) 动态 kind 会撞上预分配中未填充的 kind。
func cloneIsComplete(kinds map[string]int32, total int) bool {
	if len(kinds) != total {
		return false
	}
	// 无空洞校验：每个 kind 必须落在 [UserBase, UserBase+len-1] 且不重复
	var seen = make(map[int32]bool, len(kinds))
	for _, k := range kinds {
		if k < serde.UserBase || k >= serde.UserBase+int32(total) {
			return false
		}
		if seen[k] {
			return false
		}
		seen[k] = true
	}
	return true
}

func TestRebuildHandlerKinds_ConcurrentClone(t *testing.T) {
	// 并发复现：Rebuild 填充期间 NewSession clone。原实现（先赋值空 map 再填充）存在
	// 半填充窗口，clone 会拿到不完整/空洞的 routeKinds → 动态 kind 撞预分配。
	// 修复后（局部 map 填充完再原子赋值）clone 永远拿到完整 map。
	var my = newManager(3*time.Second, time.Minute)
	const total = 50
	for i := 0; i < total; i++ {
		my.AddHandler(fmt.Sprintf("svc.route_%02d", i), &component.Handler{})
	}
	my.RebuildHandlerKinds()

	// 预热：确认基础状态完整
	if !cloneIsComplete(my.CloneRouteKinds(), total) {
		t.Fatalf("warmup: routeKinds not complete")
	}

	var wg sync.WaitGroup
	for i := 0; i < 200; i++ {
		wg.Add(2)
		go func() { defer wg.Done(); my.RebuildHandlerKinds() }()
		go func() {
			defer wg.Done()
			var kinds = my.CloneRouteKinds()
			if !cloneIsComplete(kinds, total) {
				t.Errorf("clone got incomplete routeKinds: len=%d want=%d", len(kinds), total)
			}
		}()
	}
	wg.Wait()
}

// TestRebuildHandlerKinds_CloneAfterRebuild：Rebuild 后 clone 必须包含全部 route 且 kind 连续
func TestRebuildHandlerKinds_CloneAfterRebuild(t *testing.T) {
	var my = newManager(3*time.Second, time.Minute)
	var routes = []string{"b.svc", "a.svc", "c.svc"}
	for _, r := range routes {
		my.AddHandler(r, &component.Handler{})
	}
	my.RebuildHandlerKinds()

	var kinds = my.CloneRouteKinds()
	if len(kinds) != len(routes) {
		t.Fatalf("clone len=%d want=%d", len(kinds), len(routes))
	}
	// 排序后 a.svc=10, b.svc=11, c.svc=12
	if kinds["a.svc"] != serde.UserBase || kinds["b.svc"] != serde.UserBase+1 || kinds["c.svc"] != serde.UserBase+2 {
		t.Fatalf("unexpected kind mapping: %+v", kinds)
	}
}
