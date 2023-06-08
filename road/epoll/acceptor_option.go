package epoll

import "time"

/********************************************************************
created:    2020-09-30
author:     lixianmin

Copyright (C) - All Rights Reserved
*********************************************************************/

type acceptorOptions struct {
	ConnChanSize      int           // GetConnChan()返回
	HeartbeatInterval time.Duration // 心跳间隔
}

type AcceptorOption func(*acceptorOptions)

func newAcceptorOptions() acceptorOptions {
	return acceptorOptions{
		ConnChanSize:      16,
		HeartbeatInterval: 5 * time.Second,
	}
}

func WithConnChanSize(size int) AcceptorOption {
	return func(options *acceptorOptions) {
		if size > 0 {
			options.ConnChanSize = size
		}
	}
}
