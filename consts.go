package gonsole

/********************************************************************
created:    2019-11-16
author:     lixianmin

Copyright (C) - All Rights Reserved
*********************************************************************/

const (
	OK            = 0   // 正确返回
	InternalError = 500 // 内部错误

	UnknownError       = 601 // 未知错误
	UnknownCommand     = 602 // 未知的debug命令
	InvalidTopic       = 603 // 非法topic
	InvalidOperation   = 604 // 非法操作

	websocketName = "ws_console"
)
