package gonsole

import (
	"github.com/lixianmin/gonsole/tools"
	"sort"
)

/********************************************************************
created:    2019-11-16
author:     lixianmin

Copyright (C) - All Rights Reserved
*********************************************************************/

type CommandHelp struct {
	BasicResponse
	Commands [][]string `json:"commands"`
	Topics   [][]string `json:"topics"`
}

func newCommandHelp(commands []*Command, topics []*Topic) *CommandHelp {
	var bean = &CommandHelp{}
	bean.Operation = "help"
	bean.Timestamp = tools.GetTimestamp()

	// commands
	{
		// 排序
		sort.Slice(commands, func(i, j int) bool {
			return commands[i].Name < commands[j].Name
		})

		var list = make([][]string, 0, len(commands))
		for i := 0; i < len(commands); i++ {
			var item = commands[i]
			list = append(list, []string{item.Name, item.Note})
		}

		bean.Commands = list
	}

	// topics
	{
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

	}

	return bean
}
