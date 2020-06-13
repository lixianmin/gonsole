package gonsole

import (
	"github.com/lixianmin/gonsole/logger"
	"github.com/lixianmin/logo"
	"time"
)

/********************************************************************
created:    2020-06-03
author:     lixianmin

Copyright (C) - All Rights Reserved
*********************************************************************/

// 进程启动的时间
var startProcessTime = time.Now()

func Init(log logo.ILogger) {
	logger.Init(log)
}
