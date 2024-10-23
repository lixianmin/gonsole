package intern

import "sync/atomic"

/********************************************************************
created:    2024-10-23
author:     lixianmin

Copyright (C) - All Rights Reserved
*********************************************************************/

type connectItem struct {
	lastConnectTs     int64
	isScanner         int32
	connectingCounter int32 // 连接计数器
	flashCloseCounter int32 // 闪断计数器
}

func (my *connectItem) ResetCounter() {
	atomic.StoreInt32(&my.connectingCounter, 0)
	atomic.StoreInt32(&my.flashCloseCounter, 0)
}
