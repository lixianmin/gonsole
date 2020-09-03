package gonsole

import (
	"github.com/lixianmin/road"
	"github.com/lixianmin/road/component"
	"github.com/lixianmin/gonsole/beans"
	"github.com/lixianmin/gonsole/ifs"
	"github.com/lixianmin/gonsole/logger"
	"github.com/lixianmin/gonsole/tools"
	"net/http"
	"net/http/pprof"
	"sync"
)

/********************************************************************
created:    2019-11-16
author:     lixianmin

Copyright (C) - All Rights Reserved
*********************************************************************/

type Server struct {
	args ServerArgs
	app  *road.App

	gpid     string
	commands sync.Map
	topics   sync.Map
}

func NewServer(mux IServeMux, args ServerArgs) *Server {
	args.checkArgs()
	logger.Init(args.Logger)

	var acceptor = newServerAcceptor(args.ReadBufferSize, args.WriteBufferSize)
	acceptor.HandleWebsocket(mux, args.UrlRoot+"/"+websocketName)

	var app = road.NewApp(road.AppArgs{
		Acceptor:        acceptor,
		DataCompression: false,
		Logger:          args.Logger,
	})

	var server = &Server{
		args: args,
		app:  app,
		gpid: tools.GetGPID(args.Port),
	}

	server.RegisterService("console", newConsoleService(server))
	server.registerHandlers(mux)
	server.registerBuiltinCommands()
	server.registerBuiltinTopics()

	if args.EnablePProf {
		server.enablePProf(mux)
	}

	app.OnSessionConnected(func(session *road.Session) {
		var client = newClient(server, session)
		session.Attachment().Put(ifs.KeyClient, client)

		var remoteAddress = session.RemoteAddr().String()
		_ = session.Push("console.challenge", beans.NewChallenge(server.gpid, remoteAddress))
		logger.Info("client connected, remoteAddress=%q.", remoteAddress)
	})

	logger.Info("Golang Console Server started~")
	return server
}

func (server *Server) RegisterService(name string, service component.Component) {
	server.app.Register(newConsoleService(server), component.WithName(name), component.WithNameFunc(ToSnakeName))
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
