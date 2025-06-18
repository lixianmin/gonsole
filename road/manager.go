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

type Manager struct {
	heartbeatInterval time.Duration
	kickInterval      time.Duration
	routeHandlers     map[string]*component.Handler
	routeKinds        map[string]int32
	kindHandlers      map[int32]*component.Handler
	routes            string
	serdeBuilders     map[string]serdeBuilder
	interceptors      []InterceptorFunc
	gid               string // client断线重连时, 基于此判断client重连的是不是上一次的同一个server进程

	heartbeatBuffer []byte
}

func newManager(heartbeatInterval time.Duration, kickInterval time.Duration) *Manager {
	var my = &Manager{
		heartbeatInterval: heartbeatInterval,
		kickInterval:      kickInterval,
		routeHandlers:     map[string]*component.Handler{},
		routeKinds:        map[string]int32{}, // 这些默认不能为nil, 否则一旦有客户端不调用RebuildHandlerKinds(), 那么这些将一直为nil, 并影响后续的操作
		kindHandlers:      map[int32]*component.Handler{},
		serdeBuilders:     map[string]serdeBuilder{},
		gid:               osx.GetGPID(0),

		heartbeatBuffer: createCommonPackBuffer(serde.Packet{Kind: serde.Heartbeat}),
	}

	return my
}

func (my *Manager) NewSession(link intern.Link) Session {
	return newSession(my, link)
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

	// 构建routes, 并排序
	var routes = make([]string, 0, size)
	for route := range my.routeHandlers {
		routes = append(routes, route)
	}

	sort.Strings(routes)
	var joined = convert.Bytes(strings.Join(routes, " "))
	var compressed, _ = compressWithDeflate(joined, flate.BestCompression)
	var based = base64.StdEncoding.EncodeToString(convert.Bytes(compressed))
	my.routes = based

	my.routeKinds = make(map[string]int32, size)
	my.kindHandlers = make(map[int32]*component.Handler, size)

	for _, route := range routes {
		var kind = int32(len(my.routeKinds)) + serde.UserBase
		my.routeKinds[route] = kind
		my.kindHandlers[kind] = my.routeHandlers[route]
	}
}

func compressWithDeflate(data []byte, level int) (string, error) {
	// 1. 创建一个 bytes.Buffer 用于接收压缩后的数据
	var b bytes.Buffer

	// 2. 创建一个新的 flate.Writer
	//    它会将压缩过的数据写入我们提供的 buffer 中
	w, err := flate.NewWriter(&b, level)
	if err != nil {
		return "", fmt.Errorf("创建 flate writer 失败: %w", err)
	}

	// 3. 将原始数据写入 writer
	//    writer 会在内部进行压缩
	_, err = w.Write(data)
	if err != nil {
		return "", fmt.Errorf("写入数据到 writer 失败: %w", err)
	}

	// 4. (关键步骤!) 关闭 writer
	//    这一步至关重要，它会将所有缓冲中的数据刷新到底层的 io.Writer (即我们的 buffer)
	//    如果不调用 Close()，压缩结果可能不完整或为空
	if err := w.Close(); err != nil {
		return "", fmt.Errorf("关闭 writer 失败: %w", err)
	}

	// 5. 返回 buffer 中的字符串数据
	return b.String(), nil
}

func (my *Manager) CloneRouteKinds() map[string]int32 {
	return maps.Clone(my.routeKinds)
}

func (my *Manager) GetHandlerByKind(kind int32) *component.Handler {
	var handler = my.kindHandlers[kind]
	return handler
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
