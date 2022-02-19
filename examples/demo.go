package main

import (
	"fmt"
	"github.com/lixianmin/gonsole"
	"github.com/lixianmin/got/loom"
	"github.com/lixianmin/got/timex"
	"github.com/lixianmin/logo"
	"log"
	"net/http"
	"sync"
	"time"
)

/********************************************************************
created:    2020-06-06
author:     lixianmin

Copyright (C) - All Rights Reserved
*********************************************************************/

func main() {
	var logger = logo.GetLogger().(*logo.Logger)
	logger.SetFilterLevel(logo.LevelDebug)

	var webPort = 8888
	var mux = http.NewServeMux()
	var server = gonsole.NewServer(mux,
		gonsole.WithPort(webPort),
		gonsole.WithPageTemplate("web/dist/console.html"),
		gonsole.WithPageBody("<H1>This is a very huge body</H1>"),
		gonsole.WithUserPasswords(map[string]string{"xmli": "123456"}),
		gonsole.WithEnablePProf(true),
		gonsole.WithUrlRoot("ws"),
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

	server.RegisterCommand(&gonsole.Command{
		Name:     "test_slice_sort_table_by_head",
		Note:     "测试slice排序",
		IsPublic: true,
		Handler: func(client *gonsole.Client, args []string) (*gonsole.Response, error) {
			type Bean struct {
				Text     string
				Name     string
				Age      int
				Time     string
				Hit      float32
				EmptyNum string
			}

			var now = time.Now()
			var layout = "2006-01-02"
			var beans = []Bean{{"hello", "world", 20, now.Format(layout), 1.1, ""},
				{"what", "is", 10, now.Add(365 * timex.Day).Format(layout), 2.2, ""},
				{"how", "are", 100, now.Add(-timex.Day).Format(layout), 4.3, "10"},
				{"oh", "my", 30, now.Add(timex.Day).Format(layout), 0.4, "11"},
				{"oh", "my", 30, now.Add(timex.Day).Format(layout), 0.4, "2"},
			}

			var html = gonsole.ToHtmlTable(beans)
			return gonsole.NewHtmlResponse(html), nil
		},
	})

	server.RegisterCommand(&gonsole.Command{
		Name:     "test_struct_sort_table_by_head",
		Note:     "测试结构体排序",
		IsPublic: true,
		Handler: func(client *gonsole.Client, args []string) (*gonsole.Response, error) {
			type Bean struct {
				Text string
				Name string
				Age  int
				Time string
				Hit  float32
			}

			var now = time.Now()
			var layout = "2006-01-02"
			var bean = Bean{"hello", "world", 20, now.Format(layout), 1.1}

			var html = gonsole.ToHtmlTable(bean)
			return gonsole.NewHtmlResponse(html), nil
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

	loom.Go(goLoop)
	log.Fatal(srv.ListenAndServe())
}

func goLoop(later loom.Later) {
	var lock = &sync.Mutex{}
	lock.Lock()
	go func() {
		lock.Lock()
	}()

	var timer = later.NewTimer(5 * time.Minute)

	for {
		select {
		case <-timer.C:
			fmt.Printf("timer triggered, unlock \n")
			lock.Unlock()
		}
	}
}
