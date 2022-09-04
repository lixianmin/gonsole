package client

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"github.com/gobwas/ws"
	"github.com/lixianmin/gonsole/road/codec"
	"github.com/lixianmin/gonsole/road/message"
	"github.com/lixianmin/gonsole/road/util/compression"
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

type PitayaClient struct {
	conn                net.Conn
	isConnected         int32
	packetEncoder       codec.PacketEncoder
	packetDecoder       codec.PacketDecoder
	receivedPacketChan  chan *codec.Packet
	receivedMessageChan chan *message.Message
	requestTimeout      time.Duration
	nextId              uint32
	messageEncoder      message.Encoder
	handshakeRequest    *HandshakeRequest
	wc                  loom.WaitClose
}

func NewPitayaClient(opts ...PitayaClientOption) *PitayaClient {

	// 默认值
	var options = pitayaClientOptions{
		requestTimeout:     5 * time.Second,
		receiverBufferSize: 10,
	}

	// 初始化
	for _, opt := range opts {
		opt(&options)
	}

	var client = &PitayaClient{
		isConnected:         0,
		packetEncoder:       codec.NewPomeloPacketEncoder(),
		packetDecoder:       codec.NewPomeloPacketDecoder(),
		receivedPacketChan:  make(chan *codec.Packet, options.receiverBufferSize),
		receivedMessageChan: make(chan *message.Message, options.receiverBufferSize),
		requestTimeout:      options.requestTimeout,
		messageEncoder:      message.NewMessagesEncoder(false),
		handshakeRequest: &HandshakeRequest{
			Sys: HandshakeClientData{
				Platform:    "mac",
				LibVersion:  "0.3.5-release",
				BuildNumber: "20",
				Version:     "2.1",
			},
			User: map[string]interface{}{
				"age": 30,
			},
		},
	}

	return client
}

func (client *PitayaClient) goLoop(later loom.Later) {
	var closeChan = client.wc.C()
	defer client.Close()

	var heartbeatTicker = later.NewTicker(10 * time.Second)

	for {
		select {
		case p := <-client.receivedPacketChan:
			switch p.Kind {
			case codec.Data:
				msg, err := message.Decode(p.Data)
				if err != nil {
					logo.Info("error decoding msg from sv: %s", string(msg.Data))
				}
				client.receivedMessageChan <- msg
			case codec.Kick:
				logo.Info("got kick packet from the server! disconnecting...")
			}
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

func (client *PitayaClient) goReadPackets(later loom.Later) {
	defer client.Close()
	var buffer = &iox.Buffer{}

	for client.IsConnected() {
		packets, err := client.readPackets(buffer)
		if err != nil && client.IsConnected() {
			logo.JsonI("err", err)
			break
		}

		for _, p := range packets {
			client.receivedPacketChan <- p
		}
	}
}

// SetClientHandshakeData sets the data to send inside handshake
func (client *PitayaClient) SetHandshakeRequest(data *HandshakeRequest) {
	client.handshakeRequest = data
}

func (client *PitayaClient) sendHandshakeRequest() error {
	enc, err := json.Marshal(client.handshakeRequest)
	if err != nil {
		return err
	}

	p, err := client.packetEncoder.Encode(codec.Handshake, enc)
	if err != nil {
		return err
	}

	_, err = client.conn.Write(p)
	return err
}

func (client *PitayaClient) handleHandshakeResponse() ([]*codec.Packet, error) {
	var buf = &iox.Buffer{}
	var packets []*codec.Packet
	var err error

	for {
		if packets, err = client.readPackets(buf); err != nil {
			return nil, err
		} else if len(packets) > 0 {
			break
		}
	}

	// 如果一次性读到多个packets的话, 后面的会被扔掉, 不合理
	var handshakePacket = packets[0]
	if handshakePacket.Kind != codec.Handshake {
		return nil, fmt.Errorf("got first packet from server that is not a handshake, aborting")
	}

	var handshake = &HandshakeResponse{}
	if compression.IsCompressed(handshakePacket.Data) {
		handshakePacket.Data, err = compression.InflateData(handshakePacket.Data)
		if err != nil {
			return nil, err
		}
	}

	err = json.Unmarshal(handshakePacket.Data, handshake)
	if err != nil {
		return nil, err
	}

	logo.Debug("got handshake from server, data: %v", handshake)

	if handshake.Sys.Dict != nil {
		_ = message.SetDictionary(handshake.Sys.Dict)
	}

	p, err := client.packetEncoder.Encode(codec.HandshakeAck, []byte{})
	if err != nil {
		return nil, err
	}
	_, err = client.conn.Write(p)
	if err != nil {
		return nil, err
	}

	atomic.StoreInt32(&client.isConnected, 1)
	return packets[1:], nil
}

func (client *PitayaClient) readPackets(buffer *iox.Buffer) ([]*codec.Packet, error) {
	var data [1024]byte // 这种方式声明的data是一个实际存储在栈上的array
	for {
		var n, err = client.conn.Read(data[:])
		if err != nil {
			return nil, err
		}

		buffer.Write(data[:n])
		if n < len(data) {
			break
		}
	}

	return client.packetDecoder.Decode(buffer)
}

// Close disconnects the client
func (client *PitayaClient) Close() error {
	return client.wc.Close(func() error {
		if client.IsConnected() {
			atomic.StoreInt32(&client.isConnected, 0)
			_ = client.conn.Close()
		}
		return nil
	})
}

// ConnectTo connects to the server at addr, for now the only supported protocol is tcp
// if tlsConfig is sent, it connects using TLS
func (client *PitayaClient) ConnectTo(addr string, tlsConfig ...*tls.Config) error {
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
	if err = client.startHandshake(); err != nil {
		return err
	}

	return nil
}

// ConnectToWS connects using web socket protocol
func (client *PitayaClient) ConnectToWS(addr string, path string, tlsConfig ...*tls.Config) error {
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
	if err = client.startHandshake(); err != nil {
		return err
	}

	return nil
}

func (client *PitayaClient) startHandshake() error {
	if err := client.sendHandshakeRequest(); err != nil {
		return err
	}

	var packets, err = client.handleHandshakeResponse()
	if err != nil {
		return err
	}

	// goLoop需要在后面的取剩余packets的前面启动, 否则可能导致block
	loom.Go(client.goLoop)

	// 把剩下的packets放到chan中
	for _, p := range packets {
		client.receivedPacketChan <- p
	}

	// goReadPackets需要放到最后, 否则可能导致receivedPacketChan收到的数据乱序
	loom.Go(client.goReadPackets)
	return nil
}

// SendRequest sends a request to the server
func (client *PitayaClient) SendRequest(route string, data []byte) (uint, error) {
	return client.sendMsg(message.Request, route, data)
}

// SendNotify sends a notification to the server
func (client *PitayaClient) SendNotify(route string, data []byte) error {
	_, err := client.sendMsg(message.Notify, route, data)
	return err
}

func (client *PitayaClient) buildPacket(msg message.Message) ([]byte, error) {
	var encMsg, err = client.messageEncoder.Encode(&msg)
	if err != nil {
		return nil, err
	}

	p, err := client.packetEncoder.Encode(codec.Data, encMsg)
	if err != nil {
		return nil, err
	}

	return p, nil
}

// sendMsg sends the request to the server
func (client *PitayaClient) sendMsg(msgType message.Kind, route string, data []byte) (uint, error) {
	// TODO mount msg and encode
	m := message.Message{
		Type:  msgType,
		Id:    uint(atomic.AddUint32(&client.nextId, 1)),
		Route: route,
		Data:  data,
		Err:   false,
	}

	p, err := client.buildPacket(m)
	if msgType == message.Request {

	}

	if err != nil {
		return m.Id, err
	}

	_, err = client.conn.Write(p)
	return m.Id, err
}

// GetReceivedChan return the incoming message channel
func (client *PitayaClient) GetReceivedChan() chan *message.Message {
	return client.receivedMessageChan
}

// IsConnected return the connection status
func (client *PitayaClient) IsConnected() bool {
	return atomic.LoadInt32(&client.isConnected) == 1
}
