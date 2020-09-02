package beans

import (
	"github.com/lixianmin/bugfly"
	"github.com/lixianmin/gonsole/ifs"
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

func NewCommandAuth(session *bugfly.Session, args []string, userPasswords map[string]string) *CommandAuth {
	var bean = &CommandAuth{}
	bean.Operation = "auth"
	bean.Timestamp = tools.GetTimestamp()

	if len(args) < 3 {
		bean.Text = "格式：auth username"
		return bean
	}

	var username = args[1]
	var password = args[2]
	if password1, ok := userPasswords[username]; !ok || password1 != password {
		bean.Text = "用户名或密码错误"
		session.Attachment().Put(ifs.KeyIsAuthorized, false)
		return bean
	}

	bean.Text = "验证成功"
	session.Attachment().Put(ifs.KeyIsAuthorized, true)
	return bean
}
