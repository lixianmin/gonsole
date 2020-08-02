package beans

import (
	"fmt"
	"github.com/lixianmin/gonsole/ifs"
	"sort"
)

/********************************************************************
created:    2019-11-16
author:     lixianmin

Copyright (C) - All Rights Reserved
*********************************************************************/

type CommandHelp struct {
	Name string
	Note string
}

func FetchCommandHelp(commands []ifs.Command, isAuthorized bool) []CommandHelp {
	var list = make([]CommandHelp, 0, len(commands)+2)
	list = append(list, CommandHelp{Name: "sub xxx", Note: "订阅主题，例：sub top"})
	list = append(list, CommandHelp{Name: "unsub xxx", Note: "取消订阅主题，例：unsub top"})
	return fetchCommandHelpImpl(list, commands, isAuthorized)
}

func FetchTopicHelp(topics []ifs.Command, isAuthorized bool) []CommandHelp {
	var list = make([]CommandHelp, 0, len(topics))
	return fetchCommandHelpImpl(list, topics, isAuthorized)
}

func fetchCommandHelpImpl(list []CommandHelp, commands []ifs.Command, isAuthorized bool) []CommandHelp {
	// 排序
	sort.Slice(commands, func(i, j int) bool {
		var a, b = commands[i], commands[j]
		if a.CheckBuiltin() && !b.CheckBuiltin() {
			return true
		} else if !a.CheckBuiltin() && b.CheckBuiltin() {
			return false
		}

		return a.GetName() < b.GetName()
	})

	for _, cmd := range commands {
		if isAuthorized || cmd.CheckPublic() {
			list = append(list, CommandHelp{Name: cmd.GetName(), Note: cmd.GetNote()})
		}
	}

	return list
}

func FetchPProfHelp(args []string) []CommandHelp {
	var host = ""
	if len(args) >= 2 {
		host = args[1]
	}

	var list = make([]CommandHelp, 0, 8)
	list = append(list, CommandHelp{
		Name: fmt.Sprintf("<a href=\"%s/debug/pprof\">pprof</a>", host),
		Note: "",
	})

	list = append(list, CommandHelp{
		Name: "<a href=\"https://github.com/lixianmin/writer/blob/master/golang/pprof.md\">参考文档</a>",
		Note: "",
	})

	list = append(list, CommandHelp{
		Name: "CPU (30s)",
		Note: addCopyButton(fmt.Sprintf("go tool pprof -http=: %s/debug/pprof/profile", host)),
	})

	list = append(list, CommandHelp{
		Name: "Heap",
		Note: addCopyButton(fmt.Sprintf("go tool pprof -http=: %s/debug/pprof/heap", host)),
	})

	list = append(list, CommandHelp{
		Name: "goroutine",
		Note: addCopyButton(fmt.Sprintf("curl -G -k %s/debug/pprof/goroutine?debug=2 > tmp.prof", host)),
	})

	return list
}

func addCopyButton(text string) string {
	var result = fmt.Sprintf("%s&nbsp;<input type=\"button\" class=\"copy_button\" onclick=\"copyToClipboard('%s')\" value=\"复制\"/>", text, text)
	return result
}
