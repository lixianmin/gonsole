package gonsole

/********************************************************************
created:    2019-11-16
author:     lixianmin

Copyright (C) - All Rights Reserved
*********************************************************************/

const (
	OK             = 0   // 正确返回
	InternalError  = 500 // 内部错误
	UnknownError   = 501 // 未知错误
	UnknownCommand = 502 // 未知的debug命令
	InvalidTopic   = 503 // 非法topic

	websocketName = "ws_console"
)
