package gonsole

/********************************************************************
created:    2020-06-04
author:     lixianmin

Copyright (C) - All Rights Reserved
*********************************************************************/

type Command struct {
	Name    string
	Remark  string
	Handler func(client *Client)
}