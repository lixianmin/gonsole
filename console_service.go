package gonsole

import (
	"context"
	"fmt"
	"github.com/lixianmin/gonsole/ifs"
	"github.com/lixianmin/gonsole/road"
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
	var session = road.GetSessionFromCtx(ctx)
	if session == nil {
		return nil, fmt.Errorf("session=nil")
	}

	var args = commandPattern.Split(request.Command, -1)
	var name = args[0]
	var cmd, _ = my.server.getCommand(name).(*Command)
	if cmd == nil {
		return nil, fmt.Errorf("invalid cmd name=%s", name)
	}

	// 要么是public方法，要么是authorized了
	var authorized = isAuthorized(session)
	if !cmd.IsPublic() && !authorized {
		return nil, fmt.Errorf("need auth")
	}

	// 防止panic
	defer func() {
		if rec := recover(); rec != nil {
			logo.Error("panic - %v \n%s", rec, debug.Stack())
		}
	}()

	var ret, err = cmd.Run(session, args)
	return ret, err
}

func fetchTopics(session road.Session) map[string]struct{} {
	const topicKey = "session.sub.topics"
	var topics, ok = session.Attachment().Get2(topicKey)
	if !ok {
		topics = make(map[string]struct{})
		session.Attachment().Set(topicKey, topics)
	}

	return topics.(map[string]struct{})
}

func (my *ConsoleService) Sub(ctx context.Context, request *subRqt) (*Response, error) {
	var session = road.GetSessionFromCtx(ctx)
	if session == nil {
		return nil, road.NewError("NilSession", "session=nil")
	}

	var name = request.Topic
	var topic = my.server.getTopic(name)

	if topic == nil || !(topic.IsPublic() || isAuthorized(session)) {
		return nil, road.NewError("InvalidTopic", "尝试订阅非法topic")
	}

	var topics = fetchTopics(session)
	if _, ok := topics[name]; ok {
		return nil, road.NewError("RepeatedSubscribe", "重复订阅同一个主题")
	}

	topic.addClient(session)
	topics[name] = struct{}{}

	var message = fmt.Sprintf("订阅成功，topic=%s", name)
	return NewDefaultResponse(message), nil
}

func (my *ConsoleService) Unsub(ctx context.Context, request *subRqt) (*Response, error) {
	var session = road.GetSessionFromCtx(ctx)
	if session == nil {
		return nil, road.NewError("NilSession", "session=nil")
	}

	var name = request.Topic
	var topic = my.server.getTopic(name)
	if topic == nil {
		return nil, road.NewError("InvalidTopic", "尝试取消非法topic")
	}

	var topics = fetchTopics(session)
	if _, ok := topics[name]; !ok {
		return nil, road.NewError("RepeatedSubscribe", "尝试取消未订阅主题")
	}

	topic.removeClient(session)
	delete(topics, name)

	var message = fmt.Sprintf("退订成功，topic=%s", name)
	return NewDefaultResponse(message), nil
}

func (my *ConsoleService) Hint(ctx context.Context, request *hintRqt) ([]byte, error) {
	var session = road.GetSessionFromCtx(ctx)
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

//func (my *ConsoleService) Default(ctx context.Context, request []byte) ([]byte, error) {
//	return nil, nil // 为了推送 console.default
//}

// 新的设计下, 服务push回客户端的消息不一定需要有kind
//func (my *ConsoleService) Html(ctx context.Context, request []byte) ([]byte, error) {
//	return nil, nil // 为了推送 console.html
//}

func isAuthorized(session road.Session) bool {
	return session.Attachment().Bool(ifs.KeyIsAuthorized)
}

//func getClient(session road.Session) *Client {
//	var client, _ = session.Attachment().Get1(ifs.KeyClient).(*Client)
//	return client
//}
