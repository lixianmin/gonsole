package client

import (
	"context"
	"crypto/tls"
	"github.com/gobwas/ws"
	"github.com/lixianmin/gonsole/road/network"
	"github.com/lixianmin/gonsole/road/serde"
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
	manager            *network.Manager
	session            network.Session
	handshake          serde.HandshakeInfo
	receivedPacketChan chan serde.Packet
	connectState       int32
	wc                 loom.WaitClose
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
		manager:            network.NewManager(2 * time.Second),
		connectState:       StateHandshake,
		receivedPacketChan: make(chan serde.Packet, options.receiverBufferSize),
	}

	return client
}

func (my *Client) goLoop(later loom.Later) {
	var closeChan = my.wc.C()
	defer my.Close()

	var heartbeatTicker = later.NewTicker(5 * time.Second)
	for {
		select {
		case <-heartbeatTicker.C:
			if err := my.session.PushByKind(serde.Heartbeat, nil); err != nil {
				logo.Info("error sending heartbeat to server: %s", err.Error())
				return
			}
		case <-closeChan:
			return
		}
	}
}

func (my *Client) onReceiveHandshake(pack serde.Packet) error {
	var info = serde.HandshakeInfo{}
	if err := my.manager.GetSerde().Deserialize(pack.Data, &info); err != nil {
		return err
	}

	logo.Debug("got handshake from server, data: %v", info)
	my.handshake = info
	atomic.StoreInt32(&my.connectState, StateConnected)

	loom.Go(my.goLoop)
	return nil
}

// Close disconnects the client
func (my *Client) Close() error {
	return my.wc.Close(func() error {
		if my.IsConnected() {
			atomic.StoreInt32(&my.connectState, StateNone)
			_ = my.session.Close()
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

	var link = network.NewTcpLink(conn)
	my.session = my.manager.NewSession(link)
	my.session.OnReceivingPacket(my.onReceivingPacket)
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

	var link = network.NewWsLink(conn)
	my.session = my.manager.NewSession(link)
	my.session.OnReceivingPacket(my.onReceivingPacket)

	return nil
}

func (my *Client) onReceivingPacket(pack serde.Packet) error {
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
		return network.ErrPacketProcessed
	}

	return nil
}

func (my *Client) PushByRoute(route string, v interface{}) error {
	return my.session.PushByRoute(route, v)
}

// GetReceivedChan return the incoming message channel
func (my *Client) GetReceivedChan() chan serde.Packet {
	return my.receivedPacketChan
}

// IsConnected return the connection status
func (my *Client) IsConnected() bool {
	return atomic.LoadInt32(&my.connectState) > StateNone
}
