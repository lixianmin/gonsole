package road

/********************************************************************
created:    2020-09-30
author:     lixianmin

Copyright (C) - All Rights Reserved
*********************************************************************/

type appOptions struct {
	DataCompression          bool // 数据是否压缩
	SessionRateLimitBySecond int  // session每秒限流
}

type AppOption func(*appOptions)

func WithDataCompression(compression bool) AppOption {
	return func(options *appOptions) {
		options.DataCompression = compression
	}
}

func WithSessionRateLimitBySecond(limit int) AppOption {
	return func(options *appOptions) {
		if limit > 0 {
			options.SessionRateLimitBySecond = limit
		}
	}
}
