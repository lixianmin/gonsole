package road

import (
	"fmt"
	"github.com/lixianmin/gonsole/road/component"
	"github.com/lixianmin/gonsole/road/epoll"
	"github.com/lixianmin/gonsole/road/internal"
	"github.com/lixianmin/got/loom"
	"github.com/lixianmin/got/taskx"
	"github.com/lixianmin/logo"
	"time"
)

/********************************************************************
created:    2020-08-28
author:     lixianmin

Copyright (C) - All Rights Reserved
*********************************************************************/

type (
	App struct {
		// 下面这组参数，有session里都会用到
		manager           *NetManager
		wheelSecond       *loom.Wheel
		heartbeatInterval time.Duration
		rateLimitBySecond int

		accept   epoll.Acceptor
		sessions loom.Map
		tasks    *taskx.Queue
		wc       loom.WaitClose

		services map[string]*component.Service // all registered service
	}

	appFetus struct {
		onHandShakenHandlers []func(session Session)
	}
)

func NewApp(accept epoll.Acceptor, opts ...AppOption) *App {
	// 默认值
	var options = appOptions{
		SessionRateLimitBySecond: 2,
	}

	// 初始化
	for _, opt := range opts {
		opt(&options)
	}

	var heartbeatInterval = accept.GetHeartbeatInterval()
	var app = &App{
		manager:           NewNetManager(),
		wheelSecond:       loom.NewWheel(time.Second, int(heartbeatInterval/time.Second)+1),
		heartbeatInterval: heartbeatInterval,
		rateLimitBySecond: options.SessionRateLimitBySecond,

		accept:   accept,
		services: make(map[string]*component.Service),
	}

	// 这个tasks，只是内部用一下，不公开
	app.tasks = taskx.NewQueue(taskx.WithSize(2), taskx.WithCloseChan(app.wc.C()))

	loom.Go(app.goLoop)
	return app
}

func (my *App) goLoop(later loom.Later) {
	var fetus = &appFetus{}

	var closeChan = my.wc.C()
	for {
		select {
		case conn := <-my.accept.GetConnChan():
			my.onNewSession(fetus, conn)
		case task := <-my.tasks.C:
			var err = task.Do(fetus)
			if err != nil {
				logo.JsonI("err", err)
			}
		case <-closeChan:
			return
		}
	}
}

func (my *App) onNewSession(fetus *appFetus, conn epoll.IConn) {
	var session = NewSession(my.manager, conn)
	var err = session.ShakeHand(float32(my.heartbeatInterval.Seconds()))
	if err != nil {
		return
	}

	var id = session.Id()
	my.sessions.Put(id, session)

	session.OnClosed(func() {
		my.sessions.Remove(id)
	})

	// for循环中小心closure的问题
	var handlers = fetus.onHandShakenHandlers
	for i := range handlers {
		var handler = handlers[i]
		handler(session)
	}
}

// OnHandShaken 暴露一个OnConnected()事件暂时没有看到很大的意义，因为handshake必须是第一个消息
// 如果需要接入握手事件的话, 可以自己注册OnHandShaken事件
func (my *App) OnHandShaken(handler func(session Session)) {
	if handler != nil {
		my.tasks.SendCallback(func(args interface{}) (result interface{}, err error) {
			var fetus = args.(*appFetus)
			fetus.onHandShakenHandlers = append(fetus.onHandShakenHandlers, handler)
			return nil, nil
		})
	}
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
		my.manager.AddHandler(route, handler)
		logo.Debug("route=%s", route)
	}

	return nil
}

// Documentation returns handler and remotes docum/7entacion
func (my *App) Documentation(getPtrNames bool) (map[string]interface{}, error) {
	handlerDocs, err := internal.HandlersDocs("game", my.services, getPtrNames)
	if err != nil {
		return nil, err
	}

	return map[string]interface{}{"handlers": handlerDocs}, nil
}
