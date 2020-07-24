package gonsole

import (
	"fmt"
	"github.com/gorilla/websocket"
	"github.com/lixianmin/gonsole/beans"
	"github.com/lixianmin/gonsole/ifs"
	"github.com/lixianmin/gonsole/logger"
	"github.com/lixianmin/gonsole/tools"
	"github.com/lixianmin/got/loom"
	"html/template"
	"io/ioutil"
	"net/http"
	"path/filepath"
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
	messageChan chan ifs.IMessage

	commands sync.Map
	topics   sync.Map
}

func NewServer(mux IServeMux, args ServerArgs) *Server {
	args.checkArgs()
	logger.Init(args.Logger)

	var upgrader = &websocket.Upgrader{
		HandshakeTimeout:  args.HandshakeTimeout,
		ReadBufferSize:    args.ReadBufferSize,
		WriteBufferSize:   args.WriteBufferSize,
		EnableCompression: true,
	}

	// todo: 不应该无条件的接受CheckOrigin
	upgrader.CheckOrigin = func(r *http.Request) bool { return true }

	var messageChan = make(chan ifs.IMessage, 32)
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

func (server *Server) registerHandlers(mux IServeMux) {
	server.handleConsolePage(mux)
	server.handleSha256Js(mux)
	server.handleLogFiles(mux)
	server.handleWebsocket(mux)
}

func (server *Server) handleConsolePage(mux IServeMux) {
	var args = server.args
	var tmpl = template.Must(template.ParseFiles(args.TemplatePath))
	var pattern = args.UrlRoot + "/console"

	mux.HandleFunc(pattern, func(writer http.ResponseWriter, request *http.Request) {
		var data struct {
			AutoLoginLimit int64
			Title          string
			UrlRoot        string
			WebsocketName  string
		}

		data.AutoLoginLimit = int64(args.AutoLoginLimit / time.Millisecond)
		data.Title = args.Title
		data.UrlRoot = args.UrlRoot
		data.WebsocketName = websocketName
		_ = tmpl.Execute(writer, data)
	})
}

func (server *Server) handleSha256Js(mux IServeMux) {
	var pattern = server.args.UrlRoot + "/sha256.min.js"
	mux.HandleFunc(pattern, func(writer http.ResponseWriter, request *http.Request) {
		var path = request.URL.Path
		if len(path) < 1 {
			return
		}

		var root = filepath.Dir(server.args.TemplatePath)
		var filename = filepath.Join(root, path)
		var bytes, err = ioutil.ReadFile(filename)
		if err == nil {
			_, _ = writer.Write(bytes)
		} else {
			var text = fmt.Sprintf("err=%q", err)
			_, _ = writer.Write([]byte(text))
		}
	})
}

// 这个方法在gin中由于pattern不一样，需要被重写
func (server *Server) handleLogFiles(mux IServeMux) {
	var pattern = "/" + server.args.LogRoot + "/"
	mux.HandleFunc(pattern, func(writer http.ResponseWriter, request *http.Request) {
		var logFilePath = request.URL.Path
		if len(logFilePath) < 1 {
			return
		}

		logFilePath = logFilePath[1:]
		RequestFileByRange(logFilePath, writer, request)
	})
}

func (server *Server) handleWebsocket(mux IServeMux) {
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
		Name:     "help",
		Note:     "帮助中心",
		IsPublic: true,
		Handler: func(client *Client, texts []string) {
			var commands = server.getCommands()
			var topics = server.getTopics()
			client.SendBean(beans.NewCommandHelp(commands, topics, client.isAuthorized))
		}})

	server.RegisterCommand(&Command{
		Name:     "auth",
		Note:     "认证后开启更多命令：auth username，然后根据提示输入password",
		IsPublic: true,
		Handler: func(client *Client, texts []string) {
			client.SendBean(beans.NewCommandAuth(client, texts, server.args.UserPasswords))
		}})

	server.RegisterCommand(&Command{
		Name:     "log.list",
		Note:     "日志文件列表",
		IsPublic: false,
		Handler: func(client *Client, texts []string) {
			client.SendBean(beans.NewCommandLogList(server.args.LogRoot))
		},
	})

	const maxTailNum = 1000
	var tailNote = fmt.Sprintf("打印文件尾：tail [-n num (<=%d)] filename", maxTailNum)
	server.RegisterCommand(&Command{
		Name:     "tail",
		Note:     tailNote,
		IsPublic: false,
		Handler: func(client *Client, texts []string) {
			client.SendHtml(beans.ReadTail(tailNote, texts, maxTailNum))
		},
	})

	server.RegisterCommand(&Command{
		Name:     "history",
		Note:     "历史命令列表",
		IsPublic: true,
		Handler: func(client *Client, texts []string) {
			client.SendBean(beans.NewBasicResponse("history", ""))
		},
	})

	server.RegisterCommand(&Command{
		Name:     "top",
		Note:     "打印进程统计信息",
		IsPublic: false,
		Handler: func(client *Client, texts []string) {
			client.SendBean(beans.NewTopicTop())
		},
	})

	server.RegisterCommand(&Command{
		Name:     "date",
		Note:     "打印当前日期",
		IsPublic: true,
		Handler: func(client *Client, texts []string) {
			const layout = "Mon 2006-01-02 15:04:05"
			var text = time.Now().Format(layout)
			client.SendBean(text)
		},
	})
}

func (server *Server) registerBuiltinTopics() {
	const intervalSeconds = 5
	server.RegisterTopic(&Topic{
		Name:     "top",
		Note:     fmt.Sprintf("广播进程统计信息（每%ds）", intervalSeconds),
		Interval: intervalSeconds * time.Second,
		IsPublic: false,
		BuildData: func() interface{} {
			return beans.NewTopicTop()
		}})
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

func (server *Server) getTopics() []ifs.Topic {
	var list []ifs.Topic
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
