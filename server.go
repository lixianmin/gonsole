package gonsole

import (
	"fmt"
	"github.com/gorilla/websocket"
	"github.com/lixianmin/gocore/loom"
	"github.com/lixianmin/gonsole/logger"
	"github.com/lixianmin/gonsole/tools"
	"html/template"
	"io/ioutil"
	"net/http"
	"sync"
	"time"
)

/********************************************************************
created:    2019-11-16
author:     lixianmin

Copyright (C) - All Rights Reserved
*********************************************************************/

type Server struct {
	args        ServerArgs
	gpid        string
	upgrader    *websocket.Upgrader
	messageChan chan IMessage

	commands sync.Map
	topics   sync.Map
}

func NewServer(mux *http.ServeMux, args ServerArgs) *Server {
	args.checkArgs()

	var upgrader = &websocket.Upgrader{
		HandshakeTimeout:  args.HandshakeTimeout,
		ReadBufferSize:    args.ReadBufferSize,
		WriteBufferSize:   args.WriteBufferSize,
		EnableCompression: true,
	}

	// todo: 不应该无条件的接受CheckOrigin
	upgrader.CheckOrigin = func(r *http.Request) bool { return true }

	var messageChan = make(chan IMessage, 32)
	var server = &Server{
		args:        args,
		gpid:        tools.GetGPID(args.Port),
		upgrader:    upgrader,
		messageChan: messageChan,
	}

	server.registerHandlers(mux)
	server.registerBuiltinCommands()
	server.registerBuiltinTopics()
	go server.goLoop()

	logger.Info("[Start()] Golang Console Server started~")
	return server
}

func (server *Server) goLoop() {
	defer loom.DumpIfPanic()
	var messageChan <-chan IMessage = server.messageChan

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
				client.SendBean(newChallenge(server.gpid, remoteAddress))
				logger.Info("[goLoop(%q)] client connected.", remoteAddress)
			case DetachClient:
				delete(clients, msg.Client)
			default:
				logger.Error("[goLoop()] Invalid msg=%v", msg)
			}
		}
	}
}

func (server *Server) registerHandlers(mux *http.ServeMux) {
	server.handleConsolePage(mux)
	server.handleLogFiles(mux)
	server.handleWebsocket(mux)
}

func (server *Server) handleConsolePage(mux *http.ServeMux) {
	var tmpl = template.Must(template.ParseFiles(server.args.TemplatePath))
	var pattern = server.args.UrlRoot + "/console"

	mux.HandleFunc(pattern, func(writer http.ResponseWriter, request *http.Request) {
		var data struct {
			UrlRoot       string
			WebsocketName string
		}

		data.UrlRoot = server.args.UrlRoot
		data.WebsocketName = websocketName
		_ = tmpl.Execute(writer, data)
	})
}

func (server *Server) handleLogFiles(mux *http.ServeMux) {
	var pattern = "/" + server.args.LogRoot + "/"
	mux.HandleFunc(pattern, func(writer http.ResponseWriter, request *http.Request) {
		var logFilePath = request.URL.Path
		if len(logFilePath) < 1 {
			return
		}

		logFilePath = logFilePath[1:]
		if tools.IsPathExist(logFilePath) {
			var bytes, err = ioutil.ReadFile(logFilePath)
			if err == nil {
				_, _ = writer.Write(bytes)
			} else {
				var text = fmt.Sprintf("%v", err)
				_, _ = writer.Write([]byte(text))
			}
		}
	})
}

func (server *Server) handleWebsocket(mux *http.ServeMux) {
	// 处理ws消息
	var pattern = server.args.UrlRoot + "/" + websocketName
	mux.HandleFunc(pattern, func(writer http.ResponseWriter, request *http.Request) {
		conn, err := server.upgrader.Upgrade(writer, request, nil)
		if err != nil {
			logger.Error("[handleWebsocket(%s)] connection upgrade failed, userAgent=%q, err=%q", request.RemoteAddr, request.UserAgent(), err)
			return
		}

		// caution: client负责conn的生命周期
		var client = newClient(server, conn)
		server.sendMessage(AttachClient{Client: client})
	})
}

func (server *Server) registerBuiltinCommands() {
	server.RegisterCommand(&Command{
		Name: "help",
		Note: "帮助中心",
		Handler: func(client *Client) {
			var commands = server.getCommands()
			var topics = server.getTopics()
			client.SendBean(newCommandHelp(commands, topics))
		}})

	server.RegisterCommand(&Command{
		Name: "logs",
		Note: "打印日志文件列表",
		Handler: func(client *Client) {
			client.SendBean(newCommandListLogFiles(server.args.LogRoot))
		},
	})
}

func (server *Server) registerBuiltinTopics() {
	server.RegisterTopic(&Topic{
		Name:     "top",
		Note:     "进程统计信息",
		Interval: 5 * time.Second,
		PrepareData: func() interface{} {
			return newTopicTop()
		}})
}

func (server *Server) RegisterCommand(cmd *Command) {
	if cmd != nil && cmd.Name != "" {
		server.commands.Store(cmd.Name, cmd)
	}
}

func (server *Server) RegisterTopic(topic *Topic) {
	if topic != nil && topic.Name != "" && topic.Interval > 0 && topic.PrepareData != nil {
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

func (server *Server) getCommands() []*Command {
	var list []*Command
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

func (server *Server) getTopics() []*Topic {
	var list []*Topic
	server.topics.Range(func(key, value interface{}) bool {
		var topic, ok = value.(*Topic)
		if ok {
			list = append(list, topic)
		}

		return true
	})

	return list
}

func (server *Server) sendMessage(msg IMessage) {
	server.messageChan <- msg
}

func (server *Server) GetGPID() string {
	return server.gpid
}
