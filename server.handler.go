package gonsole

import (
	"fmt"
	"github.com/lixianmin/gonsole/beans"
	"github.com/lixianmin/gonsole/road"
	"github.com/lixianmin/gonsole/tools"
	"github.com/lixianmin/got/convert"
	"github.com/lixianmin/got/iox"
	"html/template"
	"io"
	"net/http"
	"os"
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
	server.handleAssets(mux)
	server.handleLogFiles(mux, options)
}

func (server *Server) handleConsolePage(mux IServeMux, websocketPath string) {
	var options = server.options
	var tmpl = template.Must(template.ParseFiles(options.PageTemplate))

	var config struct {
		Directory     string `json:"directory"`
		WebsocketPath string `json:"websocketPath"`
		Title         string `json:"title"`
		Body          string `json:"body"`
	}

	config.Directory = options.Directory
	config.WebsocketPath = websocketPath
	config.Title = options.PageTitle
	config.Body = options.PageBody

	var data struct {
		Data string
	}

	// 真正传的是一个json序列化后的Data字段
	data.Data = convert.String(convert.ToJson(config))
	// 模板数据刷到cache中, 后续只使用cache即可
	var cache = &iox.Buffer{}
	_ = tmpl.Execute(cache, data)

	// 刷新的时候，console间隔性的pending刷新不出来，这个有可能是http.ServeMux的问题，使用gin之后无此bug
	var pattern = options.getPathByDirectory("/console")

	mux.HandleFunc(pattern, func(writer http.ResponseWriter, request *http.Request) {
		_, _ = cache.Seek(0, io.SeekStart)
		_, _ = io.Copy(writer, cache)
	})
}

func (server *Server) handleAssets(mux IServeMux) {
	var isValidAsset = func(path string) bool {
		var extensions = []string{".css", ".html", ".ico", ".js"}
		for _, extension := range extensions {
			if strings.HasSuffix(path, extension) {
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
			case ".html":
				return "text/html"
			case ".ico":
				return "image/x-icon"
			case ".js":
				return "text/javascript"
			}
		}

		return "text/plain"
	}

	var pageRoot = filepath.Dir(server.options.PageTemplate)
	//var walkRoot = filepath.Join(pageRoot, "assets")
	var walkRoot = pageRoot

	// 如果是windows平台，dirName="web\\dist"
	const dirName = "web" + string(os.PathSeparator) + "dist"
	const dirLength = len(dirName)

	if err := filepath.Walk(walkRoot, func(relativePath string, info os.FileInfo, err error) error {
		if err == nil && !info.IsDir() && isValidAsset(relativePath) {
			var index = strings.Index(relativePath, dirName)
			var pattern = strings.Replace(relativePath[index+dirLength:], "\\", "/", -1) // 如果是windows平台，则需要把\替换为/
			var contentType = getContentType(relativePath)

			mux.HandleFunc(pattern, func(writer http.ResponseWriter, request *http.Request) {
				if contentType != "text/plain" {
					var header = writer.Header()
					header.Set("Content-Type", contentType)
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
func (server *Server) handleLogFiles(mux IServeMux, options serverOptions) {
	var pattern = options.getPathByDirectory("/" + server.options.LogListRoot + "/")
	var cutLength = len(pattern) - len(server.options.LogListRoot) - 1
	mux.HandleFunc(pattern, func(writer http.ResponseWriter, request *http.Request) {
		var logFilePath = request.URL.Path
		if len(logFilePath) < 1 {
			return
		}

		logFilePath = logFilePath[cutLength:]
		RequestFileByRange(logFilePath, writer, request)
	})
}

//func (server *Server) handleHealth(mux IServeMux) {
//	var pattern = server.options.Directory + "/health"
//	mux.HandleFunc(pattern, func(writer http.ResponseWriter, request *http.Request) {
//		var message = `{"status":"UP"}`
//		_, _ = writer.Write([]byte(message))
//	})
//}

func (server *Server) registerBuiltinCommands(port int) {
	server.RegisterCommand(&Command{
		Name: "help",
		Note: "帮助中心",
		Flag: flagBuiltin | FlagPublic,
		Handler: func(session road.Session, args []string) (*Response, error) {
			var isAuthorized = isAuthorized(session)
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
		Name:    "auth",
		Example: "auth username",
		Note:    "认证后开启更多命令：auth username，然后根据提示输入password",
		Flag:    flagBuiltin | FlagPublic,
		Handler: func(session road.Session, args []string) (*Response, error) {
			server.lastAuthTime.Store(time.Now())
			var options = server.options

			var data = beans.NewCommandAuth(session, args, options.SecretKey, options.UserPasswords, options.AutoLoginTime, port)
			return NewDefaultResponse(data), nil
		}})

	server.RegisterCommand(&Command{
		Name: "log.list",
		Note: "日志文件列表",
		Flag: flagBuiltin,
		Handler: func(session road.Session, args []string) (*Response, error) {
			var data = beans.NewCommandLogList(server.options.LogListRoot, server.options.SecretKey)
			var ret = &Response{Operation: "log.list", Data: data}
			return ret, nil
		},
	})

	const maxHeadNum = 1000
	var headNote = fmt.Sprintf("打印文件头：head [-n num (<=%d)] [-f fitler] [-s startLine] filename", maxHeadNum)
	server.RegisterCommand(&Command{
		Name: "head",
		Note: headNote,
		Flag: flagBuiltin,
		Handler: func(session road.Session, args []string) (*Response, error) {
			var data = beans.ReadFileHead(headNote, args, maxHeadNum)
			return NewHtmlResponse(data), nil
		},
	})

	const maxTailNum = maxHeadNum
	var tailNote = fmt.Sprintf("打印文件尾：tail [-n num (<=%d)] [-f filter] filename", maxTailNum)
	server.RegisterCommand(&Command{
		Name: "tail",
		Note: tailNote,
		Flag: flagBuiltin,
		Handler: func(session road.Session, args []string) (*Response, error) {
			var data = beans.ReadFileTail(tailNote, args, maxTailNum)
			return NewHtmlResponse(data), nil
		},
	})

	server.RegisterCommand(&Command{
		Name: "history",
		Note: "历史命令列表",
		Flag: flagBuiltin | FlagPublic,
		Handler: func(session road.Session, args []string) (*Response, error) {
			return &Response{Operation: "history"}, nil
		},
	})

	server.RegisterCommand(&Command{
		Name: "top",
		Note: "打印进程统计信息",
		Flag: flagBuiltin,
		Handler: func(session road.Session, args []string) (*Response, error) {
			var html = tools.ToHtmlTable(beans.NewTopicTop())
			return NewHtmlResponse(html), nil
		},
	})

	server.RegisterCommand(&Command{
		Name: "app.info",
		Note: "打印app信息",
		Flag: flagBuiltin,
		Handler: func(session road.Session, args []string) (*Response, error) {
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
		Name: "date",
		Note: "打印当前日期",
		Flag: flagBuiltin | FlagPublic,
		Handler: func(session road.Session, args []string) (*Response, error) {
			const layout = "Mon 2006-01-02 15:04:05"
			var text = time.Now().Format(layout)
			return NewDefaultResponse(text), nil
		},
	})

	server.RegisterCommand(&Command{
		Name: "deadlock.detect",
		Note: "deadlock.detect [-a (show all)] ：按IO wait时间打印goroutine，辅助死锁排查",
		Flag: flagBuiltin,
		Handler: func(session road.Session, args []string) (*Response, error) {
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
		Name:     "top",
		Note:     fmt.Sprintf("广播进程统计信息（每%ds）", intervalSeconds),
		Interval: intervalSeconds * time.Second,
		Flag:     flagBuiltin,
		BuildResponse: func() *Response {
			var html = tools.ToHtmlTable(beans.NewTopicTop())
			return NewHtmlResponse(html)
		}})
}
