package main

import (
	"fmt"
	"github.com/lixianmin/gonsole/road"
	"github.com/lixianmin/gonsole/road/client"
	"github.com/lixianmin/gonsole/road/epoll"
	"github.com/lixianmin/got/convert"
	"github.com/lixianmin/logo"
	"math/rand"
	"sync"
	"testing"
)

/********************************************************************
created:    2023-01-07
author:     lixianmin

Copyright (C) - All Rights Reserved
*********************************************************************/

func TestPitayaClient(t *testing.T) {
	var tcpPort = 6666
	var tcpAddress = fmt.Sprintf("127.0.0.1:%d", tcpPort)
	var acceptor = epoll.NewTcpAcceptor(tcpAddress)
	var app = road.NewApp(acceptor,
		road.WithSessionRateLimitBySecond(1000),
	)

	var count = 300
	var wg sync.WaitGroup
	wg.Add(count)

	app.OnHandShaken(func(session road.Session) {
		type Challenge struct {
			Nonce int32 `json:"nonce"`
		}

		for i := 0; i < count; i++ {
			if err := session.Push("player.challenge", Challenge{
				Nonce: rand.Int31(),
			}); err != nil {
				logo.JsonE("err", err)
			}
		}
	})

	if err := pitayaConnect(fmt.Sprintf("127.0.0.1:%d", tcpPort), &wg); err != nil {
		logo.JsonE("err", err)
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
					case "player.challenge":
						logo.JsonI("data", bean.Data)
					default:
						logo.JsonI("route", bean.Route, "bean", bean)
					}
				}
				wg.Done()
				break
			}
		}
	}()

	return nil
}
