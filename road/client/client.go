package client

import (
	"compress/flate"
	"crypto/tls"
	"encoding/base64"
	"fmt"
	"io"
	"net"
	"strings"
	"sync"
	"time"

	"github.com/lixianmin/gonsole/road"
	"github.com/lixianmin/gonsole/road/serde"
	"github.com/lixianmin/got/convert"
	"github.com/lixianmin/got/iox"
	"github.com/lixianmin/got/loom"
	"github.com/lixianmin/logo"
)

/********************************************************************
created:    2023-11-26
author:     lixianmin

Copyright (C) - All Rights Reserved
*********************************************************************/

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

	// routeKinds/kindRoutes/requestHandlers/registeredHandlers 会被两个goroutine并发访问：
	//  - goLoop goroutine：onReceivedHandshake/onReceivedRouteKind 写入；fetchHandler 读取+删除
	//  - 用户 goroutine：Request/On/Send 写入与读取
	// 因此必须用 mapLock 保护，否则是data race
	mapLock            sync.RWMutex
	routeKinds         map[string]int32
	kindRoutes         map[int32]string
	requestHandlers    map[int32]func([]byte, *road.Error)
	registeredHandlers map[string]func([]byte, *road.Error)
}

func NewClient() *Client {
	var my = &Client{
		// 仅用于日志标识（如 close session(%d)），不需要全局唯一。
		// 不用包级原子计数器：宪法4.2禁止全局可变状态传递运行时状态（计数器）。
		id:                 time.Now().UnixNano(),
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
	default:
		return my.onReceivedUserdata(pack)
	}
	return nil
}

// inflateRaw 解压 deflate-raw 格式的 base64 编码数据
func inflateRaw(base64CompressedData string) (string, error) {
	// base64 解码
	compressedData, err := base64.StdEncoding.DecodeString(base64CompressedData)
	if err != nil {
		return "", fmt.Errorf("failed to decode base64: %w", err)
	}

	// 创建 deflate-raw 解压缩读取器
	reader := flate.NewReader(strings.NewReader(string(compressedData)))
	defer reader.Close()

	// 读取解压后的数据
	decompressedData, err := io.ReadAll(reader)
	if err != nil {
		return "", fmt.Errorf("failed to decompress deflate-raw: %w", err)
	}

	return string(decompressedData), nil
}

func (my *Client) onReceivedHandshake(pack serde.Packet) error {
	var handshake serde.JsonHandshake
	var err = convert.FromJsonE(pack.Data, &handshake)
	if err != nil {
		return err
	}

	logo.JsonI("handshake", handshake)
	my.heartbeatInterval = time.Duration(handshake.Heartbeat) * time.Second

	my.mapLock.Lock()
	clear(my.routeKinds)
	clear(my.kindRoutes)

	// 解压并解析 routes
	var routes []string
	if handshake.Routes != "" {
		// 先尝试新格式：base64 + deflate-raw 压缩
		if decompressedRoutes, err := inflateRaw(handshake.Routes); err == nil {
			// 用空格分割 routes
			routes = strings.Split(decompressedRoutes, " ")
		} else {
			// 如果解压失败，可能是旧格式或其他问题
			logo.Warn("failed to decompress routes, error: %v", err)
			my.mapLock.Unlock()
			return fmt.Errorf("failed to process routes: %w", err)
		}
	}

	// 构建 route-kind 映射
	for i := 0; i < len(routes); i++ {
		var kind = serde.UserBase + int32(i)
		var route = routes[i]
		my.routeKinds[route] = kind
		my.kindRoutes[kind] = route
	}
	my.nonce = handshake.Nonce
	my.mapLock.Unlock()

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

	my.mapLock.Lock()
	my.routeKinds[bean.Route] = bean.Kind
	my.kindRoutes[bean.Kind] = bean.Route
	my.mapLock.Unlock()
	return nil
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
		my.mapLock.Lock()
		var handler, ok = my.requestHandlers[requestId]
		if ok {
			delete(my.requestHandlers, requestId)
		}
		my.mapLock.Unlock()
		if ok {
			return handler
		}
	} else {
		my.mapLock.RLock()
		var route = my.kindRoutes[pack.Kind]
		var handler, ok = my.registeredHandlers[route]
		my.mapLock.RUnlock()
		if ok {
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

	my.mapLock.RLock()
	var kind, ok = my.routeKinds[route]
	my.mapLock.RUnlock()
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

	my.mapLock.RLock()
	var kind, ok = my.routeKinds[route]
	my.mapLock.RUnlock()
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
		var wrapped = func(data1 []byte, err *road.Error) {
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

		my.mapLock.Lock()
		my.requestHandlers[requestId] = wrapped
		my.mapLock.Unlock()
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

	my.mapLock.Lock()
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
	my.mapLock.Unlock()

	return nil
}

func (my *Client) Nonce() int32 {
	my.mapLock.RLock()
	defer my.mapLock.RUnlock()

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
