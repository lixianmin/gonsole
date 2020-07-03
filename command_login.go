package gonsole

import (
	"github.com/lixianmin/gonsole/tools"
)

/********************************************************************
created:    2020-07-03
author:     lixianmin

Copyright (C) - All Rights Reserved
*********************************************************************/

type CommandLogin struct {
	BasicResponse
	Text string `json:"text"`
}

func newCommandLogin(client *Client, texts []string, userPasswords map[string]string) *CommandLogin {
	var bean = &CommandLogin{}
	bean.Operation = "login"
	bean.Timestamp = tools.GetTimestamp()

	if len(texts) < 3 {
		bean.Text = "格式：login username password"
		return bean
	}

	var username = texts[1]
	var password = texts[2]
	if password1, ok := userPasswords[username]; !ok || password1 != password {
		bean.Text = "用户名或密码错误"
		client.isLogin = false
		return bean
	}

	bean.Text = "登陆成功"
	client.isLogin = true
	return bean
}
