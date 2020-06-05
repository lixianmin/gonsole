package gonsole

import "github.com/lixianmin/gonsole/tools"

/********************************************************************
created:    2019-11-17
author:     lixianmin

Copyright (C) - All Rights Reserved
*********************************************************************/

type CommandListTopics struct {
	BasicResponse
	Topics []string `json:"topics"`
}

func newCommandListTopics() *CommandListTopics {
	var bean = &CommandListTopics{}
	bean.Operation = "listTopics"
	bean.Timestamp = tools.GetTimestamp()

	bean.Topics = []string{
		"待实现",
	}

	return bean
}
