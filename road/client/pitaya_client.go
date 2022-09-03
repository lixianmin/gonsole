// Copyright (c) TFG Co. All Rights Reserved.
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in all
// copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.

package client

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"github.com/gobwas/ws"
	"github.com/lixianmin/gonsole/road/codec"
	"github.com/lixianmin/gonsole/road/message"
	"github.com/lixianmin/gonsole/road/packet"
	"github.com/lixianmin/gonsole/road/util/compression"
	"github.com/lixianmin/got/loom"
	"github.com/lixianmin/logo"
	"net"
	"net/url"
	"sync/atomic"
	"time"
)

// HandshakeSys struct
type HandshakeSys struct {
	Dict       map[string]uint16 `json:"dict"`
	Heartbeat  int               `json:"heartbeat"`
	Serializer string            `json:"serializer"`
}

// HandshakeResponse struct
type HandshakeResponse struct {
	Code int          `json:"code"`
	Sys  HandshakeSys `json:"sys"`
}

type PitayaClient struct {
	conn             net.Conn
	isConnected      int32
	packetEncoder    codec.PacketEncoder
	packetDecoder    codec.PacketDecoder
	packetChan       chan *packet.Packet
	IncomingMsgChan  chan *message.Message
	requestTimeout   time.Duration
	nextId           uint32
	messageEncoder   message.Encoder
	handshakeRequest *HandshakeRequest
	wc               loom.WaitClose
}

// New returns a new client
func New(requestTimeout ...time.Duration) *PitayaClient {

	reqTimeout := 5 * time.Second
	if len(requestTimeout) > 0 {
		reqTimeout = requestTimeout[0]
	}

	return &PitayaClient{
		isConnected:    0,
		packetEncoder:  codec.NewPomeloPacketEncoder(),
		packetDecoder:  codec.NewPomeloPacketDecoder(),
		packetChan:     make(chan *packet.Packet, 10),
		requestTimeout: reqTimeout,
		messageEncoder: message.NewMessagesEncoder(false),
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
}

// MsgChannel return the incoming message channel
func (client *PitayaClient) MsgChannel() chan *message.Message {
	return client.IncomingMsgChan
}

// IsConnected return the connection status
func (client *PitayaClient) IsConnected() bool {
	return atomic.LoadInt32(&client.isConnected) == 1
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

	p, err := client.packetEncoder.Encode(packet.Handshake, enc)
	if err != nil {
		return err
	}

	_, err = client.conn.Write(p)
	return err
}

func (client *PitayaClient) handleHandshakeResponse() error {
	buf := bytes.NewBuffer(nil)
	packets, err := client.readPackets(buf)
	if err != nil || len(packets) == 0 {
		return err
	}

	// 这里, 如果一次性读到多个packets的话, 后面的会被扔掉, 不合理
	handshakePacket := packets[0]
	if handshakePacket.Type != packet.Handshake {
		return fmt.Errorf("got first packet from server that is not a handshake, aborting")
	}

	handshake := &HandshakeResponse{}
	if compression.IsCompressed(handshakePacket.Data) {
		handshakePacket.Data, err = compression.InflateData(handshakePacket.Data)
		if err != nil {
			return err
		}
	}

	err = json.Unmarshal(handshakePacket.Data, handshake)
	if err != nil {
		return err
	}

	logo.Debug("got handshake from sv, data: %v", handshake)

	if handshake.Sys.Dict != nil {
		_ = message.SetDictionary(handshake.Sys.Dict)
	}

	p, err := client.packetEncoder.Encode(packet.HandshakeAck, []byte{})
	if err != nil {
		return err
	}
	_, err = client.conn.Write(p)
	if err != nil {
		return err
	}

	atomic.StoreInt32(&client.isConnected, 1)

	go client.sendHeartbeats(handshake.Sys.Heartbeat)
	go client.handleServerMessages()
	go client.handlePackets()

	return nil
}

func (client *PitayaClient) handlePackets() {
	for {
		select {
		case p := <-client.packetChan:
			switch p.Type {
			case packet.Data:
				m, err := message.Decode(p.Data)
				if err != nil {
					logo.Info("error decoding msg from sv: %s", string(m.Data))
				}
				client.IncomingMsgChan <- m
			case packet.Kick:
				logo.Info("got kick packet from the server! disconnecting...")
				client.Close()
			}
		case <-client.wc.C():
			return
		}
	}
}

func (client *PitayaClient) readPackets(buf *bytes.Buffer) ([]*packet.Packet, error) {
	// listen for server messages
	var data = make([]byte, 1024)
	var n = len(data)
	var err error

	for n == len(data) {
		n, err = client.conn.Read(data)
		if err != nil {
			return nil, err
		}

		buf.Write(data[:n])
	}

	packets, err := client.packetDecoder.Decode(buf.Bytes())
	if err != nil {
		logo.Info("error decoding packet from server: %s", err.Error())
		return nil, err
	}

	totalProcessed := 0
	for _, p := range packets {
		totalProcessed += codec.HeadLength + p.Length
	}
	buf.Next(totalProcessed)

	return packets, nil
}

func (client *PitayaClient) handleServerMessages() {
	buf := bytes.NewBuffer(nil)
	defer client.Close()

	for client.IsConnected() {
		packets, err := client.readPackets(buf)
		if err != nil && client.IsConnected() {
			logo.JsonI("err", err)
			break
		}

		for _, p := range packets {
			client.packetChan <- p
		}
	}
}

func (client *PitayaClient) sendHeartbeats(interval int) {
	t := time.NewTicker(time.Duration(interval) * time.Second)
	defer func() {
		t.Stop()
		_ = client.Close()
	}()

	for {
		select {
		case <-t.C:
			p, _ := client.packetEncoder.Encode(packet.Heartbeat, []byte{})
			_, err := client.conn.Write(p)
			if err != nil {
				logo.Info("error sending heartbeat to server: %s", err.Error())
				return
			}
		case <-client.wc.C():
			return
		}
	}
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
	client.IncomingMsgChan = make(chan *message.Message, 10)

	if err = client.handleHandshake(); err != nil {
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

	client.conn = newClientConn(conn)
	client.IncomingMsgChan = make(chan *message.Message, 10)

	if err = client.handleHandshake(); err != nil {
		return err
	}

	return nil
}

func (client *PitayaClient) handleHandshake() error {
	if err := client.sendHandshakeRequest(); err != nil {
		return err
	}

	if err := client.handleHandshakeResponse(); err != nil {
		return err
	}

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
	encMsg, err := client.messageEncoder.Encode(&msg)
	if err != nil {
		return nil, err
	}

	p, err := client.packetEncoder.Encode(packet.Data, encMsg)
	if err != nil {
		return nil, err
	}

	return p, nil
}

// sendMsg sends the request to the server
func (client *PitayaClient) sendMsg(msgType message.Type, route string, data []byte) (uint, error) {
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
