package road

import (
	"fmt"
	"sync"
	"time"

	"github.com/lixianmin/gonsole/road/component"
	"github.com/lixianmin/gonsole/road/intern"
	"github.com/lixianmin/got/loom"
	"github.com/lixianmin/logo"
)

/********************************************************************
created:    2020-08-28
author:     lixianmin

Copyright (C) - All Rights Reserved
*********************************************************************/

type (
	App struct {
		// 下面这组参数，在session里都会用到
		manager              *Manager
		wheelSecond          *loom.Wheel
		onHandShakenHandlers []func(session Session)

		accept   Acceptor
		sessions sync.Map
		wc       loom.WaitClose

		services map[string]*component.Service // all registered service
	}
)

func NewApp(accept Acceptor, opts ...AppOption) *App {
	// 默认值
	var options = appOptions{
		SerdeBuilders:     map[string]serdeBuilder{},
		HeartbeatInterval: 3 * time.Second,
		KickInterval:      time.Minute,
	}

	// 初始化
	for _, opt := range opts {
		opt(&options)
	}

	var app = &App{
		manager:     newManager(options.HeartbeatInterval, options.KickInterval),
		wheelSecond: loom.NewWheel(time.Second, int(options.HeartbeatInterval/time.Second)+1),
		accept:      accept,
		services:    make(map[string]*component.Service),
	}

	// 除默认支持JsonSerde外, 可额外添加ProtoSerde等支持
	for name, factory := range options.SerdeBuilders {
		app.manager.AddSerdeBuilder(name, factory)
	}

	loom.Go(app.goLoop)
	return app
}

func (my *App) goLoop(later loom.Later) {
	var closeChan = my.wc.C()
	// 心跳包只有一个类型, 没有具体的数据字段, 因此与serde无关
	var heartbeatBuffer = my.manager.heartbeatBuffer

	for {
		select {
		case conn := <-my.accept.GetLinkChan():
			// 外网环境是非常恶劣的, 有大量扫描器
			// 先写个心跳包检测一下链接的可用性, 如果失败了就无需建立session了
			if _, err1 := conn.Write(heartbeatBuffer); err1 == nil {
				my.onNewSession(conn)
			}
		case <-closeChan:
			return
		}
	}
}

func (my *App) onNewSession(conn intern.Link) {
	var session = my.manager.NewSession(conn)
	var err = session.Handshake()
	if err != nil {
		return
	}

	var id = session.Id()
	my.sessions.Store(id, session)

	session.OnClosed(func() {
		my.sessions.Delete(id)
	})

	// 这个在session的go loop中回调, 因此onHandShakenHandlers放在
	session.OnHandShaken(func() {
		// 不能直接使用for循环, 小心closure的问题
		var handlers = my.onHandShakenHandlers
		for i := range handlers {
			var handler = handlers[i]
			handler(session)
		}
	})
}

// OnHandShaken 暴露一个OnConnected()事件暂时没有看到很大的意义，因为handshake必须是第一个消息
// 如果需要接入握手事件的话, 可以自己注册OnHandShaken事件
func (my *App) OnHandShaken(handler func(session Session)) {
	if handler != nil {
		my.onHandShakenHandlers = append(my.onHandShakenHandlers, handler)
	}
}

func (my *App) AddInterceptor(interceptor InterceptorFunc) {
	my.manager.AddInterceptor(interceptor)
}

// RangeSessions 设计这个方法的目的是为了排查如下bug: 2个相同的uid登录了server, playerManager中好像有player丢失了
func (my *App) RangeSessions(handler func(session Session)) {
	if handler == nil {
		return
	}

	my.sessions.Range(func(key, value any) bool {
		var session = value.(Session)
		handler(session)
		return true
	})
}

func (my *App) Register(comp component.Component, opts ...component.Option) error {
	var service = component.NewService(comp, opts)

	if _, ok := my.services[service.Name]; ok {
		return fmt.Errorf("handler: service already defined: %s", service.Name)
	}

	if err := service.ExtractHandler(); err != nil {
		return err
	}

	// register all handlers
	my.services[service.Name] = service
	for name, handler := range service.Handlers {
		var route = fmt.Sprintf("%s.%s", service.Name, name)
		handler.Route = route

		my.manager.AddHandler(route, handler)
		logo.Debug("route=%s", route)
	}

	my.manager.RebuildHandlerKinds()
	return nil
}

// Documentation returns handler and remotes documentation
func (my *App) Documentation(getPtrNames bool) (map[string]any, error) {
	handlerDocs, err := intern.HandlersDocs("game", my.services, getPtrNames)
	if err != nil {
		return nil, err
	}

	return map[string]any{"handlers": handlerDocs}, nil
}

// Listen 初始化完成后才启动Listen, 才可以接受session连接进来
func (my *App) Listen() {
	my.accept.Listen()
}
