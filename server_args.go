package gonsole

import "time"

/********************************************************************
created:    2020-06-03
author:     lixianmin

Copyright (C) - All Rights Reserved
*********************************************************************/

type ServerArgs struct {
	HandshakeTimeout time.Duration
	ReadBufferSize   int
	WriteBufferSize  int

	Port         int    // 服务端口
	UrlRoot      string // 项目目录，表现在url中
	TemplatePath string // console.html模板文件的路径名
	LogRoot      string // 日志文件根目录
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
}
