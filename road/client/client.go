package client

import (
	"context"
	"crypto/tls"
	"github.com/gobwas/ws"
	"github.com/lixianmin/gonsole/road/codec"
	"github.com/lixianmin/gonsole/road/message"
	"github.com/lixianmin/gonsole/road/serde"
	"github.com/lixianmin/got/iox"
	"github.com/lixianmin/got/loom"
	"github.com/lixianmin/logo"
	"net"
	"net/url"
	"sync/atomic"
	"time"
)

/********************************************************************
created:    2022-09-03
author:     lixianmin

Copyright (C) - All Rights Reserved
*********************************************************************/

type Client struct {
	conn      net.Conn
	serde     serde.Serde
	handshake serde.HandshakeInfo

	connectState        int32
	packetEncoder       codec.PacketEncoder
	packetDecoder       codec.PacketDecoder
	receivedMessageChan chan *message.Message
	requestTimeout      time.Duration
	nextId              uint32
	messageEncoder      message.Encoder
	handshakeRequest    *HandshakeRequest
	wc                  loom.WaitClose
}

func NewClient(opts ...ClientOption) *Client {
	// 默认值
	var options = pitayaClientOptions{
		requestTimeout:     5 * time.Second,
		receiverBufferSize: 10,
	}

	// 初始化
	for _, opt := range opts {
		opt(&options)
	}

	var client = &Client{
		serde: &serde.JsonSerde{},

		connectState:        StateHandshake,
		packetEncoder:       codec.NewPomeloPacketEncoder(),
		packetDecoder:       codec.NewPomeloPacketDecoder(),
		receivedMessageChan: make(chan *message.Message, options.receiverBufferSize),
		requestTimeout:      options.requestTimeout,
		messageEncoder:      message.NewMessagesEncoder(false),
	}

	return client
}

func (client *Client) goLoop(later loom.Later) {
	var closeChan = client.wc.C()
	defer client.Close()

	var heartbeatTicker = later.NewTicker(10 * time.Second)

	for {
		select {
		case <-heartbeatTicker.C:
			p, _ := client.packetEncoder.Encode(codec.Heartbeat, []byte{})
			if _, err := client.conn.Write(p); err != nil {
				logo.Info("error sending heartbeat to server: %s", err.Error())
				return
			}
		case <-closeChan:
			return
		}
	}
}

func (client *Client) goReceiveData(later loom.Later) {
	defer client.Close()

	//var data [512]byte // 这种方式声明的data是一个实际存储在栈上的array
	var buffer = make([]byte, 1024)
	var stream = &iox.OctetsStream{}
	var reader = iox.NewOctetsReader(stream)

	for client.IsConnected() {
		var num, err1 = client.conn.Read(buffer)
		if err1 != nil {
			logo.JsonI("err1", err1)
			return
		}

		_ = stream.Write(buffer[:num])
		var packets, err2 = serde.Decode(reader)
		if err2 != nil {
			logo.JsonI("err2", err2)
		}
		stream.Tidy()

		for _, pack := range packets {
			if err := client.onReceivedPacket(pack); err != nil {
				logo.JsonI("err", err)
				return
			}
		}
	}
}

func (client *Client) onReceivedPacket(pack serde.Packet) error {
	switch pack.Kind {
	case serde.Handshake:
		if err := client.onReceiveHandshake(pack); err != nil {
			return err
		}
	case serde.Heartbeat:
	case serde.Kick:
		return ErrKicked
	default:
		msg, err := message.Decode(pack.Data)
		if err != nil {
			return err
		}

		// duplicate msg.Data because it will be forward to another goroutine
		var data = make([]byte, len(msg.Data))
		copy(data, msg.Data)
		msg.Data = data

		client.receivedMessageChan <- msg
	}

	return nil
}

func (client *Client) onReceiveHandshake(pack serde.Packet) error {
	var info = serde.HandshakeInfo{}
	if err := client.serde.Deserialize(pack.Data, &info); err != nil {
		return err
	}

	logo.Debug("got handshake from server, data: %v", info)
	client.handshake = info
	atomic.StoreInt32(&client.connectState, StateConnected)
	return nil
}

// Close disconnects the client
func (client *Client) Close() error {
	return client.wc.Close(func() error {
		if client.IsConnected() {
			atomic.StoreInt32(&client.connectState, StateNone)
			_ = client.conn.Close()
		}
		return nil
	})
}

// ConnectTo connects to the server at addr, for now the only supported protocol is tcp
// if tlsConfig is sent, it connects using TLS
func (client *Client) ConnectTo(addr string, tlsConfig ...*tls.Config) error {
	var conn net.Conn
	var err error
	if len(tlsConfig) > 0 {
		conn, err = tls.Dial("tcp", addr, tlsConfig[0])
	} else {
		conn, err = net.Dial("tcp", addr)
	}

	if err != nil {
		return err
	}

	client.conn = conn
	loom.Go(client.goLoop)        // goLoop需要从receivedPacketChan中取packets，因此必须在goReceiveData前启动, 否则可能导致block
	loom.Go(client.goReceiveData) // goReceiveData需要放到最后, 否则可能导致receivedPacketChan收到的数据乱序

	return nil
}

// todo 这个方法可能有问题，因为websocket的读数据逻辑跟tcp的不一样，但ws_client_conn是单独写的，是不是也能还需要仔细过一遍
// ConnectToWS connects using web socket protocol
func (client *Client) ConnectToWS(addr string, path string, tlsConfig ...*tls.Config) error {
	var uri = url.URL{Scheme: "ws", Host: addr, Path: path}
	var dialer = ws.DefaultDialer

	if len(tlsConfig) > 0 {
		dialer.TLSConfig = tlsConfig[0]
		uri.Scheme = "wss"
	}

	conn, _, _, err := dialer.Dial(context.Background(), uri.String())
	if err != nil {
		return err
	}

	client.conn = newWsClientConn(conn)
	loom.Go(client.goLoop)        // goLoop需要从receivedPacketChan中取packets，因此必须在goReceiveData前启动, 否则可能导致block
	loom.Go(client.goReceiveData) // goReceiveData需要放到最后, 否则可能导致receivedPacketChan收到的数据乱序

	return nil
}

// SendRequest sends a request to the server
func (client *Client) SendRequest(route string, data []byte) (uint, error) {
	return client.sendMsg(message.Request, route, data)
}

// sendMsg sends the request to the server
func (client *Client) sendMsg(msgType message.Kind, route string, data []byte) (uint, error) {
	var msg = message.Message{
		Type:  msgType,
		Id:    uint(atomic.AddUint32(&client.nextId, 1)),
		Route: route,
		Data:  data,
		Err:   false,
	}

	var encMsg, err = client.messageEncoder.Encode(&msg)
	if err != nil {
		return 0, err
	}

	p, err := client.packetEncoder.Encode(codec.Data, encMsg)
	if err != nil {
		return 0, err
	}

	_, err = client.conn.Write(p)
	return msg.Id, err
}

// GetReceivedChan return the incoming message channel
func (client *Client) GetReceivedChan() chan *message.Message {
	return client.receivedMessageChan
}

// IsConnected return the connection status
func (client *Client) IsConnected() bool {
	return atomic.LoadInt32(&client.connectState) > StateNone
}
