package gonsole

import (
	"github.com/lixianmin/gonsole/tools"
	"github.com/lixianmin/logo"
	"time"
)

/********************************************************************
created:    2020-06-03
author:     lixianmin

Copyright (C) - All Rights Reserved
*********************************************************************/

type ServerArgs struct {
	HandshakeTimeout time.Duration
	ReadBufferSize   int
	WriteBufferSize  int

	Port          int               // 服务端口
	UrlRoot       string            // 项目目录，表现在url中
	TemplatePath  string            // console.html模板文件的路径名
	LogRoot       string            // 日志文件根目录
	UserPasswords map[string]string // 可以登陆的用户名与密码
	Logger        logo.ILogger      // 自定义日志对象，默认只输出到控制台
}

func (args *ServerArgs) checkArgs() {
	if args.HandshakeTimeout <= 0 {
		args.HandshakeTimeout = time.Second
	}

	if args.ReadBufferSize <= 0 {
		args.ReadBufferSize = 2048
	}

	if args.WriteBufferSize <= 0 {
		args.WriteBufferSize = 2048
	}

	if args.TemplatePath == "" {
		args.TemplatePath = "vendor/github.com/lixianmin/gonsole/console.html"
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
