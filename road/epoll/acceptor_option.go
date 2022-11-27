package epoll

import "time"

/********************************************************************
created:    2020-09-30
author:     lixianmin

Copyright (C) - All Rights Reserved
*********************************************************************/

type acceptorOptions struct {
	ConnChanSize      int           // GetConnChan()返回
	PollBufferSize    int           // poll的事件缓冲的长度
	HeartbeatInterval time.Duration // 心跳间隔
}

type AcceptorOption func(*acceptorOptions)

func WithConnChanSize(size int) AcceptorOption {
	return func(options *acceptorOptions) {
		if size > 0 {
			options.ConnChanSize = size
		}
	}
}

func WithPollBufferSize(size int) AcceptorOption {
	return func(options *acceptorOptions) {
		if size > 0 {
			options.PollBufferSize = size
		}
	}
}

func WithHeartbeatInterval(interval time.Duration) AcceptorOption {
	return func(options *acceptorOptions) {
		if interval > 0 {
			options.HeartbeatInterval = interval
		}
	}
}
