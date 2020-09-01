package gonsole

import (
	"github.com/lixianmin/gonsole/beans"
	"github.com/lixianmin/gonsole/ifs"
	"github.com/lixianmin/gonsole/logger"
	"github.com/lixianmin/gonsole/network"
	"github.com/lixianmin/gonsole/tools"
	"github.com/lixianmin/got/loom"
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
	app  *network.App

	gpid        string
	messageChan chan ifs.IMessage

	commands sync.Map
	topics   sync.Map
}

func NewServer(mux IServeMux, args ServerArgs) *Server {
	args.checkArgs()
	logger.Init(args.Logger)

	var acceptor = newServerAcceptor(args.ReadBufferSize, args.WriteBufferSize)
	var app = network.NewApp(network.AppArgs{
		Acceptor:        acceptor,
		DataCompression: false,
	})

	var messageChan = make(chan ifs.IMessage, 32)
	var server = &Server{
		args:        args,
		app:         app,
		gpid:        tools.GetGPID(args.Port),
		messageChan: messageChan,
	}

	server.registerHandlers(mux)
	server.registerBuiltinCommands()
	server.registerBuiltinTopics()
	go server.goLoop()

	if args.EnablePProf {
		server.enablePProf(mux)
	}

	logger.Info("Golang Console Server started~")
	return server
}

func (server *Server) goLoop() {
	defer loom.DumpIfPanic()
	var messageChan <-chan ifs.IMessage = server.messageChan

	// 注册的client列表
	var clients = make(map[*Client]struct{}, 16)
	for {
		select {
		case msg := <-messageChan:
			switch msg := msg.(type) {
			case AttachClient:
				var client = msg.Client
				clients[client] = struct{}{}

				var remoteAddress = client.GetRemoteAddress()
				client.SendBean(beans.NewChallenge(server.gpid, remoteAddress))
				logger.Info("client connected, remoteAddress=%q.", remoteAddress)
			case DetachClient:
				delete(clients, msg.Client)
			default:
				logger.Error("Invalid msg=%v", msg)
			}
		}
	}
}

func (server *Server) RegisterCommand(cmd *Command) {
	if cmd != nil && cmd.Name != "" {
		server.commands.Store(cmd.Name, cmd)
	}
}

func (server *Server) RegisterTopic(topic *Topic) {
	if topic != nil && topic.Name != "" && topic.Interval > 0 && topic.BuildData != nil {
		server.topics.Store(topic.Name, topic)
		topic.start()
	}
}

func (server *Server) getCommand(name string) *Command {
	var box, ok = server.commands.Load(name)
	if ok {
		var cmd, _ = box.(*Command)
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

func (server *Server) sendMessage(msg ifs.IMessage) {
	server.messageChan <- msg
}

func (server *Server) GetGPID() string {
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
