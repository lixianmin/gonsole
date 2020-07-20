package gonsole

import (
	"encoding/json"
	"github.com/gorilla/websocket"
	"github.com/lixianmin/gonsole/logger"
	"github.com/lixianmin/gonsole/tools"
	"github.com/lixianmin/got/loom"
	"runtime/debug"
	"strings"
	"sync"
	"time"
)

/********************************************************************
created:    2019-11-16
author:     lixianmin

Copyright (C) - All Rights Reserved
*********************************************************************/

const (
	// Send pings to peer with this period. Must be less than pongWait.
	pingPeriod = 20 * time.Second

	readDeadline  = 60 * time.Second
	writeDeadline = 60 * time.Second
)

type Client struct {
	wc             *loom.WaitClose
	remoteAddress  string
	writeChan      chan []byte
	messageChan    chan IMessage
	server         *Server
	topics         map[string]struct{}
	isAuthorized   bool
	onCloseHandler func()
	Attachment     sync.Map
}

// newClient 创建一个新的client对象
func newClient(server *Server, conn *websocket.Conn) *Client {
	const chanSize = 8
	var readChan = make(chan IBean, chanSize)
	var messageChan = make(chan IMessage, chanSize)

	var client = &Client{
		wc:            loom.NewWaitClose(),
		remoteAddress: conn.RemoteAddr().String(),
		writeChan:     make(chan []byte, chanSize),
		messageChan:   messageChan,
		server:        server,
		topics:        make(map[string]struct{}),
		isAuthorized:  server.args.IsDefaultAuthorized,
	}

	go client.goReadPump(conn, readChan)
	go client.goWritePump(conn, readChan)
	go client.goLoop(readChan)
	return client
}

func (client *Client) goReadPump(conn *websocket.Conn, readChan chan<- IBean) {
	defer loom.DumpIfPanic()

	const maxMessageSize = 65536
	conn.SetReadLimit(maxMessageSize)

	_ = conn.SetReadDeadline(time.Now().Add(readDeadline))
	// 据说h5中的websocket会自动回复pong消息，但需要验证
	// 如果web端无法及时返回pong消息的话，会引起ReadDeadline超时，因此会引发ReadMessage()的websocket.CloseGoingAway
	// ，此时调用client.Close()请求断开链接
	conn.SetPongHandler(func(string) error {
		_ = conn.SetReadDeadline(time.Now().Add(readDeadline))
		return nil
	})

	for {
		// 返回的第一个参数是messageType
		_, message, err := conn.ReadMessage()

		if err != nil {
			// CloseGoingAway: indicates that an endpoint is "going away", such as a server going down or a browser having navigated away from a page.
			// https://tools.ietf.org/html/rfc6455
			if websocket.IsUnexpectedCloseError(err, websocket.CloseNormalClosure, websocket.CloseGoingAway) {
				logger.Info("[goReadPump(%q)] unexpected disconnect, err=%q", client.GetRemoteAddress(), err)
			} else {
				logger.Info("[goReadPump(%q)] disconnect normally, err=%q", client.GetRemoteAddress(), err)
			}

			client.Close()
			break
		}

		// 只要读到消息了，就可以重置readDeadline
		_ = conn.SetReadDeadline(time.Now().Add(readDeadline))

		var basicBean = BasicRequest{}
		err = json.Unmarshal(message, &basicBean)
		if err != nil {
			logger.Warn("[goReadPump(%q)] Invalid message, err=%q, message=%q", client.GetRemoteAddress(), err, message)
			continue
		}

		var bean = createBean(basicBean.Operation)
		if bean == nil {
			logger.Warn("[goReadPump(%q)] Invalid bean.Operation=", client.GetRemoteAddress(), basicBean.Operation)
			continue
		}

		err = json.Unmarshal(message, &bean)
		if err != nil {
			logger.Warn("[goReadPump(%q)] Invalid message, err=%q", client.GetRemoteAddress(), err)
			continue
		}

		select {
		case readChan <- bean:
		case <-client.wc.CloseChan:
			return
		}
	}
}

/*
	goLoop 是client的主循环。
	1. goLoop()不能与goWritePump()合并为一个。早期的确是这样设计，后来发现有deadlock:在处理订阅消息的cmd时，最终需要调用sendBean()
		发送数据到writeChan，但是由于生产者、消费者由同一个loop处理，导致在生产的过程中无法同时消费，因此导致了deadlock
	2.因为是主循环，所以相关的容器类会放到这里，比如topics
*/
func (client *Client) goLoop(readChan <-chan IBean) {
	defer loom.DumpIfPanic()
	// 本client订阅的topic列表
	var messageChan <-chan IMessage = client.messageChan

	for {
		select {
		case bean := <-readChan:
			switch bean := bean.(type) {
			case *Subscribe:
				loopClientSubscribe(client, bean)
			case *Unsubscribe:
				loopClientUnsubscribe(client, bean)
			case *CommandRequest:
				loopClientCommandRequest(client, bean.RequestId, bean.Command)
			case *HintRequest:
				client.SendBean(newHintResponse(bean.Head, client.server.getCommands(), client.isAuthorized))
			case *Ping:
				var pong = &Pong{BasicResponse{Operation: "pong"}}
				client.SendBean(pong)
			default:
				logger.Error("unexpected bean type: %T", bean)
			}
		case msg := <-messageChan:
			switch msg := msg.(type) {
			default:
				logger.Error("unexpected message type: %T", msg)
			}
		case <-client.wc.CloseChan:
			if nil != client.onCloseHandler {
				client.onCloseHandler()
			}
			return
		}
	}
}

func loopClientSubscribe(client *Client, bean *Subscribe) {
	var topicId = bean.TopicId
	var topic = client.server.getTopic(topicId)
	if topic == nil || !(topic.IsPublic || client.isAuthorized) {
		client.SendBean(newBadRequestRe(bean.RequestId, InvalidTopic, "尝试订阅非法topic"))
		return
	}

	if _, ok := client.topics[topicId]; ok {
		client.SendBean(newBadRequestRe(bean.RequestId, InvalidOperation, "重复订阅同一个主题"))
		return
	}

	topic.addClient(client)
	client.topics[topicId] = struct{}{}
	client.SendBean(newSubscribeRe(bean.RequestId, topicId))
	//client.SendBean(topic.BuildData())
}

func loopClientUnsubscribe(client *Client, bean *Unsubscribe) {
	var topicId = bean.TopicId

	var topic = client.server.getTopic(topicId)
	if topic == nil {
		client.SendBean(newBadRequestRe(bean.RequestId, InvalidTopic, "尝试取消非法topic"))
		return
	}

	if _, ok := client.topics[topicId]; !ok {
		client.SendBean(newBadRequestRe(bean.RequestId, InvalidOperation, "尝试取消未订阅主题"))
		return
	}

	topic.removeClient(client)
	delete(client.topics, topicId)
	client.SendBean(newUnsubscribeRe(bean.RequestId, topicId))
}

func loopClientCommandRequest(client *Client, requestId string, command string) {
	var texts = strings.Split(command, " ")
	var name = texts[0]
	var cmd = client.server.getCommand(name)
	// 要么是public方法，要么是authorized了
	if cmd != nil && cmd.Name == name && (cmd.IsPublic || client.isAuthorized) {
		// 防止panic
		defer func() {
			if rec := recover(); rec != nil {
				logger.Error("panic - %v \n%s", rec, debug.Stack())
			}
		}()

		cmd.Handler(client, texts)
	} else {
		client.SendBean(newBadRequestRe(requestId, InternalError, command))
	}
}

// 这个不使用启goroutine去写client.writeChan，虽然不卡死了，但是无法保证顺序了，这就完蛋了
func (client *Client) innerSendBytes(data []byte) {
	select {
	case client.writeChan <- data:
	case <-client.wc.CloseChan:
	}
}

func (client *Client) SendBean(bean interface{}) {
	if bean != nil {
		var jsonBytes, err = tools.MarshalUnescape(bean)
		if err == nil {
			client.innerSendBytes(jsonBytes)
		} else {
			logger.Warn("[SendBean()] Can not marshal bean=%v, err=%s", bean, err)
		}
	}
}

func (client *Client) OnClose(handler func()) {
	client.onCloseHandler = handler
}

func (client *Client) sendMessage(msg IMessage) {
	select {
	case client.messageChan <- msg:
	case <-client.wc.CloseChan:
	}
}

func (client *Client) GetRemoteAddress() string {
	return client.remoteAddress
}

func (client *Client) Close() {
	_ = client.wc.Close()
}
