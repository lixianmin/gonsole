package epoll

import (
	"github.com/lixianmin/gonsole/road/network"
	"time"
)

/********************************************************************
created:    2020-10-05
author:     lixianmin

Copyright (C) - All Rights Reserved
*********************************************************************/

type Acceptor interface {
	GetConnChan() chan network.Connection
	GetHeartbeatInterval() time.Duration
}
