package gonsole

import "github.com/lixianmin/gonsole/beans"

/********************************************************************
created:    2019-11-16
author:     lixianmin

Copyright (C) - All Rights Reserved
*********************************************************************/

type CommandRequest struct {
	beans.BasicRequest
	Command string `json:"command"`
}
