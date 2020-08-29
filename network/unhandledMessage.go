package network

import (
	"context"
	"github.com/lixianmin/gonsole/network/conn/message"
	"github.com/lixianmin/gonsole/network/route"
)

/********************************************************************
created:    2020-08-29
author:     lixianmin

Copyright (C) - All Rights Reserved
*********************************************************************/

type unhandledMessage struct {
	ctx   context.Context
	route *route.Route
	msg   *message.Message
}
