package road

import (
	"github.com/lixianmin/got/iox"
	"net"
	"time"
)

/********************************************************************
created:    2020-09-06
author:     lixianmin

Copyright (C) - All Rights Reserved
*********************************************************************/

type OnReadHandler func(reader *iox.OctetsReader, err error)

type Link interface {
	GoLoop(heartbeatInterval time.Duration, onReadHandler OnReadHandler)
	Write(data []byte) (int, error)
	Close() error
	RemoteAddr() net.Addr
}
