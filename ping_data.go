package gonsole

/********************************************************************
created:    2019-11-16
author:     lixianmin

Copyright (C) - All Rights Reserved
*********************************************************************/

type PingData struct {
	BasicRequest
}

type PingDataRe struct {
	BasicResponse
}

func newPingDataRe(requestID string) *PingDataRe {
	var res = &PingDataRe{
	}

	res.BasicResponse = *newBasicResponse("pong", requestID)
	return res
}
