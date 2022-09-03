package client

import (
	"github.com/gobwas/ws"
	"github.com/gobwas/ws/wsutil"
	"net"
	"time"
)

/********************************************************************
created:    2020-09-07
author:     lixianmin

Copyright (C) - All Rights Reserved
*********************************************************************/

type WsClientConn struct {
	conn   net.Conn
	reader *wsutil.Reader
}

func newWsClientConn(conn net.Conn) *WsClientConn {
	var item = &WsClientConn{
		conn:   conn,
		reader: wsutil.NewReader(conn, ws.StateClientSide),
	}

	return item
}

func (my *WsClientConn) Read(b []byte) (int, error) {
	var reader = my.reader
	var _, err = reader.NextFrame()
	if err != nil {
		return 0, err
	}

	return reader.Read(b)
}

func (my *WsClientConn) Write(b []byte) (int, error) {
	var err = wsutil.WriteClientBinary(my.conn, b)
	if err != nil {
		return 0, err
	}

	return len(b), nil
}

func (my *WsClientConn) Close() error {
	return my.conn.Close()
}

// LocalAddr returns the local address.
func (my *WsClientConn) LocalAddr() net.Addr {
	return my.conn.LocalAddr()
}

// RemoteAddr returns the remote address.
func (my *WsClientConn) RemoteAddr() net.Addr {
	return my.conn.RemoteAddr()
}

// SetDeadline sets the read and write deadlines associated
// with the connection. It is equivalent to calling both
// SetReadDeadline and SetWriteDeadline.
//
// A deadline is an absolute time after which I/O operations
// fail with a timeout (see type Error) instead of
// blocking. The deadline applies to all future and pending
// I/O, not just the immediately following call to Read or
// Write. After a deadline has been exceeded, the connection
// can be refreshed by setting a deadline in the future.
//
// An idle timeout can be implemented by repeatedly extending
// the deadline after successful Read or Write calls.
//
// A zero value for t means I/O operations will not time out.
func (my *WsClientConn) SetDeadline(t time.Time) error {
	if err := my.SetReadDeadline(t); err != nil {
		return err
	}

	return my.SetWriteDeadline(t)
}

// SetReadDeadline sets the deadline for future Read calls
// and any currently-blocked Read call.
// A zero value for t means Read will not time out.
func (my *WsClientConn) SetReadDeadline(t time.Time) error {
	return my.conn.SetReadDeadline(t)
}

// SetWriteDeadline sets the deadline for future Write calls
// and any currently-blocked Write call.
// Even if write times out, it may return n > 0, indicating that
// some data was successfully written.
// A zero value for t means Write will not time out.
func (my *WsClientConn) SetWriteDeadline(t time.Time) error {
	return my.conn.SetWriteDeadline(t)
}
