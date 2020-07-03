package gonsole

/********************************************************************
created:    2020-06-04
author:     lixianmin

Copyright (C) - All Rights Reserved
*********************************************************************/

type Command struct {
	Name     string                                // 名称
	Note     string                                // 描述
	IsPublic bool                                  // 非public方法需要登陆
	Handler  func(client *Client, texts [] string) // 处理方法
}
