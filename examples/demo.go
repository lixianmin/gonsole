package main

import (
	"fmt"
	"github.com/lixianmin/gonsole"
	"github.com/lixianmin/logo"
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
	var server = gonsole.NewServer(mux,
		gonsole.WithPort(webPort),
		gonsole.WithPageTemplate("console.html"),
		gonsole.WithPageBody("<H1>This is a very huge body</H1>"),
		gonsole.WithUserPasswords(map[string]string{"xmli": "123456"}),
		gonsole.WithEnablePProf(true),
	)

	var app = server.App()
	app.AddHook(func(rawMethod func() (interface{}, error)) (interface{}, error) {
		var start = time.Now()
		var ret, err = rawMethod()
		var delta = time.Since(start)
		logo.Info("cost time = %s", delta)
		return ret, err
	})

	app.AddHook(func(rawMethod func() (interface{}, error)) (interface{}, error) {
		var ret, err = rawMethod()
		logo.Info("hello world")
		return ret, err
	})

	server.RegisterCommand(&gonsole.Command{
		Name:     "hi",
		Note:     "打印 hi console",
		IsPublic: false,
		Handler: func(client *gonsole.Client, args []string) (*gonsole.Response, error) {
			var bean struct {
				Text string
			}

			bean.Text = "hello world"
			return gonsole.NewDefaultResponse(bean), nil
		},
	})

	server.RegisterTopic(&gonsole.Topic{
		Name:     "hi",
		Note:     "广播hi console（每5s）",
		Interval: 5 * time.Second,
		IsPublic: true,
		BuildResponse: func() *gonsole.Response {
			return gonsole.NewDefaultResponse("hi console")
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
