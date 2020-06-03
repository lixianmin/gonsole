package gonsole

import (
	"fmt"
	"github.com/gorilla/websocket"
	"github.com/lixianmin/gocore/loom"
	"github.com/lixianmin/logo"
	"html/template"
	"log"
	"net/http"
	"time"
)

/********************************************************************
created:    2019-11-16
author:     lixianmin

Copyright (C) - All Rights Reserved
*********************************************************************/

type Server struct {
	gpid        string
	upgrader    *websocket.Upgrader
	commandChan chan ICommand
}

func NewServer() *Server {
	var upgrader = &websocket.Upgrader{
		HandshakeTimeout:  1 * time.Second,
		ReadBufferSize:    2048,
		WriteBufferSize:   2048,
		EnableCompression: true,
	}

	// todo: 不应该无条件的接受CheckOrigin
	upgrader.CheckOrigin = func(r *http.Request) bool { return true }

	var commandChan = make(chan ICommand, 32)
	var server = &Server{
		upgrader:    upgrader,
		commandChan: commandChan,
	}

	return server
}

func (server *Server) Start(mux *http.ServeMux) {
	server.registerServices(mux)
	go server.goLoop()

	logo.Info("[Start()] Server started~")
}

func (server *Server) goLoop() {
	defer loom.DumpIfPanic()
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
				logo.Info("[goLoop(%q)] client connected.", remoteAddress)
			case DetachClient:
				delete(clients, cmd.Client)
			default:
				logo.Error("[goLoop()]Invalid cmd=%v", cmd)
			}
		}
	}
}

func (server *Server) registerServices(mux *http.ServeMux) {
	// 项目目录，表现在url中
	const rootDirectory = ""

	// 处理debug消息
	var tmpl = template.Must(template.ParseFiles("core/gonsole/gonsole.html"))
	mux.HandleFunc(rootDirectory+"/gonsole", func(w http.ResponseWriter, r *http.Request) {
		var data struct {
			RootDirectory string
		}

		data.RootDirectory = rootDirectory
		_ = tmpl.Execute(w, data)
	})

	// 处理ws消息
	mux.HandleFunc(rootDirectory+"/ws", func(w http.ResponseWriter, r *http.Request) {
		conn, err := server.upgrader.Upgrade(w, r, nil)
		if err != nil {
			logo.Error("[RegisterServices(%s)]connection upgrade failed, userAgent=%q, err=%q", r.RemoteAddr, r.UserAgent(), err)
			return
		}

		// caution: client负责conn的生命周期
		var client = newClient(server, conn)
		server.SendCommand(AttachClient{Client: client})
	})
}

func (server *Server) listenAndServe(mux *http.ServeMux, port int) {
	var readTimeout = 5 * time.Second
	var writeTimeout = 2 * time.Second

	// 要绑定网卡，而不是判断某个ip地址
	var address = fmt.Sprintf(":%d", port)
	server.gpid = ""

	logo.Info("[listenAndServe(%s)] try to listen at address=%q", "", address)
	var srv = &http.Server{
		Addr:           address,
		Handler:        mux,
		ReadTimeout:    readTimeout,
		WriteTimeout:   writeTimeout,
		MaxHeaderBytes: 1 << 20,
	}

	log.Fatal(srv.ListenAndServe())
}

func (server *Server) SendCommand(cmd ICommand) {
	server.commandChan <- cmd
}

func (server *Server) GetGPID() string {
	return server.gpid
}
