package main

import (
	"fmt"
	"github.com/lixianmin/gonsole"
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
		Port:          webPort,
		TemplatePath:  "console.html",
		UserPasswords: map[string]string{"panda": "1029"},
	})

	server.RegisterCommand(&gonsole.Command{
		Name:     "hi",
		Note:     "打印 hi console",
		IsPublic: true,
		Handler: func(client *gonsole.Client, texts [] string) {
			var bean struct {
				Text string
			}

			bean.Text = "hello <br/> world"
			client.SendBean(bean)
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
