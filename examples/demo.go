package main

import (
	"fmt"
	"github.com/lixianmin/gonsole"
	"github.com/lixianmin/gonsole/road"
	"github.com/lixianmin/got/loom"
	"github.com/lixianmin/got/timex"
	"github.com/lixianmin/logo"
	"log"
	"net/http"
	"sync"
	"sync/atomic"
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
		gonsole.WithDirectory("ws"),
	)

	//var app = server.App()
	//app.AddHook(func(rawMethod func() (interface{}, error)) (interface{}, error) {
	//	var start = time.Now()
	//	var ret, err = rawMethod()
	//	var delta = time.Since(start)
	//	logo.Info("cost time = %s", delta)
	//	return ret, err
	//})
	//
	//app.AddHook(func(rawMethod func() (interface{}, error)) (interface{}, error) {
	//	var ret, err = rawMethod()
	//	logo.Info("hello world")
	//	return ret, err
	//})

	registerCommands(server)

	var srv = &http.Server{
		Addr:           fmt.Sprintf(":%d", webPort),
		Handler:        mux,
		ReadTimeout:    2 * time.Second,
		WriteTimeout:   2 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}

	loom.Go(goLoop)

	// 使用mkcert生成自签名证书，以启用并测试http/2和https，支持frame传输
	// mkcert -cert-file localhost.crt -key-file localhost.key localhost 127.0.0.1 `ipconfig getifaddr en1`

	var certFile = "res/ssl/localhost.crt"
	var keyFile = "res/ssl/localhost.key"
	log.Fatal(srv.ListenAndServeTLS(certFile, keyFile))
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

func registerCommands(server *gonsole.Server) {
	server.RegisterCommand(&gonsole.Command{
		Name: "hi",
		Note: "打印 hi console",
		Flag: 0,
		Handler: func(session road.Session, args []string) (*gonsole.Response, error) {
			var bean struct {
				Text string
			}

			bean.Text = "hello world"
			return gonsole.NewDefaultResponse(bean), nil
		},
	})

	server.RegisterCommand(&gonsole.Command{
		Name: "test_slice_sort_table_by_head",
		Note: "测试slice排序",
		Flag: gonsole.FlagPublic,
		Handler: func(session road.Session, args []string) (*gonsole.Response, error) {
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
		Name: "test_element_struct",
		Note: "测试element_plus结构体",
		Flag: gonsole.FlagPublic | gonsole.FlagInvisible,
		Handler: func(session road.Session, args []string) (*gonsole.Response, error) {
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

			return gonsole.NewTableResponse(bean), nil
		},
	})

	server.RegisterCommand(&gonsole.Command{
		Name:    "test_struct_sort_table_by_head",
		Example: "hello world",
		Note:    "测试结构体排序",
		Flag:    gonsole.FlagPublic,
		Handler: func(session road.Session, args []string) (*gonsole.Response, error) {
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

	server.RegisterCommand(&gonsole.Command{
		Name: "test_element_table",
		Note: "测试element_table",
		Flag: gonsole.FlagPublic,
		Handler: func(session road.Session, args []string) (*gonsole.Response, error) {
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
				{"oh", "this is an extremely long sentence, and this will be used for testing column with", 30, now.Add(timex.Day).Format(layout), 0.4, "11"},
				{"oh", "my", 30, now.Add(timex.Day).Format(layout), 0.4, "2"},
			}

			return gonsole.NewTableResponse(beans), nil
		},
	})

	server.RegisterCommand(&gonsole.Command{
		Name: "test_send_stream",
		Note: "测试 road.SendStream()",
		Flag: gonsole.FlagPublic,
		Handler: func(session road.Session, args []string) (*gonsole.Response, error) {
			_ = road.SendStream(session, "", false)
			for i := 0; i < 1000; i++ {
				var text = fmt.Sprintf("%v, ", i)
				_ = road.SendStream(session, text, false)

				if i%5 == 0 {
					_ = road.SendStream(session, "\n", false)
				}

				time.Sleep(time.Millisecond * 20)
			}

			_ = road.SendStream(session, "", true)
			return gonsole.NewDefaultResponse("this is the returned object"), nil
		},
	})

	server.RegisterCommand(&gonsole.Command{
		Name: "test_echo",
		Note: "测试 Echo()",
		Flag: gonsole.FlagPublic,
		Handler: func(session road.Session, args []string) (*gonsole.Response, error) {
			var counter int32 = 0
			for i := 0; i < 100; i++ {
				go func() {
					_ = session.Echo(func() {
						atomic.AddInt32(&counter, 1)
						var text = fmt.Sprintf("i=%d, counter=%d", i, counter)
						_ = road.SendDefault(session, text)
					})
				}()
			}

			return gonsole.NewDefaultResponse("this is the returned object"), nil
		},
	})

	server.RegisterTopic(&gonsole.Topic{
		Name:     "hi",
		Note:     "广播hi console（每5s）",
		Interval: 5 * time.Second,
		Flag:     gonsole.FlagPublic,
		BuildResponse: func() *gonsole.Response {
			return gonsole.NewDefaultResponse("hi console")
		},
	})
}
