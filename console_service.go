package gonsole

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
	hintRqt struct {
		Head string `json:"head"`
	}

	hintRe struct {
		Hints []string `json:"hints"`
	}

	commandRqt struct {
		Command string `json:"command"`
	}

	ConsoleService struct {
		server *Server
	}
)

func newConsoleService(server *Server) *ConsoleService {
	var service = &ConsoleService{
		server: server,
	}

	return service
}

func (my *ConsoleService) Command(ctx context.Context, request *commandRqt) (*CommandRe, error) {
	var session = bugfly.GetSessionFromCtx(ctx)

	var args = commandPattern.Split(request.Command, -1)
	var name = args[0]
	var cmd, _ = my.server.GetCommand(name).(*Command)
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

	var client, _ = session.Attachment().Get1(ifs.KeyClient).(*Client)
	var ret, err = cmd.Run(client, args)
	return ret, err
}

func (my *ConsoleService) Hint(ctx context.Context, request *hintRqt) (*hintRe, error) {
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

	var result = &hintRe{Hints: hints}
	return result, nil
}
