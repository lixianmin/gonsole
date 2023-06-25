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
	HeartbeatInterval        time.Duration
	SessionRateLimitBySecond int // session每秒限流
	Serde                    serde.Serde
}

type AppOption func(*appOptions)

func WithSessionRateLimitBySecond(limit int) AppOption {
	return func(options *appOptions) {
		if limit > 0 {
			options.SessionRateLimitBySecond = limit
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

func WithSerde(serde serde.Serde) AppOption {
	return func(options *appOptions) {
		if serde != nil {
			options.Serde = serde
		}
	}
}
