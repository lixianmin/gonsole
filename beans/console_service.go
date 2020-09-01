package beans

import (
	"context"
	"fmt"
	"github.com/lixianmin/bugfly"
	"github.com/lixianmin/gonsole/ifs"
	"github.com/lixianmin/gonsole/logger"
	"regexp"
	"runtime/debug"
	"sort"
	"strings"
)

/********************************************************************
created:    2020-08-31
author:     lixianmin

Copyright (C) - All Rights Reserved
*********************************************************************/

var (
	commandPattern, _ = regexp.Compile(`\s+`)
)

type (
	Hint struct {
		Head string `json:"head"`
	}

	HintRe struct {
		Hints []string `json:"hints"`
	}

	Command struct {
		Command string `json:"command"`
	}

	CommandRe struct {
	}

	ConsoleService struct {
		server ifs.Server
	}
)

func NewConsoleService(server ifs.Server) *ConsoleService {
	var service = &ConsoleService{
		server: server,
	}

	return service
}

func (my *ConsoleService) Command(ctx context.Context, request *Command) (*CommandRe, error) {
	var session = bugfly.GetSessionFromCtx(ctx)

	var args = commandPattern.Split(request.Command, -1)
	var name = args[0]
	var cmd = my.server.GetCommand(name)
	if cmd == nil {
		return nil, fmt.Errorf("invalid cmd name=%s", name)
	}

	// 要么是public方法，要么是authorized了
	var isAuthorized = session.Attachment().Bool(ifs.KeyIsAuthorized)
	if !cmd.CheckPublic() && !isAuthorized {
		return nil, fmt.Errorf("need auth")
	}

	// 防止panic
	defer func() {
		if rec := recover(); rec != nil {
			logger.Error("panic - %v \n%s", rec, debug.Stack())
		}
	}()

	var client = session.Attachment().Get1(ifs.KeyClient).(ifs.Client)
	cmd.Run(client, args)

	var ret = &CommandRe{}
	return ret, nil
}

func (my *ConsoleService) Hint(ctx context.Context, request *Hint) (*HintRe, error) {
	var session = bugfly.GetSessionFromCtx(ctx)
	var isAuthorized = session.Attachment().Bool(ifs.KeyIsAuthorized)

	var head = strings.TrimSpace(request.Head)
	var commands = my.server.GetCommands()
	var hints = make([]string, 0, len(commands))

	for _, cmd := range commands {
		if (isAuthorized || cmd.CheckPublic()) && strings.HasPrefix(cmd.GetName(), head) {
			hints = append(hints, cmd.GetName())
		}
	}

	sort.Slice(hints, func(i, j int) bool {
		return hints[i] < hints[j]
	})

	var result = &HintRe{Hints: hints}
	return result, nil
}
