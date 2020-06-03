package gonsole

/********************************************************************
created:    2019-11-17
author:     lixianmin

Copyright (C) - All Rights Reserved
*********************************************************************/

type DebugListTopics struct {
	BasicResponse
	Topics []string `json:"topics"`
}

func newDebugListTopics() *DebugListTopics {
	var bean = &DebugListTopics{}
	bean.Operation = "listTopics"
	bean.Timestamp = GetTimestamp()

	bean.Topics = []string{
		"rsi",
	}

	return bean
}
