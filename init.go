package gonsole

/********************************************************************
created:    2020-06-03
author:     lixianmin

Copyright (C) - All Rights Reserved
*********************************************************************/

var logger ILogger = &ConsoleLogger{}

func init() {

}

func Init(log ILogger) {
	if log != nil {
		logger = log
	}
}
