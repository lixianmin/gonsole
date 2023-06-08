package epoll

import "time"

/********************************************************************
created:    2020-09-30
author:     lixianmin

Copyright (C) - All Rights Reserved
*********************************************************************/

type acceptorOptions struct {
	LinkChanSize      int           // GetLinkChan()返回
	HeartbeatInterval time.Duration // 心跳间隔
}

type AcceptorOption func(*acceptorOptions)

func newAcceptorOptions() acceptorOptions {
	return acceptorOptions{
		LinkChanSize:      16,
		HeartbeatInterval: 5 * time.Second,
	}
}

func WithLinkChanSize(size int) AcceptorOption {
	return func(options *acceptorOptions) {
		if size > 0 {
			options.LinkChanSize = size
		}
	}
}
