package gonsole

import (
	"time"
)

/********************************************************************
created:    2019-11-16
author:     lixianmin

Copyright (C) - All Rights Reserved
*********************************************************************/

// 进程启动的时间
var startProcessTime = time.Now()

type Challenge struct {
	Operation           string `json:"op"` // 协议类型
	Timestamp           int64  `json:"ts"` // 服务器的时间戳
	GPID                string `json:"gpid"`
	ClientRemoteAddress string `json:"client"`
	UpTime              string `json:"uptime"`
}

func newChallenge(gpid string, clientRemoteAddress string) *Challenge {
	var bean = &Challenge{
		Operation:           "challenge",
		Timestamp:           GetTimestamp(),
		GPID:                gpid,
		ClientRemoteAddress: clientRemoteAddress,
		UpTime:              time.Now().Sub(startProcessTime).String(),
	}

	return bean
}
