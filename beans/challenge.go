package beans

import (
	"fmt"
	"github.com/lixianmin/gonsole/tools"
	"github.com/lixianmin/got/timex"
	"time"
)

/********************************************************************
created:    2019-11-16
author:     lixianmin

Copyright (C) - All Rights Reserved
*********************************************************************/

type Challenge struct {
	Operation           string `json:"op"` // 协议类型
	Timestamp           int64  `json:"ts"` // 服务器的时间戳
	GPID                string `json:"gpid"`
	ClientRemoteAddress string `json:"client"`
	UpTime              string `json:"uptime"`
}

func NewChallenge(gpid string, clientRemoteAddress string) *Challenge {
	var uptime = time.Now().Sub(startProcessTime)
	var bean = &Challenge{
		Operation:           "challenge",
		Timestamp:           tools.GetTimestamp(),
		GPID:                gpid,
		ClientRemoteAddress: clientRemoteAddress,
		UpTime:              fmt.Sprintf("%s ( %s )", tools.FormatDuration(uptime), startProcessTime.Format(timex.Layout)),
	}

	return bean
}
