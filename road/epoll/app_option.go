package epoll

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
	Serdes                   []serde.Serde // 支持的serde列表
	HeartbeatInterval        time.Duration // heartbeat间隔
	SessionRateLimitBySecond int           // session每秒限流
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

func WithSessionRateLimitBySecond(limit int) AppOption {
	return func(options *appOptions) {
		if limit > 0 {
			options.SessionRateLimitBySecond = limit
		}
	}
}
