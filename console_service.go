package gonsole

import (
	"context"
	"fmt"
	"github.com/lixianmin/gonsole/ifs"
	"github.com/lixianmin/gonsole/road/network"
	"github.com/lixianmin/got/convert"
	"github.com/lixianmin/logo"
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
		Name string
		Note string
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
	var session = network.GetSessionFromCtx(ctx)

	var args = commandPattern.Split(request.Command, -1)
	var name = args[0]
	var cmd, _ = my.server.getCommand(name).(*Command)
	if cmd == nil {
		return nil, fmt.Errorf("invalid cmd name=%s", name)
	}

	// 要么是public方法，要么是authorized了
	var isAuthorized = isAuthorized(session)
	if !cmd.IsPublic() && !isAuthorized {
		return nil, fmt.Errorf("need auth")
	}

	// 防止panic
	defer func() {
		if rec := recover(); rec != nil {
			logo.Error("panic - %v \n%s", rec, debug.Stack())
		}
	}()

	var client = getClient(session)
	if client == nil {
		return nil, fmt.Errorf("client=nil")
	}

	var ret, err = cmd.Run(client, args)
	return ret, err
}

func (my *ConsoleService) Sub(ctx context.Context, request *subRqt) (*Response, error) {
	var session = network.GetSessionFromCtx(ctx)
	var client = getClient(session)
	if client == nil {
		return nil, network.NewError("NilClient", "client=nil")
	}

	var name = request.Topic
	var topic = my.server.getTopic(name)

	if topic == nil || !(topic.IsPublic() || isAuthorized(session)) {
		return nil, network.NewError("InvalidTopic", "尝试订阅非法topic")
	}

	if _, ok := client.topics[name]; ok {
		return nil, network.NewError("RepeatedSubscribe", "重复订阅同一个主题")
	}

	topic.addClient(client)
	client.topics[name] = struct{}{}

	var message = fmt.Sprintf("订阅成功，topic=%s", name)
	return NewDefaultResponse(message), nil
}

func (my *ConsoleService) Unsub(ctx context.Context, request *subRqt) (*Response, error) {
	var session = network.GetSessionFromCtx(ctx)
	var client = getClient(session)
	if client == nil {
		return nil, network.NewError("NilClient", "client=nil")
	}

	var name = request.Topic
	var topic = my.server.getTopic(name)
	if topic == nil {
		return nil, network.NewError("InvalidTopic", "尝试取消非法topic")
	}

	if _, ok := client.topics[name]; !ok {
		return nil, network.NewError("RepeatedSubscribe", "尝试取消未订阅主题")
	}

	topic.removeClient(client)
	delete(client.topics, name)

	var message = fmt.Sprintf("退订成功，topic=%s", name)
	return NewDefaultResponse(message), nil
}

func (my *ConsoleService) Hint(ctx context.Context, request *hintRqt) ([]byte, error) {
	var session = network.GetSessionFromCtx(ctx)
	var isAuthorized = isAuthorized(session)

	var head = strings.TrimSpace(request.Head)
	var commands = my.server.getCommands()

	var results = make([]hintRe, 0, len(commands)+len(subUnsubNames))

	for i := range subUnsubNames {
		if strings.HasPrefix(subUnsubNames[i], head) {
			results = append(results, hintRe{subUnsubNames[i], subUnsubNotes[i]})
		}
	}

	for _, cmd := range commands {
		if (isAuthorized || cmd.IsPublic()) && !cmd.IsInvisible() && strings.HasPrefix(cmd.GetName(), head) {
			results = append(results, hintRe{cmd.GetName(), cmd.GetNote()})
		}
	}

	sort.Slice(results, func(i, j int) bool {
		return results[i].Name < results[j].Name
	})

	return convert.ToJson(results), nil
}

func (my *ConsoleService) Default(ctx context.Context, request []byte) ([]byte, error) {
	return nil, nil // 为了推送 console.default
}

func (my *ConsoleService) Html(ctx context.Context, request []byte) ([]byte, error) {
	return nil, nil // 为了推送 console.html
}

func isAuthorized(session network.Session) bool {
	return session.Attachment().Bool(ifs.KeyIsAuthorized)
}

func getClient(session network.Session) *Client {
	var client, _ = session.Attachment().Get1(ifs.KeyClient).(*Client)
	return client
}
