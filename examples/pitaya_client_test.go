package main

import (
	"fmt"
	"github.com/lixianmin/gonsole/road"
	"github.com/lixianmin/gonsole/road/client"
	"github.com/lixianmin/gonsole/road/epoll"
	"github.com/lixianmin/gonsole/road/network"
	"github.com/lixianmin/logo"
	"sync"
	"testing"
	"time"
)

/********************************************************************
created:    2023-01-07
author:     lixianmin

Copyright (C) - All Rights Reserved
*********************************************************************/

func initLogo() {
	var theLogger = logo.GetLogger().(*logo.Logger)
	// 开启异步写标记，提高日志输出性能
	theLogger.AddFlag(logo.LogAsyncWrite)

	// 调整theLogger的filterLevel
	var level = logo.LevelInfo
	theLogger.SetFilterLevel(level)

	// 文件日志
	const flag = logo.FlagDate | logo.FlagTime | logo.FlagShortFile | logo.FlagLevel
	var rollingFile = logo.NewRollingFileHook(logo.RollingFileHookArgs{Flag: flag, FilterLevel: level})
	theLogger.AddHook(rollingFile)
}

func TestPitayaClient(t *testing.T) {
	initLogo()

	var tcpPort = 6666
	var tcpAddress = fmt.Sprintf("127.0.0.1:%d", tcpPort)
	var acceptor = epoll.NewTcpAcceptor(tcpAddress)
	var app = road.NewApp(acceptor,
		road.WithSessionRateLimitBySecond(1000),
	)

	app.OnHandShaken(func(session network.Session) {
		type Challenge struct {
			Nonce int `json:"nonce"`
		}

		for i := 0; i < 100; i++ {
			if err := session.PushByRoute("player.challenge", Challenge{
				Nonce: i,
			}); err != nil {
				logo.JsonE("session", session.Id(), "err", err)
			}
		}
	})

	var wg sync.WaitGroup
	for i := 0; i < 100; i++ {
		wg.Add(1)
		if err := pitayaConnect(fmt.Sprintf("127.0.0.1:%d", tcpPort), &wg); err != nil {
			logo.JsonE("err", err)
		}
	}

	//select {}
	wg.Wait()
}

func pitayaConnect(serverAddress string, wg *sync.WaitGroup) error {

	var pClient = client.NewClient()
	if err := pClient.ConnectTo(serverAddress); err != nil {
		wg.Done()
		return network.NewError("ConnectFailed", "尝试连接游戏服务器失败，serverAddress=%q", serverAddress)
	}

	var timer = time.NewTimer(5 * time.Second)
	go func() {
		defer wg.Done()
		for {
			select {
			case pack := <-pClient.GetReceivedChan():
				logo.JsonI("pack", pack)
				break
			case <-timer.C:
				return
			}
		}
	}()

	return nil
}
