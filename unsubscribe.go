package gonsole

/********************************************************************
created:    2019-11-16
author:     lixianmin

Copyright (C) - All Rights Reserved
*********************************************************************/

type Unsubscribe struct {
	BasicRequest
	TopicId string `json:"topic"`
}

type UnsubscribeRe struct {
	BasicResponse
	TopicId string `json:"topic"`
}

func NewUnsubscribeRe(requestId string, topicId string) *UnsubscribeRe {
	var res = &UnsubscribeRe{
		TopicId: topicId,
	}

	res.BasicResponse = *NewBasicResponse("unsubscribeRe", requestId)
	return res
}
