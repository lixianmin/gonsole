package beans

import (
	"time"
)

/********************************************************************
created:    2020-06-03
author:     lixianmin

Copyright (C) - All Rights Reserved
*********************************************************************/

// 进程启动的时间
var processStartAt = time.Now()

// processStartTime 返回进程启动时间，包私有函数
func processStartTime() time.Time {
	return processStartAt
}
