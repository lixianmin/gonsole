package gonsole

/********************************************************************
created:    2019-11-16
author:     lixianmin

Copyright (C) - All Rights Reserved
*********************************************************************/

type DebugHelp struct {
	BasicResponse
	Commands [][]string `json:"commands"`
}

func newDebugHelp(remoteAddress string) *DebugHelp {
	var bean = &DebugHelp{}
	bean.Operation = "help"
	bean.Timestamp = GetTimestamp()
	bean.Commands = [][]string{
		{"dailyIncome", "计算每每日收入"},
		{"fund", "资金费率"},
		{"hedgeLong size", "对冲多仓，注意markPrice在哪边"},
		{"hedgeShort size", "对冲空仓，注意markPrice在哪边"},
		{"help", "帮助中心"},
		{"ls", "打印主题列表"},
		{"order 6", "查询订单状态"},
		{"sub topicId", "订阅主题"},
		{"sum", "账户概览"},
		{"unsub topicId", "取消订阅，不带topicId则取消所有"},
	}

	return bean
}
