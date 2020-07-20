package gonsole

import (
	"github.com/lixianmin/gonsole/tools"
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

func newHintResponse(head string, commands []*Command, isAuthorized bool) *HintResponse {
	var bean = &HintResponse{}
	bean.Operation = "hintResponse"
	bean.Timestamp = tools.GetTimestamp()

	head = strings.TrimSpace(head)
	var hints = make([]string, 0, len(commands))
	for _, cmd := range commands {
		if (isAuthorized || cmd.IsPublic) && strings.HasPrefix(cmd.Name, head) {
			hints = append(hints, cmd.Name)
		}
	}

	bean.Hints = hints
	return bean
}
