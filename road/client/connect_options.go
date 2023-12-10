package client

import (
	"crypto/tls"
	"github.com/lixianmin/gonsole/road/serde"
)

/********************************************************************
created:    2020-08-29
author:     lixianmin

Copyright (C) - All Rights Reserved
*********************************************************************/

type connectOptions struct {
	serde        serde.Serde
	tlsConfig    *tls.Config
	onHandShaken func(bean *serde.JsonHandshake)
}

type ConnectOption func(options *connectOptions)

func WithSerde(serde serde.Serde) ConnectOption {
	return func(opt *connectOptions) {
		opt.serde = serde
	}
}

func WithTlsConfig(config *tls.Config) ConnectOption {
	return func(opt *connectOptions) {
		opt.tlsConfig = config
	}
}

func WithOnHandShaken(handler func(bean *serde.JsonHandshake)) ConnectOption {
	return func(opt *connectOptions) {
		opt.onHandShaken = handler
	}
}
