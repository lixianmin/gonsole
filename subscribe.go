package gonsole

/********************************************************************
created:    2019-11-16
author:     lixianmin

Copyright (C) - All Rights Reserved
*********************************************************************/

type Subscribe struct {
	BasicRequest
	TopicId string `json:"topic"`
}

type SubscribeRe struct {
	BasicResponse
	TopicId string `json:"topic"`
}

func NewSubscribeRe(requestId string, topicId string) *SubscribeRe {
	var res = &SubscribeRe{
		TopicId: topicId,
	}

	res.BasicResponse = *newBasicResponse("subscribeRe", requestId)
	return res
}
