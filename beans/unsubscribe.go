package gonsole

import "github.com/lixianmin/gonsole/beans"

/********************************************************************
created:    2019-11-16
author:     lixianmin

Copyright (C) - All Rights Reserved
*********************************************************************/

type Unsubscribe struct {
	beans.BasicRequest
	TopicId string `json:"topic"`
}

type UnsubscribeRe struct {
	beans.BasicResponse
	TopicId string `json:"topic"`
}

func newUnsubscribeRe(requestId string, topicId string) *UnsubscribeRe {
	var res = &UnsubscribeRe{
		TopicId: topicId,
	}

	res.BasicResponse = *beans.NewBasicResponse("unsubscribeRe", requestId)
	return res
}
