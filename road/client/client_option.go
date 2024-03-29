package client

import "time"

/********************************************************************
created:    2022-09-04
author:     lixianmin

Copyright (C) - All Rights Reserved
*********************************************************************/

type pitayaClientOptions struct {
	requestTimeout     time.Duration // 心跳间隔
	receiverBufferSize int           // sender的发送缓冲区大小
}

type ClientOption func(*pitayaClientOptions)

func WithRequestTimeout(timeout time.Duration) ClientOption {
	return func(options *pitayaClientOptions) {
		if timeout > 0 {
			options.requestTimeout = timeout
		}
	}
}

func WithReceiverBufferSize(size int) ClientOption {
	return func(options *pitayaClientOptions) {
		if size > 0 {
			options.receiverBufferSize = size
		}
	}
}
