package gonsole

import (
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
	"sync"
)

/********************************************************************
created:    2019-11-16
author:     lixianmin

Copyright (C) - All Rights Reserved
*********************************************************************/

type Server struct {
	options serverOptions
	app     *road.App

	gpid     string
	commands sync.Map
	topics   sync.Map
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

	var servePath = options.UrlRoot + "/" + options.WebSocketPath
	var acceptor = epoll.NewWsAcceptor(mux, servePath)
	var app = road.NewApp(acceptor,
		road.WithSessionRateLimitBySecond(2),
		road.WithSenderCount(2), // 当前游戏里使用tcp链接，这个用不到，默认16个太多了
	)

	var server = &Server{
		options: options,
		app:     app,
		gpid:    tools.GetGPID(options.Port),
	}

	server.RegisterService("console", newConsoleService(server))
	server.registerHandlers(mux, options.WebSocketPath)
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
	logo.Info("Starting Gonsole Server")
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

func (server *Server) App() *road.App {
	return server.app
}

func (server *Server) enablePProf(mux IServeMux) {
	var handler = func(name string) func(w http.ResponseWriter, r *http.Request) {
		return pprof.Handler(name).ServeHTTP
	}

	const root = ""
	mux.HandleFunc(root+"/debug/pprof/", pprof.Index)
	mux.HandleFunc(root+"/debug/pprof/cmdline", pprof.Cmdline)
	mux.HandleFunc(root+"/debug/pprof/profile", pprof.Profile)
	mux.HandleFunc(root+"/debug/pprof/symbol", pprof.Symbol)
	mux.HandleFunc(root+"/debug/pprof/trace", pprof.Trace)
	mux.HandleFunc(root+"/debug/pprof/block", handler("block"))
	mux.HandleFunc(root+"/debug/pprof/goroutine", handler("goroutine"))
	mux.HandleFunc(root+"/debug/pprof/heap", handler("heap"))
	mux.HandleFunc(root+"/debug/pprof/mutex", handler("mutex"))
	mux.HandleFunc(root+"/debug/pprof/threadcreate", handler("threadcreate"))
}
