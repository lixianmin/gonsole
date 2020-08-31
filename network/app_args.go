package network

import "time"

/********************************************************************
created:    2020-08-31
author:     lixianmin

Copyright (C) - All Rights Reserved
*********************************************************************/

type AppArgs struct {
	ListenAddress    string        // 监听地址
	HeartbeatTimeout time.Duration // 心跳超时时间
	DataCompression  bool          // 数据是否压缩
}
