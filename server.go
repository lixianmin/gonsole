package gonsole

import (
	"fmt"
	"github.com/gorilla/websocket"
	"github.com/lixianmin/gonsole/logger"
	"github.com/lixianmin/gonsole/tools"
	"github.com/lixianmin/got/loom"
	"html/template"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
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

	logger.Info("Golang Console Server started~")
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

// https://delveshal.github.io/2018/05/17/golang-%E5%AE%9E%E7%8E%B0%E6%96%87%E4%BB%B6%E6%96%AD%E7%82%B9%E7%BB%AD%E4%BC%A0-demo/
func RequestFileByRange(fullPath string, writer http.ResponseWriter, request *http.Request) {
	var start, end int64
	_, _ = fmt.Sscanf(request.Header.Get("Range"), "bytes=%d-%d", &start, &end)
	file, err := os.Open(fullPath)
	if err != nil {
		logger.Debug(err)
		http.NotFound(writer, request)
		return
	}

	info, err := file.Stat()
	if err != nil {
		logger.Debug(err)
		http.NotFound(writer, request)
		return
	}

	if start < 0 || start >= info.Size() || end < 0 || end >= info.Size() {
		writer.WriteHeader(http.StatusBadRequest)
		_, _ = writer.Write([]byte(fmt.Sprintf("out of index, length:%d", info.Size())))
		return
	}

	if end == 0 {
		end = info.Size() - 1
	}

	var header = writer.Header()
	header.Add("Accept-ranges", "bytes")
	header.Add("Content-Length", strconv.FormatInt(end-start+1, 10))
	header.Add("Content-Range", "bytes "+strconv.FormatInt(start, 10)+"-"+strconv.FormatInt(end, 10)+"/"+strconv.FormatInt(info.Size()-start, 10))
	header.Add("Content-Disposition", "attachment; filename="+info.Name())

	_, err = file.Seek(start, 0)
	if err != nil {
		logger.Debug(err)
		writer.WriteHeader(http.StatusInternalServerError)
		return
	}

	_, err = io.CopyN(writer, file, end-start+1)
	if err != nil {
		logger.Debug(err)
	}
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
			client.SendBean(newCommandHelp(commands, topics, client.isAuthorized))
		}})

	server.RegisterCommand(&Command{
		Name:     "auth",
		Note:     "认证后开启更多命令：auth username password",
		IsPublic: true,
		Handler: func(client *Client, texts []string) {
			client.SendBean(newCommandAuth(client, texts, server.args.UserPasswords))
		}})

	server.RegisterCommand(&Command{
		Name:     "log.list",
		Note:     "日志文件列表",
		IsPublic: false,
		Handler: func(client *Client, texts []string) {
			client.SendBean(newCommandListLogFiles(server.args.LogRoot))
		},
	})

	server.RegisterCommand(&Command{
		Name:     "history",
		Note:     "历史命令列表",
		IsPublic: true,
		Handler: func(client *Client, texts []string) {
			client.SendBean(newBasicResponse("listHistoryCommands", ""))
		},
	})

	server.RegisterCommand(&Command{
		Name:     "top",
		Note:     "返回进程统计信息",
		IsPublic: false,
		Handler: func(client *Client, texts []string) {
			client.SendBean(newTopicTop())
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

	//server.RegisterCommand(&Command{
	//	Name:     "ping",
	//	Note:     "Ping一下服务器是不是通的",
	//	IsPublic: true,
	//	Handler: func(client *Client, texts []string) {
	//		client.SendBean(newBasicResponse("pong", ""))
	//	},
	//})
}

func (server *Server) registerBuiltinTopics() {
	const intervalSeconds = 5
	server.RegisterTopic(&Topic{
		Name:     "top",
		Note:     fmt.Sprintf("广播进程统计信息（每%ds）", intervalSeconds),
		Interval: intervalSeconds * time.Second,
		IsPublic: false,
		BuildData: func() interface{} {
			return newTopicTop()
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
