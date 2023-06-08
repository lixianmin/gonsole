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
	conn               net.Conn
	serde              serde.Serde
	handshake          serde.HandshakeInfo
	receivedPacketChan chan serde.Packet
	connectState       int32

	packetEncoder    codec.PacketEncoder
	packetDecoder    codec.PacketDecoder
	requestTimeout   time.Duration
	nextId           uint32
	messageEncoder   message.Encoder
	handshakeRequest *HandshakeRequest
	wc               loom.WaitClose
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

		connectState:       StateHandshake,
		packetEncoder:      codec.NewPomeloPacketEncoder(),
		packetDecoder:      codec.NewPomeloPacketDecoder(),
		receivedPacketChan: make(chan serde.Packet, options.receiverBufferSize),
		requestTimeout:     options.requestTimeout,
		messageEncoder:     message.NewMessagesEncoder(false),
	}

	return client
}

func (my *Client) goLoop(later loom.Later) {
	var closeChan = my.wc.C()
	defer my.Close()

	var heartbeatTicker = later.NewTicker(10 * time.Second)

	for {
		select {
		case <-heartbeatTicker.C:
			p, _ := my.packetEncoder.Encode(codec.Heartbeat, []byte{})
			if _, err := my.conn.Write(p); err != nil {
				logo.Info("error sending heartbeat to server: %s", err.Error())
				return
			}
		case <-closeChan:
			return
		}
	}
}

func (my *Client) goReceiveData(later loom.Later) {
	defer my.Close()

	//var data [512]byte // 这种方式声明的data是一个实际存储在栈上的array
	var buffer = make([]byte, 1024)
	var stream = &iox.OctetsStream{}
	var reader = iox.NewOctetsReader(stream)

	for my.IsConnected() {
		var num, err1 = my.conn.Read(buffer)
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
			if err := my.onReceivedPacket(pack); err != nil {
				logo.JsonI("err", err)
				return
			}
		}
	}
}

func (my *Client) onReceivedPacket(pack serde.Packet) error {
	switch pack.Kind {
	case serde.Handshake:
		if err := my.onReceiveHandshake(pack); err != nil {
			return err
		}
	case serde.Heartbeat:
	case serde.Kick:
		return ErrKicked
	default:
		my.receivedPacketChan <- pack
	}

	return nil
}

func (my *Client) onReceiveHandshake(pack serde.Packet) error {
	var info = serde.HandshakeInfo{}
	if err := my.serde.Deserialize(pack.Data, &info); err != nil {
		return err
	}

	logo.Debug("got handshake from server, data: %v", info)
	my.handshake = info
	atomic.StoreInt32(&my.connectState, StateConnected)
	return nil
}

// Close disconnects the client
func (my *Client) Close() error {
	return my.wc.Close(func() error {
		if my.IsConnected() {
			atomic.StoreInt32(&my.connectState, StateNone)
			_ = my.conn.Close()
		}
		return nil
	})
}

// ConnectTo connects to the server at addr, for now the only supported protocol is tcp
// if tlsConfig is sent, it connects using TLS
func (my *Client) ConnectTo(addr string, tlsConfig ...*tls.Config) error {
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

	my.conn = conn
	loom.Go(my.goLoop)        // goLoop需要从receivedPacketChan中取packets，因此必须在goReceiveData前启动, 否则可能导致block
	loom.Go(my.goReceiveData) // goReceiveData需要放到最后, 否则可能导致receivedPacketChan收到的数据乱序

	return nil
}

// todo 这个方法可能有问题，因为websocket的读数据逻辑跟tcp的不一样，但ws_client_conn是单独写的，是不是也能还需要仔细过一遍
// ConnectToWS connects using web socket protocol
func (my *Client) ConnectToWS(addr string, path string, tlsConfig ...*tls.Config) error {
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

	my.conn = newWsClientConn(conn)
	loom.Go(my.goLoop)        // goLoop需要从receivedPacketChan中取packets，因此必须在goReceiveData前启动, 否则可能导致block
	loom.Go(my.goReceiveData) // goReceiveData需要放到最后, 否则可能导致receivedPacketChan收到的数据乱序

	return nil
}

func (my *Client) Push(route string, v interface{}) error {
	if my.wc.IsClosed() {
		return nil
	}

	//var kind, ok = my.handshake.RouteKinds[route]
	//if !ok {
	//	return road.ErrInvalidRoute
	//}
	//
	//var data, err1 = my.serde.Serialize(v)
	//if err1 != nil {
	//	return err1
	//}
	//
	//var pack = serde.Packet{Kind: kind, Data: data}
	//var err2 = my.writePacket(pack)
	//return err2
	return nil
}

// sendMsg sends the request to the server
func (my *Client) sendMsg(msgType message.Kind, route string, data []byte) (uint, error) {
	var msg = message.Message{
		Type:  msgType,
		Id:    uint(atomic.AddUint32(&my.nextId, 1)),
		Route: route,
		Data:  data,
		Err:   false,
	}

	var encMsg, err = my.messageEncoder.Encode(&msg)
	if err != nil {
		return 0, err
	}

	p, err := my.packetEncoder.Encode(codec.Data, encMsg)
	if err != nil {
		return 0, err
	}

	_, err = my.conn.Write(p)
	return msg.Id, err
}

// GetReceivedChan return the incoming message channel
func (my *Client) GetReceivedChan() chan serde.Packet {
	return my.receivedPacketChan
}

// IsConnected return the connection status
func (my *Client) IsConnected() bool {
	return atomic.LoadInt32(&my.connectState) > StateNone
}
