package gonsole

import (
	"time"
)

/********************************************************************
created:    2021-01-28
author:     lixianmin

Copyright (C) - All Rights Reserved
*********************************************************************/

type serverOptions struct {
	PageTemplate string // 主页（console.html）模板文件的路径名
	PageTitle    string // 主页（console.html）标题
	PageBody     string // 主页（console.html）主体

	AutoLoginTime   time.Duration     // 如果手动登录过，则在时限内自动登录
	DeadlockIgnores []string          // 死锁检查时可以忽略的调用字符串
	Directory       string            // 项目根目录，表现在url中
	EnablePProf     bool              // 激活pprof
	LogListRoot     string            // log.list命令显示的日志文件根目录
	Port            int               // 服务器端口
	Tts             bool              // 是否使用tts, 默认true
	UserPasswords   map[string]string // 可以登陆的用户名与密码
	WebSocketPath   string            // websocket监听的路径
}

func (options serverOptions) getPathByDirectory(path string) string {
	if options.Directory != "" {
		path = "/" + options.Directory + path
	}

	return path
}

type ServerOption func(*serverOptions)

// WithAutoLoginTime 如果手动登录过，则在时限内自动登录
func WithAutoLoginTime(d time.Duration) ServerOption {
	return func(options *serverOptions) {
		if d > 0 {
			options.AutoLoginTime = d
		}
	}
}

// WithEnablePProf 激活pprof
func WithEnablePProf(enable bool) ServerOption {
	return func(options *serverOptions) {
		options.EnablePProf = enable
	}
}

// WithLogListRoot log.list命令显示的日志文件根目录
func WithLogListRoot(path string) ServerOption {
	return func(options *serverOptions) {
		options.LogListRoot = path
	}
}

// WithPort 服务器端口
func WithPort(port int) ServerOption {
	return func(options *serverOptions) {
		if port > 0 {
			options.Port = port
		}
	}
}

// WithPageTemplate 主页（console.html）模板文件的路径名
func WithPageTemplate(path string) ServerOption {
	return func(options *serverOptions) {
		options.PageTemplate = path
	}
}

// WithPageTitle 主页（console.html）标题
func WithPageTitle(title string) ServerOption {
	return func(options *serverOptions) {
		options.PageTitle = title
	}
}

// WithPageBody 主页（console.html）主体
func WithPageBody(body string) ServerOption {
	return func(options *serverOptions) {
		options.PageBody = body
	}
}

// WithDirectory 项目根目录，表现在url中
func WithDirectory(path string) ServerOption {
	return func(options *serverOptions) {
		options.Directory = path
	}
}

// WithUserPasswords 可以登陆的用户名与密码
func WithUserPasswords(passwords map[string]string) ServerOption {
	return func(options *serverOptions) {
		if len(passwords) > 0 {
			options.UserPasswords = passwords
			//const key = "hey pet!"
			//for k, v := range passwords {
			//	var digest = tools.HmacSha256(key, v)
			//	options.UserPasswords[k] = digest
			//}
		}
	}
}

// WithDeadlockIgnores 死锁检查时可以忽略的调用字符串
func WithDeadlockIgnores(ignores []string) ServerOption {
	return func(options *serverOptions) {
		options.DeadlockIgnores = ignores
	}
}

// WithWebSocketPath websocket监听的路径
func WithWebSocketPath(path string) ServerOption {
	return func(options *serverOptions) {
		options.WebSocketPath = path
	}
}

func WithTts(enable bool) ServerOption {
	return func(options *serverOptions) {
		options.Tts = enable
	}
}
