package gonsole

import (
	"github.com/gorilla/websocket"
	"html/template"
	"net/http"
	"sync"
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
	handlers    sync.Map
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
		gpid:        "", // todo 计算gpid
		upgrader:    upgrader,
		messageChan: messageChan,
	}

	server.registerServices(mux)
	go server.goLoop()

	logger.Info("[Start()] Golang Console Server started~")
	return server
}

func (server *Server) goLoop() {
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

func (server *Server) registerServices(mux *http.ServeMux) {
	// 项目目录，表现在url中
	const rootDirectory = ""
	const websocketName = "ws_console"

	// 处理console页面
	var tmpl = template.Must(template.ParseFiles(server.args.HtmlFilePath))
	mux.HandleFunc(rootDirectory+"/console", func(w http.ResponseWriter, r *http.Request) {
		var data struct {
			RootDirectory string
			WebsocketName string
		}

		data.RootDirectory = rootDirectory
		data.WebsocketName = websocketName
		_ = tmpl.Execute(w, data)
	})

	// 处理ws消息
	mux.HandleFunc(rootDirectory+"/"+websocketName, func(writer http.ResponseWriter, request *http.Request) {
		conn, err := server.upgrader.Upgrade(writer, request, nil)
		if err != nil {
			logger.Error("[registerServices(%s)]connection upgrade failed, userAgent=%q, err=%q", request.RemoteAddr, request.UserAgent(), err)
			return
		}

		// caution: client负责conn的生命周期
		var client = newClient(server, conn)
		server.SendMessage(AttachClient{Client: client})
	})
}

func (server *Server) registerDebugHandlers() {
	server.AddCommand("help", "帮助中心", func(client *Client) {
		var remoteAddress = client.GetRemoteAddress()
		client.SendBean(newDebugHelp(remoteAddress))
	})

	server.AddCommand("ls", "打印主题列表", func(client *Client) {
		client.SendBean(newDebugListTopics())
	})
}

func (server *Server) AddCommand(cmd string, remark string, handler func(client *Client)) {
	server.handlers.Store(cmd, handler)
}

func (server *Server) SendMessage(msg IMessage) {
	server.messageChan <- msg
}

func (server *Server) GetGPID() string {
	return server.gpid
}
