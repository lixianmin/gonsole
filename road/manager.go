package road

import (
	"github.com/lixianmin/gonsole/road/component"
	"github.com/lixianmin/gonsole/road/serde"
	"sort"
	"time"
)

/********************************************************************
created:    2023-06-05
author:     lixianmin

Copyright (C) - All Rights Reserved
*********************************************************************/

type Manager struct {
	heartbeatInterval time.Duration
	routeHandlers     map[string]*component.Handler
	routeKinds        map[string]int32
	kindHandlers      map[int32]*component.Handler
	routes            []string
	serde             serde.Serde
}

func NewManager(heartbeatInterval time.Duration) *Manager {
	var my = &Manager{
		heartbeatInterval: heartbeatInterval,
		routeHandlers:     map[string]*component.Handler{},
		serde:             &serde.JsonSerde{},
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
	}
}

func (my *Manager) GetKindByRoute(route string) (int32, bool) {
	var kind, ok = my.routeKinds[route]
	return kind, ok
}

func (my *Manager) GetHandlerByKind(kind int32) *component.Handler {
	var handler = my.kindHandlers[kind]
	return handler
}

func (my *Manager) GetSerde() serde.Serde {
	return my.serde
}
