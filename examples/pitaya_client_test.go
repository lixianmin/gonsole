package main

import (
	"fmt"
	"github.com/lixianmin/gonsole/road"
	"github.com/lixianmin/gonsole/road/client"
	"github.com/lixianmin/gonsole/road/epoll"
	"github.com/lixianmin/got/convert"
	"github.com/lixianmin/logo"
	"sync"
	"testing"
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
	var level = logo.LevelDebug
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

	var count = 300
	var wg sync.WaitGroup

	app.OnHandShaken(func(session road.Session) {
		type Challenge struct {
			Nonce int `json:"nonce"`
		}

		for i := 0; i < count; i++ {
			if err := session.Push("player.challenge", Challenge{
				Nonce: i,
			}); err != nil {
				logo.JsonE("session", session.Id(), "err", err)
			}
		}
	})

	for i := 0; i < 100; i++ {
		wg.Add(count)
		if err := pitayaConnect(fmt.Sprintf("127.0.0.1:%d", tcpPort), &wg); err != nil {
			logo.JsonE("err", err)
		}
	}

	//select {}
	wg.Wait()
}

func pitayaConnect(serverAddress string, wg *sync.WaitGroup) error {
	var pClient = client.NewPitayaClient()
	if err := pClient.ConnectTo(serverAddress); err != nil {
		return road.NewError("ConnectFailed", "尝试连接游戏服务器失败，serverAddress=%q", serverAddress)
	}

	go func() {
		for {
			select {
			case msg := <-pClient.GetReceivedChan():
				if msg != nil {
					var bean struct {
						Route string `json:"route"`
						Data  string `json:"data"`
						Err   bool   `json:"err"`
					}

					bean.Route = msg.Route
					bean.Data = convert.String(msg.Data)
					bean.Err = msg.Err

					switch bean.Route {
					default:
						logo.JsonI("data", msg.Data)
					}
				}
				wg.Done()
				break
			}
		}
	}()

	return nil
}
