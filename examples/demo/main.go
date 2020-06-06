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
	gonsole.NewServer(mux, gonsole.ServerArgs{
		Port:         webPort,
		TemplatePath: "../../console.html",
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
