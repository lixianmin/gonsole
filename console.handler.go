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

func (my *Console) registerHandlers(mux IServeMux, options consoleOptions) {
	my.handleConsolePage(mux, options.WebSocketPath)
	my.handleAssets(mux)
	my.handleLogFiles(mux, options)
}

func (my *Console) handleConsolePage(mux IServeMux, websocketPath string) {
	var options = my.options
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

func (my *Console) handleAssets(mux IServeMux) {
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

	var pageRoot = filepath.Dir(my.options.PageTemplate)
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
func (my *Console) handleLogFiles(mux IServeMux, options consoleOptions) {
	var pattern = options.getPathByDirectory("/" + my.options.LogListRoot + "/")
	var cutLength = len(pattern) - len(my.options.LogListRoot) - 1
	mux.HandleFunc(pattern, func(writer http.ResponseWriter, request *http.Request) {
		var logFilePath = request.URL.Path
		if len(logFilePath) < 1 {
			return
		}

		logFilePath = logFilePath[cutLength:]
		RequestFileByRange(logFilePath, writer, request)
	})
}

//func (console *Console) handleHealth(mux IServeMux) {
//	var pattern = console.options.Directory + "/health"
//	mux.HandleFunc(pattern, func(writer http.ResponseWriter, request *http.Request) {
//		var message = `{"status":"UP"}`
//		_, _ = writer.Write([]byte(message))
//	})
//}

func (my *Console) registerBuiltinCommands(port int) {
	my.RegisterCommand(&Command{
		Name: "help",
		Note: "帮助中心",
		Flag: flagBuiltin | FlagPublic,
		Handler: func(session road.Session, args []string) (*Response, error) {
			var isAuthorized = isAuthorized(session)
			var commandHelp = beans.FetchCommandHelp(my.getCommands(), isAuthorized)
			var data = fmt.Sprintf("<br/><b>命令列表：</b> <br> %s", ToHtmlTable(commandHelp))

			var topicHelp = beans.FetchTopicHelp(my.getTopics(), isAuthorized)
			if len(topicHelp) > 0 {
				data += fmt.Sprintf("<br/><b>主题列表：</b> <br> %s", ToHtmlTable(topicHelp))
			}

			if isAuthorized && my.options.EnablePProf {
				data += "<br/><b>PProf：</b> <br>" + ToHtmlTable(beans.FetchPProfHelp(args))
			}

			return NewHtmlResponse(data), nil
		}})

	my.RegisterCommand(&Command{
		Name:    "auth",
		Example: "auth username",
		Note:    "认证后开启更多命令：auth username，然后根据提示输入password",
		Flag:    flagBuiltin | FlagPublic,
		Handler: func(session road.Session, args []string) (*Response, error) {
			my.lastAuthTime.Store(time.Now())
			var options = my.options

			var data = beans.NewCommandAuth(session, args, options.SecretKey, options.UserPasswords, options.AutoLoginTime, port)
			return NewDefaultResponse(data), nil
		}})

	my.RegisterCommand(&Command{
		Name: "log.list",
		Note: "日志文件列表",
		Flag: flagBuiltin,
		Handler: func(session road.Session, args []string) (*Response, error) {
			var data = beans.NewCommandLogList(my.options.LogListRoot, my.options.SecretKey)
			var ret = &Response{Operation: "log.list", Data: data}
			return ret, nil
		},
	})

	const maxHeadNum = 1000
	var headNote = "打印文件头"
	my.RegisterCommand(&Command{
		Name:    "head",
		Example: fmt.Sprintf("head [-n num (<=%d)] [-f fitler] [-s startLine] filename", maxHeadNum),
		Note:    headNote,
		Flag:    flagBuiltin,
		Handler: func(session road.Session, args []string) (*Response, error) {
			var data = beans.ReadFileHead(headNote, args, maxHeadNum)
			return NewHtmlResponse(data), nil
		},
	})

	const maxTailNum = maxHeadNum
	var tailNote = "打印文件尾"
	my.RegisterCommand(&Command{
		Name:    "tail",
		Example: fmt.Sprintf("tail [-n num (<=%d)] [-f filter] filename", maxTailNum),
		Note:    tailNote,
		Flag:    flagBuiltin,
		Handler: func(session road.Session, args []string) (*Response, error) {
			var data = beans.ReadFileTail(tailNote, args, maxTailNum)
			return NewHtmlResponse(data), nil
		},
	})

	my.RegisterCommand(&Command{
		Name: "history",
		Note: "历史命令列表",
		Flag: flagBuiltin | FlagPublic,
		Handler: func(session road.Session, args []string) (*Response, error) {
			return &Response{Operation: "history"}, nil
		},
	})

	my.RegisterCommand(&Command{
		Name: "top",
		Note: "打印进程统计信息",
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
			html += "<br>" + tools.ToHtmlTable(beans.NewTopicTop())
			return NewHtmlResponse(html), nil
		},
	})

	//my.RegisterCommand(&Command{
	//	Name: "app.info",
	//	Note: "打印app信息",
	//	Flag: flagBuiltin,
	//	Handler: func(session road.Session, args []string) (*Response, error) {
	//		var info = beans.CommandAppInfo{
	//			GoVersion:        runtime.Version(),
	//			GitBranchName:    GitBranchName,
	//			GitCommitId:      GitCommitId,
	//			GitCommitMessage: GitCommitMessage,
	//			GitCommitTime:    GitCommitTime,
	//			AppBuildTime:     AppBuildTime,
	//		}
	//
	//		var html = tools.ToHtmlTable(info)
	//		return NewHtmlResponse(html), nil
	//	},
	//})

	my.RegisterCommand(&Command{
		Name: "date",
		Note: "打印当前日期",
		Flag: flagBuiltin | FlagPublic,
		Handler: func(session road.Session, args []string) (*Response, error) {
			const layout = "Mon 2006-01-02 15:04:05"
			var text = time.Now().Format(layout)
			return NewDefaultResponse(text), nil
		},
	})

	my.RegisterCommand(&Command{
		Name:    "deadlock.detect",
		Example: "deadlock.detect [-a (show all)]",
		Note:    "按IO wait时间打印goroutine，辅助死锁排查",
		Flag:    flagBuiltin,
		Handler: func(session road.Session, args []string) (*Response, error) {
			var html = beans.DeadlockDetect(args, my.options.DeadlockIgnores)
			if html != "" {
				return NewHtmlResponse(html), nil
			} else {
				return NewDefaultResponse("暂时没有等待时间超长的goroutine"), nil
			}
		},
	})
}

func (my *Console) registerBuiltinTopics() {
	const intervalSeconds = 5
	my.RegisterTopic(&Topic{
		Name:     "top",
		Note:     fmt.Sprintf("广播进程统计信息（每%ds）", intervalSeconds),
		Interval: intervalSeconds * time.Second,
		Flag:     flagBuiltin,
		BuildResponse: func() *Response {
			var html = tools.ToHtmlTable(beans.NewTopicTop())
			return NewHtmlResponse(html)
		}})
}
