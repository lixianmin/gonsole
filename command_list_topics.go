package gonsole

import (
	"github.com/lixianmin/gonsole/tools"
	"sort"
)

/********************************************************************
created:    2019-11-17
author:     lixianmin

Copyright (C) - All Rights Reserved
*********************************************************************/

type CommandListTopics struct {
	BasicResponse
	Topics [][]string `json:"topics"`
}

func newCommandListTopics(topics []*Topic) *CommandListTopics {
	var bean = &CommandListTopics{}
	bean.Operation = "listTopics"
	bean.Timestamp = tools.GetTimestamp()

	// 排序
	sort.Slice(topics, func(i, j int) bool {
		return topics[i].Name < topics[j].Name
	})

	var list = make([][]string, 0, len(topics))
	for i := 0; i < len(topics); i++ {
		var item = topics[i]
		list = append(list, []string{item.Name, item.Note})
	}

	bean.Topics = list
	return bean
}
