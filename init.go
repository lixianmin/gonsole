package gonsole

import (
	"github.com/lixianmin/gonsole/logger"
	"time"
)

/********************************************************************
created:    2020-06-03
author:     lixianmin

Copyright (C) - All Rights Reserved
*********************************************************************/

// 进程启动的时间
var startProcessTime = time.Now()

func Init(log logger.ILogger) {
	logger.Init(log)
}
