package gonsole

import (
	"sort"
)

/********************************************************************
created:    2019-11-16
author:     lixianmin

Copyright (C) - All Rights Reserved
*********************************************************************/

type DebugHelp struct {
	BasicResponse
	Commands [][]string `json:"commands"`
}

func newDebugHelp(commands []Command) *DebugHelp {
	var bean = &DebugHelp{}
	bean.Operation = "help"
	bean.Timestamp = GetTimestamp()

	// 排序
	sort.Slice(commands, func(i, j int) bool {
		return commands[i].Name < commands[j].Name
	})

	var list = make([][]string, 0, len(commands))
	for i := 0; i < len(commands); i++ {
		var item = commands[i]
		list = append(list, []string{item.Name, item.Remark})
	}

	bean.Commands = list

	return bean
}
