package road

import (
	"bytes"
	"compress/flate"
	"encoding/base64"
	"fmt"
	"maps"
	"slices"
	"sort"
	"strings"
	"sync/atomic"
	"time"

	"github.com/lixianmin/gonsole/road/component"
	"github.com/lixianmin/gonsole/road/intern"
	"github.com/lixianmin/gonsole/road/serde"
	"github.com/lixianmin/got/convert"
	"github.com/lixianmin/got/iox"
	"github.com/lixianmin/got/osx"
)

/********************************************************************
created:    2023-06-05
author:     lixianmin

Copyright (C) - All Rights Reserved
*********************************************************************/

type InterceptorFunc func(session Session, route string) error

// handlerKinds 预分配route kind的不可变快照（routeKinds + kindHandlers + routes）。
// RebuildHandlerKinds()整体构建后一次性原子发布，并发读取（CloneRouteKinds/GetHandlerByKind/Handshake）
// 永远看到完整（旧或新）的快照：既不存在"半填充"窗口，也没有未同步的字段读写
// （Go race detector 对未同步的 map 字段读写会直接报 race）。
type handlerKinds struct {
	routeKinds   map[string]int32
	kindHandlers map[int32]*component.Handler
	routes       string
}

type Manager struct {
	heartbeatInterval time.Duration
	kickInterval      time.Duration
	routeHandlers     map[string]*component.Handler
	kinds             atomic.Pointer[handlerKinds] // 预分配kind快照，见handlerKinds注释
	serdeBuilders     map[string]serdeBuilder
	interceptors      []InterceptorFunc
	gid               string // client断线重连时, 基于此判断client重连的是不是上一次的同一个server进程

	heartbeatBuffer []byte
	idGenerator     atomic.Int64 // Session ID生成器，替代全局变量
	dynamicKindSeq  atomic.Int64 // 动态route kind分配序列，始终高于所有预分配kind（预分配只增不减，动态kind必须避开）
}

func newManager(heartbeatInterval time.Duration, kickInterval time.Duration) *Manager {
	var my = &Manager{
		heartbeatInterval: heartbeatInterval,
		kickInterval:      kickInterval,
		routeHandlers:     map[string]*component.Handler{},
		serdeBuilders:     map[string]serdeBuilder{},
		gid:               osx.GetGPID(0),

		heartbeatBuffer: createCommonPackBuffer(serde.Packet{Kind: serde.Heartbeat}),
	}

	// routeKinds等默认不能为nil, 否则一旦有客户端不调用RebuildHandlerKinds(), 那么这些将一直为nil, 并影响后续的操作
	var empty = &handlerKinds{
		routeKinds:   map[string]int32{},
		kindHandlers: map[int32]*component.Handler{},
	}
	my.kinds.Store(empty)

	// 动态kind必须从UserBase之上开始（UserBase之下是协议包kind，如Handshake/RouteKind）
	my.dynamicKindSeq.Store(serde.UserBase)

	return my
}

func (my *Manager) NewSession(link intern.Link) Session {
	var id = my.idGenerator.Add(1)
	return newSession(my, link, id)
}

func (my *Manager) AddHandler(route string, handler *component.Handler) {
	if handler != nil {
		my.routeHandlers[route] = handler
	}
}

func (my *Manager) RebuildHandlerKinds() {
	var size = len(my.routeHandlers)
	if size == 0 {
		return
	}

	// 预分配kind覆盖[UserBase, UserBase+size-1]，动态kind必须从UserBase+size开始。
	// 必须在发布新routes/map之前抬升dynamicKindSeq：这样并发的 nextRouteKind()（比如
	// 持有旧clone的会话动态注册）要么拿到抬升后的值（>新预分配上限），要么CAS失败重试，
	// 绝不会撞上新预分配的kind。若本次Rebuild后续步骤失败，序列只会更高，不会破坏单调性。
	var floor = int64(serde.UserBase) + int64(size) - 1
	for {
		var cur = my.dynamicKindSeq.Load()
		if cur >= floor {
			break
		}
		if my.dynamicKindSeq.CompareAndSwap(cur, floor) {
			break
		}
	}

	// 构建routes, 排序, 并压缩
	var routes = make([]string, 0, size)
	for route := range my.routeHandlers {
		routes = append(routes, route)
	}

	sort.Strings(routes)
	var joined = convert.Bytes(strings.Join(routes, " "))
	var compressed, _ = compressWithDeflate(joined, flate.BestCompression)
	var routesString = base64.StdEncoding.EncodeToString(compressed)

	// bugfix(2026-08-09): 原实现先 `my.routeKinds = make(...)` 赋值空 map 再循环填充，
	// 填充期间并发的 NewSession()->CloneRouteKinds() 会克隆到"半填充"的 map（kind 只分配了
	// 一部分，但 len 已增长），该 session 的动态注册 sendRouteKind(len+UserBase) 就会撞上
	// 预分配中尚未填充的 kind（实测 .chat_stream 与 refresh_notify 同一连接都拿到 109）。
	// 修复：先在局部 map 填充完整，最后一次 Store() 原子发布，并发 clone 永远
	// 拿到完整（旧或新）的快照，不存在半填充窗口。
	var routeKinds = make(map[string]int32, size)
	var kindHandlers = make(map[int32]*component.Handler, size)

	for _, route := range routes {
		var kind = int32(len(routeKinds)) + serde.UserBase
		routeKinds[route] = kind
		kindHandlers[kind] = my.routeHandlers[route]
	}

	var snapshot = &handlerKinds{
		routeKinds:   routeKinds,
		kindHandlers: kindHandlers,
		routes:       routesString,
	}
	my.kinds.Store(snapshot)
}

// nextRouteKind 分配一个动态route kind（CAS单调递增，多session并发安全）。
// 返回值始终大于所有已发布预分配的kind：
//   - 预分配kind = UserBase + 排序索引，只增不减（route只增不删）
//   - RebuildHandlerKinds()在发布新routes前会把序列抬升到UserBase+len-1
//
// 为什么不能像旧实现那样用 len(routeKinds)+UserBase 或者仅按session内max+1分配？
// 会话的routeKinds是NewSession时克隆的快照，热更注册新route后旧会话的快照是陈旧的：
// 旧实现按快照len/max分配会撞上新预分配的kind（实测.chat_stream与refresh_notify同拿109），
// 客户端把两个route映射到同一个kind，推送数据被错误的handler解码。
func (my *Manager) nextRouteKind() int32 {
	for {
		var cur = my.dynamicKindSeq.Load()
		var next = cur + 1
		if my.dynamicKindSeq.CompareAndSwap(cur, next) {
			return int32(next)
		}
	}
}

func compressWithDeflate(data []byte, level int) ([]byte, error) {
	// 1. 创建一个 bytes.Buffer 用于接收压缩后的数据
	var b bytes.Buffer

	// 2. 创建一个新的 flate.Writer
	//    它会将压缩过的数据写入我们提供的 buffer 中
	w, err := flate.NewWriter(&b, level)
	if err != nil {
		return nil, fmt.Errorf("创建 flate writer 失败: %w", err)
	}

	// 3. 将原始数据写入 writer
	//    writer 会在内部进行压缩
	_, err = w.Write(data)
	if err != nil {
		return nil, fmt.Errorf("写入数据到 writer 失败: %w", err)
	}

	// 4. (关键步骤!) 关闭 writer
	//    这一步至关重要，它会将所有缓冲中的数据刷新到底层的 io.Writer (即我们的 buffer)
	//    如果不调用 Close()，压缩结果可能不完整或为空
	if err := w.Close(); err != nil {
		return nil, fmt.Errorf("关闭 writer 失败: %w", err)
	}

	// 5. 返回 buffer 中的二进制数据
	return b.Bytes(), nil
}

func (my *Manager) CloneRouteKinds() map[string]int32 {
	return maps.Clone(my.kinds.Load().routeKinds)
}

func (my *Manager) GetHandlerByKind(kind int32) *component.Handler {
	var snapshot = my.kinds.Load()
	var handler = snapshot.kindHandlers[kind]
	return handler
}

func (my *Manager) getRoutes() string {
	return my.kinds.Load().routes
}

func (my *Manager) AddSerdeBuilder(name string, builder serdeBuilder) {
	if name != "" && builder != nil {
		my.serdeBuilders[name] = builder
	}
}

func (my *Manager) CreateSerde(name string, session Session) serde.Serde {
	var builder = my.serdeBuilders[name]
	if builder != nil {
		return builder(session)
	}

	return nil
}

func (my *Manager) AddInterceptor(interceptor InterceptorFunc) {
	if interceptor != nil {
		my.interceptors = append(my.interceptors, interceptor)
	}
}

func createCommonPackBuffer(pack serde.Packet) []byte {
	var stream = &iox.OctetsStream{}
	var writer = iox.NewOctetsWriter(stream)
	serde.EncodePacket(writer, pack)

	var buffer = stream.Bytes()
	var result = slices.Clone(buffer)

	return result
}
