package beans

import (
	"context"
	"github.com/lixianmin/gonsole/ifs"
	"sort"
	"strings"
)

/********************************************************************
created:    2020-08-31
author:     lixianmin

Copyright (C) - All Rights Reserved
*********************************************************************/

type (
	Hint struct {
		Head string `json:"head"`
	}

	HintRe struct {
		Hints []string `json:"hints"`
	}

	ConsoleService struct {
		server       ifs.Server
		isAuthorized bool
	}
)

func NewConsoleService(server ifs.Server, isAuthorized bool) *ConsoleService {
	var service = &ConsoleService{
		server:       server,
		isAuthorized: false,
	}

	return service
}

func (my *ConsoleService) Hint(ctx context.Context, request *Hint) (*HintRe, error) {
	var head = strings.TrimSpace(request.Head)
	var commands = my.server.GetCommands()
	var hints = make([]string, 0, len(commands))
	for _, cmd := range commands {
		if (my.isAuthorized || cmd.CheckPublic()) && strings.HasPrefix(cmd.GetName(), head) {
			hints = append(hints, cmd.GetName())
		}
	}

	sort.Slice(hints, func(i, j int) bool {
		return hints[i] < hints[j]
	})

	var result = &HintRe{Hints: hints}
	return result, nil
}
