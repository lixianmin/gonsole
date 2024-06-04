package main

import (
	"context"
	"fmt"
	"github.com/lixianmin/gonsole"
	"github.com/lixianmin/gonsole/road"
	"github.com/lixianmin/gonsole/road/client"
	"github.com/lixianmin/gonsole/road/component"
	"github.com/lixianmin/gonsole/road/epoll"
	"github.com/lixianmin/gonsole/road/serde"
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

type PlayerGroup struct {
}

type GetPlayerInfo struct {
	Id int32 `json:"rid"`
}

type GetPlayerInfoRe struct {
	Id     int32 `json:"rid"`
	UserId int64 `json:"uid"`
}

func (my *PlayerGroup) GetPlayerInfo(ctx context.Context, request *GetPlayerInfo) (*GetPlayerInfoRe, error) {
	var response = &GetPlayerInfoRe{
		Id:     request.Id,
		UserId: 123,
	}

	return response, nil
}

func TestRoadClient(t *testing.T) {
	initLogo()

	var tcpPort = 6666
	var tcpAddress = fmt.Sprintf("127.0.0.1:%d", tcpPort)
	var acceptor = epoll.NewTcpAcceptor(tcpAddress)
	var app = road.NewApp(acceptor,
		epoll.WithSessionRateLimitBySecond(1000),
	)

	var group = &PlayerGroup{}
	_ = app.Register(group, component.WithName("player"), component.WithNameFunc(gonsole.ToSnakeName))

	app.OnHandShaken(func(session road.Session) {
		for i := 0; i < 100; i++ {
			//if err := session.Send("player.get_player_info", GetPlayerInfo{
			//	Id: int32(i),
			//}); err != nil {
			//	logo.JsonE("session", session.Id(), "err", err)
			//}
		}
	})

	var wg sync.WaitGroup
	for i := 0; i < 100; i++ {
		wg.Add(1)
		if err := roadConnect(fmt.Sprintf("127.0.0.1:%d", tcpPort), &wg); err != nil {
			logo.JsonE("err", err)
		}
	}

	//select {}
	wg.Wait()
}

func roadConnect(serverAddress string, wg *sync.WaitGroup) error {
	var pClient = client.NewClient()
	if err := pClient.Connect(serverAddress, func(bean *serde.JsonHandshake) {
		var request = &GetPlayerInfo{
			Id: 100,
		}

		var response GetPlayerInfoRe

		_ = pClient.Request("player.get_player_info", request, &response, func(err *road.Error) {
			logo.JsonI("response", response, "err", err)
		})
	}); err != nil {
		wg.Done()
		return road.NewError("ConnectFailed", "尝试连接游戏服务器失败，serverAddress=%q", serverAddress)
	}

	var timer = time.NewTimer(5 * time.Second)
	go func() {
		defer wg.Done()
		for {
			select {
			case <-timer.C:
				return
			}
		}
	}()

	return nil
}
