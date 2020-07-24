package gonsole

import "github.com/lixianmin/gonsole/beans"

/********************************************************************
created:    2020-07-20
author:     lixianmin

Copyright (C) - All Rights Reserved
*********************************************************************/

type Ping struct {
	beans.BasicRequest
}

type Pong struct {
	beans.BasicResponse
}
