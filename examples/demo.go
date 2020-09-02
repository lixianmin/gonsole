package main

import (
	"fmt"
	"github.com/lixianmin/gonsole"
	"github.com/lixianmin/gonsole/ifs"
	"log"
	"net/http"
	"time"
)

/********************************************************************
created:    2020-06-06
author:     lixianmin

Copyright (C) - All Rights Reserved
*********************************************************************/

func main() {
	var webPort = 8888
	var mux = http.NewServeMux()
	var server = gonsole.NewServer(mux, gonsole.ServerArgs{
		AutoLoginLimit: 24 * time.Hour,                      // 一天内不需要重新认证
		Port:           webPort,                             // webserver端口
		TemplatePath:   "console.html",                      // 页面文件模板
		UserPasswords:  map[string]string{"xmli": "123456"}, // 认证使用的用户名密码
		EnablePProf:    true,
	})

	server.RegisterCommand(&gonsole.Command{
		Name:     "hi",
		Note:     "打印 hi console",
		IsPublic: false,
		Handler: func(client *gonsole.Client, args [] string) (*gonsole.CommandRe, error) {
			var bean struct {
				Text string
			}

			bean.Text = "hello world"
			client.Push(ifs.RouteDefault, bean)
			return nil, nil
		},
	})

	server.RegisterTopic(&gonsole.Topic{
		Name:     "hi",
		Note:     "广播hi console（每5s）",
		Interval: 5 * time.Second,
		IsPublic: true,
		BuildData: func() interface{} {
			return "hi console";
		},
	})

	var srv = &http.Server{
		Addr:           fmt.Sprintf(":%d", webPort),
		Handler:        mux,
		ReadTimeout:    2 * time.Second,
		WriteTimeout:   2 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}

	log.Fatal(srv.ListenAndServe())
}
