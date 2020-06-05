package gonsole

import "github.com/lixianmin/gonsole/logger"

/********************************************************************
created:    2020-06-03
author:     lixianmin

Copyright (C) - All Rights Reserved
*********************************************************************/

func Init(log logger.ILogger) {
	logger.Init(log)
}
