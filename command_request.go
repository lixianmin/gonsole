package gonsole

/********************************************************************
created:    2019-11-16
author:     lixianmin

Copyright (C) - All Rights Reserved
*********************************************************************/

type CommandRequest struct {
	BasicRequest
	Command string `json:"command"`
}
