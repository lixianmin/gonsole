package gonsole

import (
	"github.com/lixianmin/gonsole/tools"
	"time"
)

/********************************************************************
created:    2020-06-03
author:     lixianmin

Copyright (C) - All Rights Reserved
*********************************************************************/

type ServerArgs struct {
	ReadBufferSize  int
	WriteBufferSize int

	AutoLoginLimit  time.Duration     // 如果手动登录过，则在时限内自动登录
	EnablePProf     bool              // 激活pprof
	LogRoot         string            // 日志文件根目录
	Port            int               // 服务端口
	TemplatePath    string            // console.html模板文件的路径名
	Title           string            // 网页的title
	UrlRoot         string            // 项目目录，表现在url中
	UserPasswords   map[string]string // 可以登陆的用户名与密码
	DeadlockIgnores []string          // 死锁检查时可以忽略的调用字符串
	WebsocketPath   string            // websocket监听的路径
}

func (args *ServerArgs) checkArgs() {
	if args.ReadBufferSize <= 0 {
		args.ReadBufferSize = 4096
	}

	if args.WriteBufferSize <= 0 {
		args.WriteBufferSize = 4096
	}

	if args.TemplatePath == "" {
		args.TemplatePath = "vendor/github.com/lixianmin/gonsole/console.html"
	}

	if args.Title == "" {
		args.Title = "Console"
	}

	if args.LogRoot == "" {
		args.LogRoot = "logs"
	}

	if args.UserPasswords == nil {
		args.UserPasswords = make(map[string]string)
	} else {
		const key = "hey pet!";
		for k, v := range args.UserPasswords {
			var digest = tools.HmacSha256(key, v)
			args.UserPasswords[k] = digest
		}
	}
}
