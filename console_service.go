package gonsole

import (
	"context"
	"fmt"
	"github.com/lixianmin/gonsole/ifs"
	"github.com/lixianmin/gonsole/logger"
	"github.com/lixianmin/road"
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

	subRqt struct {
		Topic string `json:"topic"`
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

func (my *ConsoleService) Command(ctx context.Context, request *commandRqt) (*Response, error) {
	var session = road.GetSessionFromCtx(ctx)

	var args = commandPattern.Split(request.Command, -1)
	var name = args[0]
	var cmd, _ = my.server.getCommand(name).(*Command)
	if cmd == nil {
		return nil, fmt.Errorf("invalid cmd name=%s", name)
	}

	// 要么是public方法，要么是authorized了
	var isAuthorized = isAuthorized(session)
	if !cmd.CheckPublic() && !isAuthorized {
		return nil, fmt.Errorf("need auth")
	}

	// 防止panic
	defer func() {
		if rec := recover(); rec != nil {
			logger.Error("panic - %v \n%s", rec, debug.Stack())
		}
	}()

	var client = getClient(session)
	var ret, err = cmd.Run(client, args)
	return ret, err
}

func (my *ConsoleService) Sub(ctx context.Context, request *subRqt) (*Response, error) {
	var session = road.GetSessionFromCtx(ctx)
	var client = getClient(session)
	var name = request.Topic
	var topic = client.server.getTopic(name)

	if topic == nil || !(topic.IsPublic || isAuthorized(session)) {
		return nil, road.NewError("InvalidTopic", "尝试订阅非法topic")
	}

	if _, ok := client.topics[name]; ok {
		return nil, road.NewError("RepeatedSubscribe", "重复订阅同一个主题")
	}

	topic.addClient(client)
	client.topics[name] = struct{}{}

	var message = fmt.Sprintf("订阅成功，topic=%s", name)
	return NewDefaultResponse(message), nil
}

func (my *ConsoleService) Unsub(ctx context.Context, request *subRqt) (*Response, error) {
	var session = road.GetSessionFromCtx(ctx)
	var client = getClient(session)
	var name = request.Topic
	var topic = client.server.getTopic(name)
	if topic == nil {
		return nil, road.NewError("InvalidTopic", "尝试取消非法topic")
	}

	if _, ok := client.topics[name]; !ok {
		return nil, road.NewError("RepeatedSubscribe", "尝试取消未订阅主题")
	}

	topic.removeClient(client)
	delete(client.topics, name)

	var message = fmt.Sprintf("退订成功，topic=%s", name)
	return NewDefaultResponse(message), nil
}

func (my *ConsoleService) Hint(ctx context.Context, request *hintRqt) (*hintRe, error) {
	var session = road.GetSessionFromCtx(ctx)
	var isAuthorized = isAuthorized(session)

	var head = strings.TrimSpace(request.Head)
	var commands = my.server.getCommands()
	var builtins = []string{"sub", "unsub"}
	var hints = make([]string, 0, len(commands)+len(builtins))

	for _, name := range builtins {
		if strings.HasPrefix(name, head) {
			hints = append(hints, name)
		}
	}

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

func isAuthorized(session *road.Session) bool {
	return session.Attachment().Bool(ifs.KeyIsAuthorized)
}

func getClient(session *road.Session) *Client {
	var client, _ = session.Attachment().Get1(ifs.KeyClient).(*Client)
	return client
}
