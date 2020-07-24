package beans

import (
	"github.com/lixianmin/gonsole/ifs"
	"github.com/lixianmin/gonsole/tools"
	"sort"
	"strings"
)

/********************************************************************
created:    2020-07-20
author:     lixianmin

Copyright (C) - All Rights Reserved
*********************************************************************/

type HintRequest struct {
	BasicRequest
	Head string `json:"head"`
}

type HintResponse struct {
	BasicResponse
	Hints []string `json:"hints"`
}

func NewHintResponse(head string, commands []ifs.Command, isAuthorized bool) *HintResponse {
	var bean = &HintResponse{}
	bean.Operation = "hintResponse"
	bean.Timestamp = tools.GetTimestamp()

	head = strings.TrimSpace(head)
	var hints = make([]string, 0, len(commands))
	for _, cmd := range commands {
		if (isAuthorized || cmd.CheckPublic()) && strings.HasPrefix(cmd.GetName(), head) {
			hints = append(hints, cmd.GetName())
		}
	}

	sort.Slice(hints, func(i, j int) bool {
		return hints[i] < hints[j]
	})

	bean.Hints = hints
	return bean
}
