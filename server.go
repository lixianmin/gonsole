package gonsole

import (
	"fmt"
	"github.com/lixianmin/gonsole/ifs"
	"github.com/lixianmin/gonsole/tools"
	"github.com/lixianmin/got/timex"
	"github.com/lixianmin/logo"
	"github.com/lixianmin/road"
	"github.com/lixianmin/road/component"
	"github.com/lixianmin/road/epoll"
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
	options serverOptions
	app     *road.App

	gpid         string
	consoleUrl   string
	commands     sync.Map
	topics       sync.Map
	lastAuthTime atomic.Value
}

func NewServer(mux IServeMux, opts ...ServerOption) *Server {
	// 默认值
	var options = serverOptions{
		ReadBufferSize:  4096,
		WriteBufferSize: 4096,

		PageTemplate: "vendor/github.com/lixianmin/gonsole/console.html",
		PageTitle:    "Console",
		PageBody:     "Input 'help' and press 'Enter' to fetch builtin commands. <a href=\"https://github.com/lixianmin/gonsole\">learn more</a>",

		AutoLoginTime:   timex.Day,
		EnablePProf:     false,
		LogListRoot:     "logs",
		Port:            8888,
		UrlRoot:         "",
		UserPasswords:   make(map[string]string),
		DeadlockIgnores: nil,
		WebSocketPath:   "",
	}

	// 初始化
	for _, opt := range opts {
		opt(&options)
	}

	// 这个是为了consoleUrl格式化的时候使用
	if options.UrlRoot != "" && options.UrlRoot[0] != '/' {
		options.UrlRoot = "/" + options.UrlRoot
	}

	var servePath = options.UrlRoot + "/" + options.WebSocketPath
	var acceptor = epoll.NewWsAcceptor(mux, servePath)
	var app = road.NewApp(acceptor,
		road.WithSessionRateLimitBySecond(2),
		road.WithSenderCount(2), // 当前游戏里使用tcp链接，这个用不到，默认16个太多了
	)

	var server = &Server{
		options:    options,
		app:        app,
		gpid:       tools.GetGPID(options.Port),
		consoleUrl: fmt.Sprintf("http://%s:%d%s/console", tools.GetLocalIP(), options.Port, options.UrlRoot),
	}

	server.lastAuthTime.Store(time.Now().Add(-timex.Day * 365))
	server.RegisterService("console", newConsoleService(server))
	server.registerHandlers(mux, options)
	server.registerBuiltinCommands(options.Port)
	server.registerBuiltinTopics()

	if options.EnablePProf {
		server.enablePProf(mux)
	}

	app.OnHandShaken(func(session *road.Session) {
		var client = newClient(session)
		session.Attachment().Put(ifs.KeyClient, client)

		var remoteAddress = session.RemoteAddr().String()
		// console.challenge协议不能随便发，因为默认情况下pitaya client不认识这个协议，会导致pitaya.connect失败
		//_ = session.Push("console.challenge", beans.NewChallenge(server.gpid, remoteAddress))
		logo.Info("client connected, remoteAddress=%q.", remoteAddress)
	})

	logo.Info("Gonsole: GoVersion     = %s", runtime.Version())
	logo.Info("Gonsole: GitBranchName = %s", GitBranchName)
	logo.Info("Gonsole: GitCommitId   = %s", GitCommitId)
	logo.Info("Gonsole: AppBuildTime  = %s", AppBuildTime)
	logo.Info("Gonsole: console       = %s", server.consoleUrl)
	logo.Info("Starting server")
	return server
}

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
	server.commands.Range(func(key, value interface{}) bool {
		var cmd, ok = value.(*Command)
		if ok {
			list = append(list, cmd)
		}

		return true
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

func (server *Server) ConsoleUrl() string {
	return server.consoleUrl
}

func (server *Server) App() *road.App {
	return server.app
}

func (server *Server) enablePProf(mux IServeMux) {
	var handler = func(processor func(w http.ResponseWriter, r *http.Request)) func(w http.ResponseWriter, r *http.Request) {
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
	mux.HandleFunc(root+"/debug/pprof/", handler(pprof.Index))
	mux.HandleFunc(root+"/debug/pprof/cmdline", handler(pprof.Cmdline))
	mux.HandleFunc(root+"/debug/pprof/profile", handler(pprof.Profile))
	mux.HandleFunc(root+"/debug/pprof/symbol", handler(pprof.Symbol))
	mux.HandleFunc(root+"/debug/pprof/trace", handler(pprof.Trace))
	mux.HandleFunc(root+"/debug/pprof/block", handler(blockHandler))
	mux.HandleFunc(root+"/debug/pprof/goroutine", handler(pprof.Handler("goroutine").ServeHTTP))
	mux.HandleFunc(root+"/debug/pprof/heap", handler(pprof.Handler("heap").ServeHTTP))
	mux.HandleFunc(root+"/debug/pprof/mutex", handler(mutexHandler))
	mux.HandleFunc(root+"/debug/pprof/threadcreate", handler(pprof.Handler("threadcreate").ServeHTTP))
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
