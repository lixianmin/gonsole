package beans

import (
	"github.com/lixianmin/gonsole/ifs"
	"github.com/lixianmin/gonsole/tools"
	"github.com/lixianmin/road"
)

/********************************************************************
created:    2020-07-03
author:     lixianmin

Copyright (C) - All Rights Reserved
*********************************************************************/

type CommandAuth struct {
	GPID          string `json:"gpid"`
	ClientAddress string `json:"client"`
	Text          string `json:"text"`
}

func NewCommandAuth(session *road.Session, args []string, userPasswords map[string]string, port int) *CommandAuth {
	var bean = &CommandAuth{}

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

	bean.GPID = tools.GetGPID(port)
	bean.ClientAddress = session.RemoteAddr().String()
	bean.Text = "验证成功"
	session.Attachment().Put(ifs.KeyIsAuthorized, true)
	return bean
}
