package gonsole

import (
	"github.com/gorilla/websocket"
	"html/template"
	"net/http"
	"time"
)

/********************************************************************
created:    2019-11-16
author:     lixianmin

Copyright (C) - All Rights Reserved
*********************************************************************/

type ServerArgs struct {
	HandshakeTimeout time.Duration
	ReadBufferSize   int
	WriteBufferSize  int
	HtmlFilePath     string
}

func (args *ServerArgs) checkArgs() {
	if args.HandshakeTimeout <= 0 {
		args.HandshakeTimeout = time.Second
	}

	if args.ReadBufferSize <= 0 {
		args.ReadBufferSize = 2048
	}

	if args.WriteBufferSize <= 0 {
		args.WriteBufferSize = 2048
	}

	if args.HtmlFilePath == "" {
		args.HtmlFilePath = "vendor/github.com/lixianmin/gonsole/console.html"
	}
}

type Server struct {
	args        ServerArgs
	gpid        string
	upgrader    *websocket.Upgrader
	commandChan chan ICommand
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

	var commandChan = make(chan ICommand, 32)
	var server = &Server{
		args:        args,
		gpid:        "",
		upgrader:    upgrader,
		commandChan: commandChan,
	}

	server.registerServices(mux)
	go server.goLoop()

	logger.Info("[Start()] Golang Console Server started~")
	return server
}

func (server *Server) goLoop() {
	var commandChan <-chan ICommand = server.commandChan

	// 注册的client列表
	var clients = make(map[*Client]struct{}, 16)
	for {
		select {
		case cmd := <-commandChan:
			switch cmd := cmd.(type) {
			case AttachClient:
				var client = cmd.Client
				clients[client] = struct{}{}

				var remoteAddress = client.GetRemoteAddress()
				client.SendBean(NewChallenge(server.gpid, remoteAddress))
				logger.Info("[goLoop(%q)] client connected.", remoteAddress)
			case DetachClient:
				delete(clients, cmd.Client)
			default:
				logger.Error("[goLoop()]Invalid cmd=%v", cmd)
			}
		}
	}
}

func (server *Server) registerServices(mux *http.ServeMux) {
	// 项目目录，表现在url中
	const rootDirectory = ""
	const websocketName = "ws_console"

	// 处理debug消息
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
		server.SendCommand(AttachClient{Client: client})
	})
}

func (server *Server) SendCommand(cmd ICommand) {
	server.commandChan <- cmd
}

func (server *Server) GetGPID() string {
	return server.gpid
}
