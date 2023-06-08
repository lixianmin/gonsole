package epoll

import (
	"github.com/lixianmin/gonsole/road/network"
)

/********************************************************************
created:    2020-10-05
author:     lixianmin

Copyright (C) - All Rights Reserved
*********************************************************************/

type Acceptor interface {
	GetLinkChan() chan network.Link
}
