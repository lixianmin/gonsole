package gonsole

import (
	"fmt"
	"github.com/lixianmin/gonsole/beans"
	"github.com/lixianmin/gonsole/logger"
	"html/template"
	"io/ioutil"
	"net/http"
	"path/filepath"
	"time"
)

/********************************************************************
created:    2020-08-01
author:     lixianmin

Copyright (C) - All Rights Reserved
*********************************************************************/

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
		Handler: func(client *Client, args []string) {
			var commands = server.getCommands()
			var topics = server.getTopics()
			client.SendBean(beans.NewCommandHelp(commands, topics, client.isAuthorized))
		}})

	server.RegisterCommand(&Command{
		Name:     "auth",
		Note:     "认证后开启更多命令：auth username，然后根据提示输入password",
		IsPublic: true,
		Handler: func(client *Client, args []string) {
			client.SendBean(beans.NewCommandAuth(client, args, server.args.UserPasswords))
		}})

	server.RegisterCommand(&Command{
		Name:     "log.list",
		Note:     "日志文件列表",
		IsPublic: false,
		Handler: func(client *Client, args []string) {
			client.SendBean(beans.NewCommandLogList(server.args.LogRoot))
		},
	})

	const maxHeadNum = 1000
	var headNote = fmt.Sprintf("打印文件头：tail [-n num (<=%d)] filename", maxHeadNum)
	server.RegisterCommand(&Command{
		Name:     "head",
		Note:     headNote,
		IsPublic: false,
		Handler: func(client *Client, args []string) {
			client.SendHtml(beans.ReadFileHead(headNote, args, maxHeadNum))
		},
	})

	const maxTailNum = maxHeadNum
	var tailNote = fmt.Sprintf("打印文件尾：tail [-n num (<=%d)] filename", maxTailNum)
	server.RegisterCommand(&Command{
		Name:     "tail",
		Note:     tailNote,
		IsPublic: false,
		Handler: func(client *Client, args []string) {
			client.SendHtml(beans.ReadFileTail(tailNote, args, maxTailNum))
		},
	})

	server.RegisterCommand(&Command{
		Name:     "history",
		Note:     "历史命令列表",
		IsPublic: true,
		Handler: func(client *Client, args []string) {
			client.SendBean(beans.NewBasicResponse("history", ""))
		},
	})

	server.RegisterCommand(&Command{
		Name:     "top",
		Note:     "打印进程统计信息",
		IsPublic: false,
		Handler: func(client *Client, args []string) {
			client.SendBean(beans.NewTopicTop())
		},
	})

	server.RegisterCommand(&Command{
		Name:     "date",
		Note:     "打印当前日期",
		IsPublic: true,
		Handler: func(client *Client, args []string) {
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
