package road

import (
	"github.com/lixianmin/gonsole/road/serde"
	"time"
)

/********************************************************************
created:    2020-09-30
author:     lixianmin

Copyright (C) - All Rights Reserved
*********************************************************************/

type appOptions struct {
	Serdes            []serde.Serde // 支持的serde列表
	HeartbeatInterval time.Duration // heartbeat间隔
	KickInterval      time.Duration // 因为玩家可能切游戏到后台很久去做其它的事情, 因此这个值必须要大一些, 太短很容易被服务器踢的, 参考link_common中的resetReadDeadline
	//SessionRateLimitBySecond int           // session每秒限流
}

type AppOption func(*appOptions)

func WithSerde(serde serde.Serde) AppOption {
	return func(options *appOptions) {
		if serde != nil {
			options.Serdes = append(options.Serdes, serde)
		}
	}
}

func WithHeartbeatInterval(interval time.Duration) AppOption {
	return func(options *appOptions) {
		if interval > 0 {
			options.HeartbeatInterval = interval
		}
	}
}

func WithKickInterval(interval time.Duration) AppOption {
	return func(options *appOptions) {
		if interval > 0 {
			options.KickInterval = interval
		}
	}
}

//func WithSessionRateLimitBySecond(limit int) AppOption {
//	return func(options *appOptions) {
//		if limit > 0 {
//			options.SessionRateLimitBySecond = limit
//		}
//	}
//}
