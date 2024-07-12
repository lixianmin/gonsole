package road

import (
	"github.com/lixianmin/gonsole/road/component"
	"github.com/lixianmin/gonsole/road/intern"
	"github.com/lixianmin/gonsole/road/serde"
	"github.com/lixianmin/got/iox"
	"github.com/lixianmin/got/osx"
	"maps"
	"slices"
	"sort"
	"time"
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
	maxKind           int32
	routes            []string
	serdeBuilders     map[string]serdeBuilder
	interceptors      []InterceptorFunc
	gid               string // client断线重连时, 基于此判断client重连的是不是上一次的同一个server进程

	heartbeatBuffer []byte
	kickBuffer      []byte
}

func newManager(heartbeatInterval time.Duration, kickInterval time.Duration) *Manager {
	var my = &Manager{
		heartbeatInterval: heartbeatInterval,
		kickInterval:      kickInterval,
		routeHandlers:     map[string]*component.Handler{},
		routeKinds:        map[string]int32{}, // 这些默认不能为nil, 否则一旦有客户端不调用RebuildHandlerKinds(), 那么这些将一直为nil, 并影响后续的操作
		kindHandlers:      map[int32]*component.Handler{},
		maxKind:           0,
		routes:            make([]string, 0),
		serdeBuilders:     map[string]serdeBuilder{},
		gid:               osx.GetGPID(0),

		heartbeatBuffer: createCommonPackBuffer(serde.Packet{Kind: serde.Heartbeat}),
		kickBuffer:      createCommonPackBuffer(serde.Packet{Kind: serde.Kick}),
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

	var routes = make([]string, 0, size)
	for route := range my.routeHandlers {
		routes = append(routes, route)
	}

	sort.Strings(routes)
	my.routeKinds = make(map[string]int32, size)
	my.kindHandlers = make(map[int32]*component.Handler, size)
	my.routes = routes

	for i, route := range routes {
		var kind = int32(i) + serde.UserBase
		my.routeKinds[route] = kind
		my.kindHandlers[kind] = my.routeHandlers[route]
		my.maxKind = kind
	}
}

//func (my *Manager) GetKindByRoute(route string) (int32, bool) {
//	var kind, ok = my.routeKinds[route]
//	return kind, ok
//}

func (my *Manager) CloneRouteKinds() (map[string]int32, int32) {
	return maps.Clone(my.routeKinds), my.maxKind
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
