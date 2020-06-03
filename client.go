package gonsole

import (
	"encoding/json"
	"github.com/gorilla/websocket"
	"github.com/lixianmin/gocore/loom"
	"sync/atomic"
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

var sendBeanPool = loom.NewGoroutinePool(128)

type Client struct {
	wd            *loom.WaitDispose
	userID        int64
	remoteAddress string
	writeChan     chan []byte
	commandChan   chan ICommand
	server        *Server
	logger        ILogger
	topics        []string
}

type ClientDescription struct {
	RemoteAddress  string `json:"remoteAddress"`
	WriteChanLen   int    `json:"writeChan"`
	CommandChanLen int    `json:"commandChan"`
}

// newClient 创建一个新的client对象
func newClient(server *Server, conn *websocket.Conn) *Client {
	const chanSize = 8
	var readChan = make(chan IBean, chanSize)
	var commandChan = make(chan ICommand, chanSize)

	var client = &Client{
		remoteAddress: conn.RemoteAddr().String(),
		writeChan:     make(chan []byte, chanSize),
		commandChan:   commandChan,
		server:        server,
		logger:        logger,
		wd:            loom.NewWaitDispose(),
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
	// ，此时调用client.Dispose()请求断开链接
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

			client.Dispose()
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
		case <-client.wd.DisposeChan:
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
	const candidateCount = 4
	var commandChan <-chan ICommand = client.commandChan

	for {
		select {
		case bean := <-readChan:
			switch bean := bean.(type) {
			case *Subscribe:
				loopClientSubscribe(client, bean)
			case *Unsubscribe:
				loopClientUnsubscribe(client, bean)
			case *DebugRequest:
				loopClientDebugRequest(client, bean.RequestId, bean.Command)
			case *PingData:
				loopClientPingData(client, *bean)
			default:
				logger.Error("unexpected bean type: %T", bean)
			}
		case cmd := <-commandChan:
			switch cmd := cmd.(type) {
			default:
				logger.Error("unexpected command type: %T", cmd)
			}
		case <-client.wd.DisposeChan:
			return
		}
	}
}

func loopClientSubscribe(client *Client, bean *Subscribe) {
	var topics = make([]string, 0, len(client.topics)+1)
	copy(topics, client.topics)
	client.topics = append(topics, bean.TopicId)
	client.SendBean(newSubscribeRe(bean.RequestId, bean.TopicId))
}

func loopClientUnsubscribe(client *Client, bean *Unsubscribe) {
	var topics []string
	if bean.TopicId != "" {
		topics = make([]string, 0, len(client.topics))
		for i := 0; i < len(client.topics); i++ {
			if client.topics[i] != bean.TopicId {
				topics = append(topics, client.topics[i])
			}
		}
	}

	client.topics = topics
	client.SendBean(newUnsubscribeRe(bean.RequestId, bean.TopicId))
}

func loopClientPingData(client *Client, data PingData) {
	client.SendBeanAsync(newPingDataRe(data.RequestId))
}

// 这个不使用启goroutine去写client.writeChan，虽然不卡死了，但是无法保证顺序了，这就完蛋了
func (client *Client) innerSendBytes(data []byte) {
	select {
	case client.writeChan <- data:
	case <-client.wd.DisposeChan:
	}
}

func (client *Client) SendBytes(data []byte) {
	if data != nil && len(data) > 0 {
		client.innerSendBytes(data)
	}
}

func (client *Client) SendBean(bean interface{}) {
	if bean != nil {
		var jsonBytes, err = json.Marshal(bean)
		if err == nil {
			client.innerSendBytes(jsonBytes)
		} else {
			logger.Warn("[SendBean()] Can not marshal bean=%v, err=%s", bean, err)
		}
	}
}

func (client *Client) SendBeanAsync(bean interface{}) {
	if bean != nil {
		sendBeanPool.Schedule(func() {
			defer loom.DumpIfPanic()
			var jsonBytes, err = json.Marshal(bean)
			if err == nil {
				client.innerSendBytes(jsonBytes)
			} else {
				logger.Warn("[SendBeanAsync()] Can not marshal bean=%v, err=%s", bean, err)
			}
		})
	}
}

// 只所以要定义一个SendCommand()方法，而不是让别人直接使用client.commandChan，是为了证明commandChan这个字段
// 需要放到struct内部(而不是放到goroutine内部)
func (client *Client) SendCommand(cmd ICommand) {
	select {
	case client.commandChan <- cmd:
	case <-client.wd.DisposeChan:
	}
}

func (client *Client) GetUserID() int64 {
	return atomic.LoadInt64(&client.userID)
}

func (client *Client) GetRemoteAddress() string {
	return client.remoteAddress
}

func (client *Client) GetTopics() []string {
	return client.topics
}

func (client *Client) Dispose() {
	client.wd.Dispose()
}
