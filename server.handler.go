package gonsole

import (
	"fmt"
	"github.com/lixianmin/gonsole/beans"
	"github.com/lixianmin/gonsole/tools"
	"github.com/lixianmin/got/convert"
	"html/template"
	"io/fs"
	"net/http"
	"path/filepath"
	"runtime"
	"strings"
	"time"
)

/********************************************************************
created:    2020-08-01
author:     lixianmin

Copyright (C) - All Rights Reserved
*********************************************************************/

func (server *Server) registerHandlers(mux IServeMux, options serverOptions) {
	server.handleConsolePage(mux, options.WebSocketPath)
	server.handleWebConfig(mux, options.WebSocketPath)
	//server.handleResources(mux, "res/js")
	server.handleAssets(mux)
	server.handleLogFiles(mux)
}

func (server *Server) handleConsolePage(mux IServeMux, websocketPath string) {
	var options = server.options
	var tmpl = template.Must(template.ParseFiles(options.PageTemplate))
	var pattern = "/" + options.UrlRoot + "/console"

	// 刷新的时候，console间隔性的pending刷新不出来，这个有可能是http.ServeMux的问题，使用gin之后无此bug
	mux.HandleFunc(pattern, func(writer http.ResponseWriter, request *http.Request) {
		var data struct {
			AutoLoginLimit int64
			Title          string
			Body           string
			UrlRoot        string
			WebsocketPath  string
		}

		data.AutoLoginLimit = int64(options.AutoLoginTime / time.Millisecond)
		data.Title = options.PageTitle
		data.Body = options.PageBody
		data.UrlRoot = options.UrlRoot
		data.WebsocketPath = websocketPath
		_ = tmpl.Execute(writer, data)
	})
}

func (server *Server) handleWebConfig(mux IServeMux, websocketPath string) {
	var options = server.options
	var pattern = "/" + options.UrlRoot + "/web_config"

	mux.HandleFunc(pattern, func(writer http.ResponseWriter, request *http.Request) {
		var header = writer.Header()
		header.Set("Content-Type", "application/json")
		header.Set("Access-Control-Allow-Origin", "*")
		header.Set("Access-Control-Allow-Credentials", "true")

		var data struct {
			AutoLoginLimit int64  `json:"autoLoginLimit"`
			Title          string `json:"title"`
			Body           string `json:"body"`
			UrlRoot        string `json:"urlRoot"`
			WebsocketPath  string `json:"websocketPath"`
		}

		data.AutoLoginLimit = int64(options.AutoLoginTime / time.Millisecond)
		data.Title = options.PageTitle
		data.Body = options.PageBody
		data.UrlRoot = options.UrlRoot
		data.WebsocketPath = websocketPath

		var json = convert.ToJson(data)
		_, _ = writer.Write(json)
	})
}

func (server *Server) handleAssets(mux IServeMux) {
	var isValidResource = func(name string) bool {
		var extensions = []string{".css", ".html", ".ico", ".js", ".png"}
		name = strings.ToLower(name)

		for _, item := range extensions {
			if strings.HasSuffix(name, item) {
				return true
			}
		}

		return false
	}

	var getContentType = func(filename string) string {
		var index = strings.LastIndex(filename, ".")
		if index > 0 {
			var extension = filename[index:]
			switch extension {
			case ".css":
				return "text/css"
			case ".js":
				return "text/javascript"
			}
		}

		return "text/plain"
	}

	var pageRoot = filepath.Dir(server.options.PageTemplate)
	var walkRoot = filepath.Join(pageRoot, "assets")

	if err := filepath.Walk(walkRoot, func(relativePath string, info fs.FileInfo, err error) error {
		if err == nil && !info.IsDir() && isValidResource(info.Name()) {
			var pattern = relativePath[len("web/dist"):]
			var contentType = getContentType(relativePath)

			mux.HandleFunc(pattern, func(writer http.ResponseWriter, request *http.Request) {
				if contentType != "text/plain" {
					writer.Header().Set("Content-Type", contentType)
				}

				RequestFileByRange(relativePath, writer, request)
			})
		}
		return err
	}); err != nil {
		panic(err)
	}
}

// 这个方法在gin中由于pattern不一样，需要被重写
func (server *Server) handleLogFiles(mux IServeMux) {
	var pattern = "/" + server.options.LogListRoot + "/"
	mux.HandleFunc(pattern, func(writer http.ResponseWriter, request *http.Request) {
		var logFilePath = request.URL.Path
		if len(logFilePath) < 1 {
			return
		}

		logFilePath = logFilePath[1:]
		RequestFileByRange(logFilePath, writer, request)
	})
}

//func (server *Server) handleHealth(mux IServeMux) {
//	var pattern = server.options.UrlRoot + "/health"
//	mux.HandleFunc(pattern, func(writer http.ResponseWriter, request *http.Request) {
//		var message = `{"status":"UP"}`
//		_, _ = writer.Write([]byte(message))
//	})
//}

func (server *Server) registerBuiltinCommands(port int) {
	server.RegisterCommand(&Command{
		Name:      "help",
		Note:      "帮助中心",
		IsPublic:  true,
		isBuiltin: true,
		Handler: func(client *Client, args []string) (*Response, error) {
			var isAuthorized = isAuthorized(client.Session())
			var commandHelp = beans.FetchCommandHelp(server.getCommands(), isAuthorized)
			var data = fmt.Sprintf("<br/><b>命令列表：</b> <br> %s", ToHtmlTable(commandHelp))

			var topicHelp = beans.FetchTopicHelp(server.getTopics(), isAuthorized)
			if len(topicHelp) > 0 {
				data += fmt.Sprintf("<br/><b>主题列表：</b> <br> %s", ToHtmlTable(topicHelp))
			}

			if isAuthorized && server.options.EnablePProf {
				data += "<br/><b>PProf：</b> <br>" + ToHtmlTable(beans.FetchPProfHelp(args))
			}

			return NewHtmlResponse(data), nil
		}})

	server.RegisterCommand(&Command{
		Name:      "auth",
		Note:      "认证后开启更多命令：auth username，然后根据提示输入password",
		IsPublic:  true,
		isBuiltin: true,
		Handler: func(client *Client, args []string) (*Response, error) {
			server.lastAuthTime.Store(time.Now())
			var data = beans.NewCommandAuth(client.Session(), args, server.options.UserPasswords, port)
			return NewDefaultResponse(data), nil
		}})

	server.RegisterCommand(&Command{
		Name:      "log.list",
		Note:      "日志文件列表",
		IsPublic:  false,
		isBuiltin: true,
		Handler: func(client *Client, args []string) (*Response, error) {
			var data = beans.NewCommandLogList(server.options.LogListRoot)
			var ret = &Response{Operation: "log.list", Data: data}
			return ret, nil
		},
	})

	const maxHeadNum = 1000
	var headNote = fmt.Sprintf("打印文件头：head [-n num (<=%d)] [-f fitler] [-s startLine] filename", maxHeadNum)
	server.RegisterCommand(&Command{
		Name:      "head",
		Note:      headNote,
		IsPublic:  false,
		isBuiltin: true,
		Handler: func(client *Client, args []string) (*Response, error) {
			var data = beans.ReadFileHead(headNote, args, maxHeadNum)
			return NewHtmlResponse(data), nil
		},
	})

	const maxTailNum = maxHeadNum
	var tailNote = fmt.Sprintf("打印文件尾：tail [-n num (<=%d)] [-f filter] filename", maxTailNum)
	server.RegisterCommand(&Command{
		Name:      "tail",
		Note:      tailNote,
		IsPublic:  false,
		isBuiltin: true,
		Handler: func(client *Client, args []string) (*Response, error) {
			var data = beans.ReadFileTail(tailNote, args, maxTailNum)
			return NewHtmlResponse(data), nil
		},
	})

	server.RegisterCommand(&Command{
		Name:      "history",
		Note:      "历史命令列表",
		IsPublic:  true,
		isBuiltin: true,
		Handler: func(client *Client, args []string) (*Response, error) {
			return &Response{Operation: "history"}, nil
		},
	})

	server.RegisterCommand(&Command{
		Name:      "top",
		Note:      "打印进程统计信息",
		IsPublic:  false,
		isBuiltin: true,
		Handler: func(client *Client, args []string) (*Response, error) {
			var html = tools.ToHtmlTable(beans.NewTopicTop())
			return NewHtmlResponse(html), nil
		},
	})

	server.RegisterCommand(&Command{
		Name:      "app.info",
		Note:      "打印app信息",
		IsPublic:  false,
		isBuiltin: true,
		Handler: func(client *Client, args []string) (*Response, error) {
			var info = beans.CommandAppInfo{
				GoVersion:        runtime.Version(),
				GitBranchName:    GitBranchName,
				GitCommitId:      GitCommitId,
				GitCommitMessage: GitCommitMessage,
				GitCommitTime:    GitCommitTime,
				AppBuildTime:     AppBuildTime,
			}

			var html = tools.ToHtmlTable(info)
			return NewHtmlResponse(html), nil
		},
	})

	server.RegisterCommand(&Command{
		Name:      "date",
		Note:      "打印当前日期",
		IsPublic:  true,
		isBuiltin: true,
		Handler: func(client *Client, args []string) (*Response, error) {
			const layout = "Mon 2006-01-02 15:04:05"
			var text = time.Now().Format(layout)
			return NewDefaultResponse(text), nil
		},
	})

	server.RegisterCommand(&Command{
		Name:      "deadlock.detect",
		Note:      "deadlock.detect [-a (show all)] ：按IO wait时间打印goroutine，辅助死锁排查",
		isBuiltin: true,
		Handler: func(client *Client, args []string) (*Response, error) {
			var html = beans.DeadlockDetect(args, server.options.DeadlockIgnores)
			if html != "" {
				return NewHtmlResponse(html), nil
			} else {
				return NewDefaultResponse("暂时没有等待时间超长的goroutine"), nil
			}
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
		BuildResponse: func() *Response {
			var html = tools.ToHtmlTable(beans.NewTopicTop())
			return NewHtmlResponse(html)
		}})
}
