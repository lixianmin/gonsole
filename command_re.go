package gonsole

/********************************************************************
created:    2020-09-02
author:     lixianmin

Copyright (C) - All Rights Reserved
*********************************************************************/

type CommandRe struct {
	Operation string      `json:"op"`
	Data      interface{} `json:"data"`
}
