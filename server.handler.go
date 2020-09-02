package gonsole

import (
	"fmt"
	"github.com/lixianmin/gonsole/beans"
	"github.com/lixianmin/gonsole/ifs"
	"github.com/lixianmin/gonsole/tools"
	"html/template"
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
	server.handlerResourceFile(mux, "/res/js/sha256.min.js")
	server.handlerResourceFile(mux, "/res/js/protocol.js")
	server.handlerResourceFile(mux, "/res/js/starx.js")
	server.handleLogFiles(mux)
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

func (server *Server) handlerResourceFile(mux IServeMux, relativePath string) {
	var pattern = server.args.UrlRoot + relativePath
	mux.HandleFunc(pattern, func(writer http.ResponseWriter, request *http.Request) {
		var path = request.URL.Path
		if len(path) < 1 {
			return
		}

		var root = filepath.Dir(server.args.TemplatePath)
		var filename = filepath.Join(root, path)
		RequestFileByRange(filename, writer, request)
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

func (server *Server) registerBuiltinCommands() {
	server.RegisterCommand(&Command{
		Name:      "help",
		Note:      "帮助中心",
		IsPublic:  true,
		isBuiltin: true,
		Handler: func(client *Client, args []string) (*CommandRe, error) {
			var isAuthorized = client.Attachment().Bool(ifs.KeyIsAuthorized)
			var commandHelp = beans.FetchCommandHelp(server.GetCommands(), isAuthorized)
			var result = fmt.Sprintf("<br/><b>命令列表：</b> <br> %s", ToHtmlTable(commandHelp))

			var topicHelp = beans.FetchTopicHelp(server.getTopics(), isAuthorized)
			if len(topicHelp) > 0 {
				result += fmt.Sprintf("<br/><b>主题列表：</b> <br> %s", ToHtmlTable(topicHelp))
			}

			if server.args.EnablePProf {
				result += "<br/><b>PProf：</b> <br>" + ToHtmlTable(beans.FetchPProfHelp(args))
			}

			client.PushHtml(result)
			return nil, nil
		}})

	server.RegisterCommand(&Command{
		Name:      "auth",
		Note:      "认证后开启更多命令：auth username，然后根据提示输入password",
		IsPublic:  true,
		isBuiltin: true,
		Handler: func(client *Client, args []string) (*CommandRe, error) {
			client.Push(ifs.RouteDefault, beans.NewCommandAuth(client.Session(), args, server.args.UserPasswords))
			return nil, nil
		}})

	server.RegisterCommand(&Command{
		Name:      "log.list",
		Note:      "日志文件列表",
		IsPublic:  false,
		isBuiltin: true,
		Handler: func(client *Client, args []string) (*CommandRe, error) {
			client.Push("console.log.list", beans.NewCommandLogList(server.args.LogRoot))
			return nil, nil
		},
	})

	const maxHeadNum = 1000
	var headNote = fmt.Sprintf("打印文件头：head [-n num (<=%d)] [-f fitler] [-s startLine] filename", maxHeadNum)
	server.RegisterCommand(&Command{
		Name:      "head",
		Note:      headNote,
		IsPublic:  false,
		isBuiltin: true,
		Handler: func(client *Client, args []string) (*CommandRe, error) {
			client.PushHtml(beans.ReadFileHead(headNote, args, maxHeadNum))
			return nil, nil
		},
	})

	const maxTailNum = maxHeadNum
	var tailNote = fmt.Sprintf("打印文件尾：tail [-n num (<=%d)] [-f filter] filename", maxTailNum)
	server.RegisterCommand(&Command{
		Name:      "tail",
		Note:      tailNote,
		IsPublic:  false,
		isBuiltin: true,
		Handler: func(client *Client, args []string) (*CommandRe, error) {
			client.PushHtml(beans.ReadFileTail(tailNote, args, maxTailNum))
			return nil, nil
		},
	})

	server.RegisterCommand(&Command{
		Name:      "history",
		Note:      "历史命令列表",
		IsPublic:  true,
		isBuiltin: true,
		Handler: func(client *Client, args []string) (*CommandRe, error) {
			client.Push(ifs.RouteDefault, beans.NewBasicResponse("history", ""))
			return nil, nil
		},
	})

	server.RegisterCommand(&Command{
		Name:      "top",
		Note:      "打印进程统计信息",
		IsPublic:  false,
		isBuiltin: true,
		Handler: func(client *Client, args []string) (*CommandRe, error) {
			var html = tools.ToHtmlTable(beans.NewTopicTopData())
			client.PushHtml(html)
			return nil, nil
		},
	})

	server.RegisterCommand(&Command{
		Name:      "date",
		Note:      "打印当前日期",
		IsPublic:  true,
		isBuiltin: true,
		Handler: func(client *Client, args []string) (*CommandRe, error) {
			const layout = "Mon 2006-01-02 15:04:05"
			var text = time.Now().Format(layout)
			client.Push(ifs.RouteDefault, text)
			return nil, nil
		},
	})

	server.RegisterCommand(&Command{
		Name:      "deadlock.detect",
		Note:      "deadlock.detect [-a (show all)] ：按IO wait时间打印goroutine，辅助死锁排查",
		isBuiltin: true,
		Handler: func(client *Client, args []string) (*CommandRe, error) {
			var html = beans.DeadlockDetect(args, server.args.DeadlockIgnores)
			if html != "" {
				client.PushHtml(html)
			} else {
				client.Push(ifs.RouteDefault, "暂时没有等待时间超长的goroutine")
			}
			return nil, nil
		},
	})
}

func (server *Server) registerBuiltinTopics() {
	const intervalSeconds = 5
	server.RegisterTopic(&Topic{
		Name:      "top",
		Note:      fmt.Sprintf("广播进程统计信息（每%ds）", intervalSeconds),
		Interval:  intervalSeconds * time.Second,
		IsPublic:  false,
		isBuiltin: true,
		BuildData: func() interface{} {
			return beans.NewTopicTop()
		}})
}
