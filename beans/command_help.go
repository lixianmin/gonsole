package beans

import (
	"github.com/lixianmin/gonsole/ifs"
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

func getHelpCommands(commands []ifs.Command, isAuthorized bool) []ifs.Command {
	if isAuthorized {
		return commands
	}

	var publicCommands = make([]ifs.Command, 0, 4)
	for _, cmd := range commands {
		if cmd.CheckPublic() {
			publicCommands = append(publicCommands, cmd)
		}
	}

	return publicCommands
}

func getHelpTopics(topics []ifs.Topic, isAuthorized bool) []ifs.Topic {
	if isAuthorized {
		return topics
	}

	var publicTopics = make([]ifs.Topic, 0, 4)
	for _, topic := range topics {
		if topic.CheckPublic() {
			publicTopics = append(publicTopics, topic)
		}
	}

	return publicTopics
}

func NewCommandHelp(commands []ifs.Command, topics []ifs.Topic, isAuthorized bool) *CommandHelp {
	var bean = &CommandHelp{}
	bean.Operation = "help"
	bean.Timestamp = tools.GetTimestamp()

	commands = getHelpCommands(commands, isAuthorized)
	topics = getHelpTopics(topics, isAuthorized)

	// commands
	{
		// 排序
		sort.Slice(commands, func(i, j int) bool {
			return commands[i].GetName() < commands[j].GetName()
		})

		var list = make([][]string, 0, len(commands)+2)
		list = append(list, []string{"sub xxx", "订阅主题，例：sub top"})
		list = append(list, []string{"unsub xxx", "取消订阅主题，例：unsub top"})

		for i := 0; i < len(commands); i++ {
			var item = commands[i]
			list = append(list, []string{item.GetName(), item.GetNote()})
		}

		bean.Commands = list
	}

	// topics
	{
		// 排序
		sort.Slice(topics, func(i, j int) bool {
			return topics[i].GetName() < topics[j].GetName()
		})

		var list = make([][]string, 0, len(topics))
		for i := 0; i < len(topics); i++ {
			var item = topics[i]
			list = append(list, []string{item.GetName(), item.GetNote()})
		}

		bean.Topics = list

	}

	return bean
}
