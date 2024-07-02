package beans

import (
	"fmt"
	"github.com/lixianmin/gonsole/ifs"
	"net/url"
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
	var list = make([]CommandHelp, 0, len(commands)+3)
	list = append(list, CommandHelp{Name: "request", Note: "直接发送request请求, 例: request console.command {\"command\":\"help\"}"})
	list = append(list, CommandHelp{Name: "sub xxx", Note: "订阅主题，例：sub top"})
	list = append(list, CommandHelp{Name: "unsub xxx", Note: "取消订阅主题，例：unsub top"})
	return fetchCommandHelpImpl(list, commands, isAuthorized)
}

func FetchTopicHelp(topics []ifs.Command, isAuthorized bool) []CommandHelp {
	var list = make([]CommandHelp, 0, len(topics))
	return fetchCommandHelpImpl(list, topics, isAuthorized)
}

func fetchCommandHelpImpl(list []CommandHelp, commands []ifs.Command, isAuthorized bool) []CommandHelp {
	for _, cmd := range commands {
		if (isAuthorized || cmd.IsPublic()) && !cmd.IsInvisible() {
			list = append(list, CommandHelp{Name: cmd.GetName(), Note: cmd.GetNote()})
		}
	}

	// 排序
	sort.Slice(list, func(i, j int) bool {
		return list[i].Name < list[j].Name
	})
	return list
}

func FetchPProfHelp(args []string) []CommandHelp {
	var host = ""
	if len(args) >= 2 {
		if u, err := url.Parse(args[1]); err == nil {
			host = u.Scheme + "://" + u.Host
		}
	}

	var list = make([]CommandHelp, 0, 8)
	list = append(list,
		CommandHelp{
			Name: fmt.Sprintf("<a href=\"%s/debug/pprof\">pprof</a>", host),
			Note: "",
		}, CommandHelp{
			Name: "<a href=\"https://github.com/lixianmin/writer/blob/master/golang/pprof.md\">参考文档</a>",
			Note: "",
		}, CommandHelp{
			Name: "CPU (30s)",
			Note: addCopyButton(fmt.Sprintf("go tool pprof -http=: %s/debug/pprof/profile", host)),
		}, CommandHelp{
			Name: "Heap",
			Note: addCopyButton(fmt.Sprintf("go tool pprof -http=: %s/debug/pprof/heap", host)),
		}, CommandHelp{
			Name: "goroutine",
			Note: addCopyButton(fmt.Sprintf("curl -G -k %s/debug/pprof/goroutine?debug=2 > goroutine.list.txt", host)),
		})

	return list
}

func addCopyButton(text string) string {
	var result = fmt.Sprintf("%s&nbsp;<input type=\"button\" class=\"copy_button\" onclick=\"copyToClipboard('%s')\" value=\"复制\"/>", text, text)
	return result
}
