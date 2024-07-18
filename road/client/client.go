package client

import (
	"crypto/tls"
	"fmt"
	"github.com/lixianmin/gonsole/road"
	"github.com/lixianmin/gonsole/road/serde"
	"github.com/lixianmin/got/convert"
	"github.com/lixianmin/got/iox"
	"github.com/lixianmin/got/loom"
	"github.com/lixianmin/logo"
	"net"
	"sync"
	"sync/atomic"
	"time"
)

/********************************************************************
created:    2023-11-26
author:     lixianmin

Copyright (C) - All Rights Reserved
*********************************************************************/

var globalIdGenerator int64 = 0

type Client struct {
	id        int64
	writeLock sync.Mutex
	writer    *iox.OctetsWriter
	wc        loom.WaitClose
	serde     serde.Serde
	nonce     int32
	conn      net.Conn

	heartbeatInterval  time.Duration
	onHandShaken       func(bean *serde.JsonHandshake)
	requestIdGenerator int32

	routeKinds         map[string]int32
	kindRoutes         map[int32]string
	requestHandlers    map[int32]func([]byte, *road.Error)
	registeredHandlers map[string]func([]byte, *road.Error)
}

func NewClient() *Client {
	var id = atomic.AddInt64(&globalIdGenerator, 1)
	var my = &Client{
		id:                 id,
		writer:             iox.NewOctetsWriter(&iox.OctetsStream{}),
		heartbeatInterval:  time.Minute, // 初始给一个大一些的值, 防止client自己超时, 回头server会重置该值
		routeKinds:         map[string]int32{},
		kindRoutes:         map[int32]string{},
		requestHandlers:    map[int32]func([]byte, *road.Error){},
		registeredHandlers: map[string]func([]byte, *road.Error){},
	}

	return my
}

func (my *Client) Close() error {
	return my.wc.Close(func() error {
		var err = my.conn.Close()
		return err
	})
}

func (my *Client) Connect(address string, opts ...ConnectOption) error {
	var options = &connectOptions{
		serde:        &serde.JsonSerde{},
		tlsConfig:    nil,
		onHandShaken: nil,
	}

	for _, opt := range opts {
		opt(options)
	}

	my.serde = options.serde

	var conn net.Conn
	var err error
	if options.tlsConfig != nil {
		conn, err = tls.Dial("tcp", address, options.tlsConfig)
	} else {
		conn, err = net.Dial("tcp", address)
	}

	if err != nil {
		return err
	}

	my.conn = conn
	my.onHandShaken = options.onHandShaken
	go my.goLoop()

	return nil
}

func (my *Client) goHeartbeat() {
	defer loom.DumpIfPanic()
	defer my.Close()

	var pack = serde.Packet{
		Kind: serde.Heartbeat,
	}

	for !my.wc.IsClosed() {
		_ = my.sendPacket(pack)
		time.Sleep(my.heartbeatInterval)
	}
}

func (my *Client) goLoop() {
	defer loom.DumpIfPanic()
	defer my.Close()

	var buffer = make([]byte, 1024)
	var stream = &iox.OctetsStream{}
	var reader = iox.NewOctetsReader(stream)

	for !my.wc.IsClosed() {
		_ = my.conn.SetReadDeadline(time.Now().Add(my.heartbeatInterval * 3))
		var num, err1 = my.conn.Read(buffer)
		if err1 != nil {
			my.onReadHandler(nil, err1)
			return
		}

		_ = stream.Write(buffer[:num])
		my.onReadHandler(reader, nil)
		stream.Tidy()
	}
}

func (my *Client) onReadHandler(reader *iox.OctetsReader, err error) {
	if err != nil {
		logo.Info("close session(%d) by err=%q", my.id, err)
		_ = my.Close()
		return
	}

	if err1 := my.onReceivedData(reader); err1 != nil {
		logo.Info("close session(%d) by onReceivedData(), err=%q", my.id, err1)
		_ = my.Close()
		return
	}
}

func (my *Client) onReceivedData(reader *iox.OctetsReader) error {
	var packets, err1 = serde.DecodePacket(reader)
	if err1 != nil {
		var err2 = fmt.Errorf("failed to decode message: %s", err1.Error())
		return err2
	}

	for _, pack := range packets {
		var err3 = my.onReceivedPacket(pack)
		if err3 != nil {
			return err3
		}
	}

	return nil
}

func (my *Client) onReceivedPacket(pack serde.Packet) error {
	switch pack.Kind {
	case serde.Handshake:
		return my.onReceivedHandshake(pack)
	case serde.Heartbeat:
		break
	case serde.Kick:
		return my.Close()
	case serde.RouteKind:
		return my.onReceivedRouteKind(pack)
	case serde.Echo:
		return my.onReceivedEcho(pack)
	default:
		return my.onReceivedUserdata(pack)
	}
	return nil
}

func (my *Client) onReceivedHandshake(pack serde.Packet) error {
	var handshake serde.JsonHandshake
	var err = convert.FromJsonE(pack.Data, &handshake)
	if err != nil {
		return err
	}

	logo.JsonI("handshake", handshake)
	my.heartbeatInterval = time.Duration(handshake.Heartbeat) * time.Second

	clear(my.routeKinds)
	clear(my.kindRoutes)

	var routes = handshake.Routes
	for i := 0; i < len(routes); i++ {
		var kind = serde.UserBase + int32(i)
		var route = routes[i]
		my.routeKinds[route] = kind
		my.kindRoutes[kind] = route
	}

	my.nonce = handshake.Nonce
	my.handshakeRe()

	if my.onHandShaken != nil {
		my.onHandShaken(&handshake)
	}

	// 启动heartbeat
	go my.goHeartbeat()
	return nil
}

func (my *Client) handshakeRe() {
	var reply = serde.JsonHandshakeRe{
		Serde: "json",
	}

	var replyData = convert.ToJson(reply)
	var pack = serde.Packet{
		Kind: serde.HandshakeRe,
		Data: replyData,
	}

	_ = my.sendPacket(pack)
}

func (my *Client) onReceivedRouteKind(pack serde.Packet) error {
	var bean serde.JsonRouteKind
	if err := convert.FromJsonE(pack.Data, &bean); err != nil {
		return err
	}

	my.routeKinds[bean.Route] = bean.Kind
	my.kindRoutes[bean.Kind] = bean.Route
	return nil
}

func (my *Client) onReceivedEcho(pack serde.Packet) error {
	return my.sendPacket(pack)
}

func (my *Client) onReceivedUserdata(pack serde.Packet) error {
	var kind = pack.Kind
	if kind < serde.UserBase {
		var err = road.NewError("ErrInvalidKind", "kind=%v", kind)
		return err
	}

	var handler = my.fetchHandler(pack)
	if handler == nil {
		// 有些协议, 真不想处理, 就不设置handlers了. 通常只要有requestId, 就是故意不处理的
		if pack.RequestId == 0 {
			logo.Warn("no handler, kind=%d, requestId=0", kind)
		}

		return nil
	}

	var hasError = len(pack.Code) > 0
	if hasError {
		var code = convert.String(pack.Code)
		var message = convert.String(pack.Data)
		var err = road.NewError(code, message)

		handler(nil, err)
	} else {
		handler(pack.Data, nil)
	}

	return nil
}

func (my *Client) fetchHandler(pack serde.Packet) func([]byte, *road.Error) {
	var requestId = pack.RequestId
	if requestId != 0 {
		if handler, ok := my.requestHandlers[requestId]; ok {
			delete(my.requestHandlers, requestId)
			return handler
		}
	} else {
		var route = my.kindRoutes[pack.Kind]
		if handler, ok := my.registeredHandlers[route]; ok {
			return handler
		}
	}

	return nil
}

func (my *Client) Send(route string, v any) error {
	if my.wc.IsClosed() {
		return nil
	}

	var data, err1 = my.serde.Serialize(v)
	if err1 != nil {
		return err1
	}

	var kind, ok = my.routeKinds[route]
	var pack = serde.Packet{Kind: kind, Data: data}
	if !ok {
		return road.ErrInvalidRoute
	}

	var err3 = my.sendPacket(pack)
	return err3
}

func (my *Client) Request(route string, request any, pResponse any, handler func(*road.Error)) error {
	if my.serde == nil {
		return road.ErrNilSerde
	}

	if route == "" || request == nil || pResponse == nil {
		return road.ErrInvalidArgument
	}

	var data, err = my.serde.Serialize(request)
	if err != nil {
		return err
	}

	var kind, ok = my.routeKinds[route]
	if !ok {
		return road.ErrInvalidRoute
	}

	my.requestIdGenerator++
	var requestId = my.requestIdGenerator
	var pack = serde.Packet{
		Kind:      kind,
		RequestId: requestId,
		Data:      data,
	}

	if handler != nil {
		my.requestHandlers[requestId] = func(data1 []byte, err *road.Error) {
			if data1 != nil {
				var err2 = my.serde.Deserialize(data1, pResponse)
				var err3 *road.Error
				if err2 != nil {
					err3 = road.NewError("ErrDeserialize", "err2=%q", err2)
				}

				handler(err3)
			} else {
				handler(err)
			}
		}
	}

	return my.sendPacket(pack)
}

func (my *Client) On(route string, pResponse any, handler func(*road.Error)) error {
	if route == "" {
		return road.ErrInvalidRoute
	}

	if handler == nil {
		return road.ErrEmptyHandler
	}

	my.registeredHandlers[route] = func(data1 []byte, err *road.Error) {
		if data1 != nil {
			var err2 = my.serde.Deserialize(data1, pResponse)
			var err3 *road.Error
			if err2 != nil {
				err3 = road.NewError("ErrDeserialize", "err2=%q", err2)
			}

			handler(err3)
		} else {
			handler(err)
		}
	}

	return nil
}

func (my *Client) Nonce() int32 {
	return my.nonce
}

func (my *Client) sendPacket(pack serde.Packet) error {
	my.writeLock.Lock()
	defer my.writeLock.Unlock()

	var writer = my.writer
	var stream = writer.Stream()
	stream.Reset()
	serde.EncodePacket(writer, pack)

	var buffer = stream.Bytes()
	var _, err = my.conn.Write(buffer)
	return err
}
