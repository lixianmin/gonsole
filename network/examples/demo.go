package main

import (
	"github.com/lixianmin/gonsole/network"
	"github.com/lixianmin/gonsole/network/component"
	"github.com/lixianmin/logo"
	"strings"
)

/********************************************************************
created:    2020-08-31
author:     lixianmin

Copyright (C) - All Rights Reserved
*********************************************************************/

func main() {
	logo.GetLogger().SetFilterLevel(logo.LevelDebug)
	var accept = network.NewWSAcceptor(":8880")
	var app = network.NewApp(network.AppArgs{
		Acceptor: accept,
	})

	var room = &Room{}
	app.Register(room, component.WithName("room"), component.WithNameFunc(strings.ToLower))
	accept.ListenAndServe()
}
