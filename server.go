package gonsole

import (
	"fmt"
	"github.com/lixianmin/gonsole/ifs"
	"github.com/lixianmin/gonsole/road"
	"github.com/lixianmin/gonsole/road/component"
	"github.com/lixianmin/gonsole/road/epoll"
	"github.com/lixianmin/gonsole/road/serde"
	"github.com/lixianmin/got/osx"
	"github.com/lixianmin/got/timex"
	"github.com/lixianmin/logo"
	"net/http"
	"net/http/pprof"
	"runtime"
	"strconv"
	"sync"
	"sync/atomic"
	"time"
)

/********************************************************************
created:    2019-11-16
author:     lixianmin

Copyright (C) - All Rights Reserved
*********************************************************************/

type Server struct {
	options      serverOptions
	app          *road.App
	gpid         string
	baseUrl      string
	commands     sync.Map
	topics       sync.Map
	lastAuthTime atomic.Value
}

func NewServer(mux IServeMux, opts ...ServerOption) *Server {
	// 默认值
	var options = serverOptions{
		PageTemplate: "vendor/github.com/lixianmin/gonsole/web/dist/console.html",
		PageTitle:    "Console",
		PageBody:     "Input 'help' and press 'Enter' to fetch builtin commands. <a href=\"https://github.com/lixianmin/gonsole\">learn more</a>",

		AutoLoginTime:   timex.Day,
		DeadlockIgnores: nil,
		Directory:       "",
		EnablePProf:     false,
		LogListRoot:     "logs",
		Port:            8888,
		SecretKey:       "Hey Pet!!",
		Tls:             true,
		UserPasswords:   make(map[string]string),
		WebSocketPath:   "",
	}

	// 初始化
	for _, opt := range opts {
		opt(&options)
	}

	// 这个是为了consoleUrl格式化的时候使用
	if options.Directory != "" && options.Directory[0] == '/' {
		options.Directory = options.Directory[1:]
	}

	var servePath = options.getPathByDirectory("/" + options.WebSocketPath)
	var acceptor = epoll.NewWsAcceptor(mux, servePath)
	var app = road.NewApp(acceptor, road.WithSerdeBuilder("json", func(session road.Session) serde.Serde {
		return &serde.JsonSerde{}
	}))

	var server = &Server{
		options: options,
		app:     app,
		gpid:    osx.GetGPID(options.Port),
	}

	server.baseUrl = buildBaseUrl(options)
	server.lastAuthTime.Store(time.Now().Add(-timex.Day * 365))
	server.RegisterService("console", newConsoleService(server))
	server.registerHandlers(mux, options)
	server.registerBuiltinCommands(options.Port)
	server.registerBuiltinTopics()

	if options.EnablePProf {
		server.enablePProf(mux)
	}

	app.OnHandShaken(func(session road.Session) {
		var remoteAddress = session.RemoteAddr().String()
		logo.Info("client connected, remoteAddress=%q.", remoteAddress)
	})

	logo.Info("gonsole: GoVersion     		= %s", runtime.Version())
	logo.Info("gonsole: GitBranchName 		= %s", GitBranchName)
	logo.Info("gonsole: GitCommitId   		= %s", GitCommitId)
	logo.Info("gonsole: GitCommitMessage		= %s", GitCommitMessage)
	logo.Info("gonsole: GitCommitTime 		= %s", GitCommitTime)
	logo.Info("gonsole: AppBuildTime  		= %s", AppBuildTime)
	logo.Info("gonsole: console       		= %s", server.baseUrl+options.getPathByDirectory("/console"))
	logo.Info("Starting server")
	return server
}

// 这里不需要开放OnHandShaken()事件, 原因是:
//	1. 目前项目中使用的app并不是gonsole中自动创建的这一个, 是项目自己单独创建的
//	2. 即使项目使用的是gonsole创建的app对象, 也可以直接通过server.App()方法拿到app对象后自己注册OnHandShaken()回调
//
// OnHandShaken 开放OnHandShaken()事件(可以反复注册多个). 原因是在用户认证流程中需要在这个时机向client发送challenge协议
//func (server *Server) OnHandShaken(handler func(session road.Session)) {
//	server.app.OnHandShaken(handler)
//}

func (server *Server) RegisterService(name string, service component.Component) {
	_ = server.app.Register(service, component.WithName(name), component.WithNameFunc(ToSnakeName))
}

func (server *Server) RegisterCommand(cmd *Command) {
	if cmd != nil && cmd.Name != "" {
		server.commands.Store(cmd.Name, cmd)
	}
}

func (server *Server) RegisterTopic(topic *Topic) {
	if topic != nil && topic.Name != "" && topic.Interval > 0 && topic.BuildResponse != nil {
		server.topics.Store(topic.Name, topic)
		topic.start()
	}
}

func (server *Server) getCommand(name string) ifs.Command {
	var box, ok = server.commands.Load(name)
	if ok {
		var cmd, _ = box.(ifs.Command)
		return cmd
	}

	return nil
}

func (server *Server) getCommands() []ifs.Command {
	var list []ifs.Command
	server.commands.Range(func(key, value any) bool {
		var cmd, ok = value.(*Command)
		if ok {
			list = append(list, cmd)
		}

		return true
	})

	list = append(list, &Command{
		Name:    "request",
		Example: `request console.command {"command":"help"}`,
		Note:    "模拟直接发送请求",
		Flag:    flagBuiltin,
	})

	return list
}

func (server *Server) getTopic(name string) *Topic {
	var box, ok = server.topics.Load(name)
	if ok {
		var client, _ = box.(*Topic)
		return client
	}

	return nil
}

func (server *Server) getTopics() []ifs.Command {
	var list []ifs.Command
	server.topics.Range(func(key, value interface{}) bool {
		var topic, ok = value.(*Topic)
		if ok {
			list = append(list, topic)
		}

		return true
	})

	return list
}

func (server *Server) GPID() string {
	return server.gpid
}

func (server *Server) BaseUrl() string {
	return server.baseUrl
}

func (server *Server) App() *road.App {
	return server.app
}

func (server *Server) enablePProf(mux IServeMux) {
	var verifyAuth = func(processor func(w http.ResponseWriter, r *http.Request)) func(w http.ResponseWriter, r *http.Request) {
		return func(w http.ResponseWriter, r *http.Request) {
			const validTime = 10 * time.Minute

			var lastAuthTime = server.lastAuthTime.Load().(time.Time)
			var pastTime = time.Now().Sub(lastAuthTime)
			if pastTime > validTime {
				// 下面返回的数据，其实识别不了，会报：unrecognized profile format
				_, _ = w.Write([]byte(fmt.Sprintf(`安全起见：使用auth指令登录后%s内可以查看pprof信息，请重新登录`, timex.FormatDuration(validTime))))
				return
			}

			processor(w, r)
		}
	}

	const root = "" // 这个不能随便改，改完了能打开页面，但看不到数据；去看看pprof.Index()的实现，里面路径写死了
	mux.HandleFunc(root+"/debug/pprof/", verifyAuth(pprof.Index))
	mux.HandleFunc(root+"/debug/pprof/cmdline", verifyAuth(pprof.Cmdline))
	mux.HandleFunc(root+"/debug/pprof/profile", verifyAuth(pprof.Profile))
	mux.HandleFunc(root+"/debug/pprof/symbol", verifyAuth(pprof.Symbol))
	mux.HandleFunc(root+"/debug/pprof/trace", verifyAuth(pprof.Trace))
	mux.HandleFunc(root+"/debug/pprof/block", verifyAuth(blockHandler))
	mux.HandleFunc(root+"/debug/pprof/goroutine", verifyAuth(pprof.Handler("goroutine").ServeHTTP))
	mux.HandleFunc(root+"/debug/pprof/heap", verifyAuth(pprof.Handler("heap").ServeHTTP))
	mux.HandleFunc(root+"/debug/pprof/mutex", verifyAuth(mutexHandler))
	mux.HandleFunc(root+"/debug/pprof/threadcreate", verifyAuth(pprof.Handler("threadcreate").ServeHTTP))
}

func blockHandler(w http.ResponseWriter, r *http.Request) {
	rate, _ := strconv.Atoi(r.FormValue("rate"))
	if rate == 0 {
		rate = 10000
	}

	// 这个需要先设值，再解绑
	runtime.SetBlockProfileRate(rate)
	defer runtime.SetBlockProfileRate(0)

	pprof.Handler("block").ServeHTTP(w, r)
}

func mutexHandler(w http.ResponseWriter, r *http.Request) {
	rate, _ := strconv.Atoi(r.FormValue("rate"))
	if rate == 0 {
		rate = 10000
	}

	// 这个需要先设值，再解绑
	runtime.SetMutexProfileFraction(rate)
	defer runtime.SetMutexProfileFraction(0)

	pprof.Handler("mutex").ServeHTTP(w, r)
}

func buildBaseUrl(options serverOptions) string {
	if options.BaseUrl != "" {
		return options.BaseUrl
	}

	var protocol = "http"
	if options.Tls {
		protocol = "https"
	}

	var baseUrl = fmt.Sprintf("%s://%s:%d", protocol, osx.GetLocalIp(), options.Port)
	return baseUrl
}
