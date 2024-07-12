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

type serdeBuilder func(session Session) serde.Serde

type appOptions struct {
	SerdeBuilders     map[string]serdeBuilder // 支持的serde列表
	HeartbeatInterval time.Duration           // heartbeat间隔
	KickInterval      time.Duration           // 因为玩家可能切游戏到后台很久去做其它的事情, 因此这个值必须要大一些, 太短很容易被服务器踢的, 参考link_common中的resetReadDeadline
}

type AppOption func(*appOptions)

func WithSerdeBuilder(name string, builder serdeBuilder) AppOption {
	return func(options *appOptions) {
		if name != "" && builder != nil {
			options.SerdeBuilders[name] = builder
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
