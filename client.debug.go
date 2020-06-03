package gonsole

import (
	"strings"
)

/********************************************************************
created:    2019-11-16
author:     lixianmin

Copyright (C) - All Rights Reserved
*********************************************************************/

func loopClientDebugRequest(client *Client, requestId string, command string) {
	var texts = strings.Split(command, " ")
	var cmd = texts[0]
	switch cmd {
	case "help":
		var remoteAddress = client.GetRemoteAddress()
		client.SendBeanAsync(NewDebugHelp(remoteAddress))
	case "ls":
		client.SendBeanAsync(newDebugListTopics())
	default:
		client.SendBeanAsync(NewBadRequestRe(requestId, InternalError, command))
	}
}
