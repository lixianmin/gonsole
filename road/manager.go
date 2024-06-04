package road

import (
	"github.com/lixianmin/gonsole/road/component"
	"github.com/lixianmin/gonsole/road/serde"
	"github.com/lixianmin/got/iox"
	"github.com/lixianmin/got/osx"
	"maps"
	"reflect"
	"slices"
	"sort"
	"time"
)

/********************************************************************
created:    2023-06-05
author:     lixianmin

Copyright (C) - All Rights Reserved
*********************************************************************/

type InterceptorFunc func(session Session, method reflect.Method) error

type Manager struct {
	heartbeatInterval time.Duration
	kickInterval      time.Duration
	routeHandlers     map[string]*component.Handler
	routeKinds        map[string]int32
	kindHandlers      map[int32]*component.Handler
	maxKind           int32
	routes            []string
	serdes            []serde.Serde
	interceptors      []InterceptorFunc
	gid               string // client断线重连时, 基于此判断client重连的是不是上一次的同一个server进程

	heartbeatBuffer []byte
	kickBuffer      []byte
}

func NewManager(heartbeatInterval time.Duration, kickInterval time.Duration) *Manager {
	var my = &Manager{
		heartbeatInterval: heartbeatInterval,
		kickInterval:      kickInterval,
		routeHandlers:     map[string]*component.Handler{},
		routeKinds:        map[string]int32{}, // 这些默认不能为nil, 否则一旦有客户端不调用RebuildHandlerKinds(), 那么这些将一直为nil, 并影响后续的操作
		kindHandlers:      map[int32]*component.Handler{},
		maxKind:           0,
		routes:            make([]string, 0),
		serdes:            []serde.Serde{&serde.JsonSerde{}}, // 默认支持json序列化
		gid:               osx.GetGPID(0),

		heartbeatBuffer: createCommonPackBuffer(serde.Packet{Kind: serde.Heartbeat}),
		kickBuffer:      createCommonPackBuffer(serde.Packet{Kind: serde.Kick}),
	}

	return my
}

func (my *Manager) NewSession(link Link) Session {
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

func (my *Manager) AddSerde(serde serde.Serde) {
	if serde != nil {
		my.serdes = append(my.serdes, serde)
	}
}

func (my *Manager) GetSerde(name string) serde.Serde {
	for _, s := range my.serdes {
		if s.GetName() == name {
			return s
		}
	}

	return nil
}

func (my *Manager) AddInterceptor(interceptor InterceptorFunc) {
	if interceptor != nil {
		my.interceptors = append(my.interceptors, interceptor)
	}
}

func (my *Manager) GetInterceptors() []InterceptorFunc {
	return my.interceptors
}

func createCommonPackBuffer(pack serde.Packet) []byte {
	var stream = &iox.OctetsStream{}
	var writer = iox.NewOctetsWriter(stream)
	serde.EncodePacket(writer, pack)

	var buffer = stream.Bytes()
	var result = slices.Clone(buffer)

	return result
}
