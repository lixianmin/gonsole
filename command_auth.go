package gonsole

import (
	"github.com/lixianmin/gonsole/tools"
)

/********************************************************************
created:    2020-07-03
author:     lixianmin

Copyright (C) - All Rights Reserved
*********************************************************************/

type CommandAuth struct {
	BasicResponse
	Text string `json:"text"`
}

func newCommandAuth(client *Client, texts []string, userPasswords map[string]string) *CommandAuth {
	var bean = &CommandAuth{}
	bean.Operation = "auth"
	bean.Timestamp = tools.GetTimestamp()

	if len(texts) < 3 {
		bean.Text = "格式：auth username"
		return bean
	}

	var username = texts[1]
	var password = texts[2]
	if password1, ok := userPasswords[username]; !ok || password1 != password {
		bean.Text = "用户名或密码错误"
		client.isAuthorized = false
		return bean
	}

	bean.Text = "验证成功"
	client.isAuthorized = true
	return bean
}
