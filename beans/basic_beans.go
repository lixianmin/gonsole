package gonsole

import "github.com/lixianmin/gonsole/tools"

/********************************************************************
created:    2020-06-02
author:     lixianmin

Copyright (C) - All Rights Reserved
*********************************************************************/

/////////////////////////////////////////////////////////////////////////////////
type BasicRequest struct {
	Operation string `json:"op"`  // 协议类型，服务器使用op创建对应的bean对象
	RequestId string `json:"rid"` // 当服务器应答client的请求时，可以带上此id，用于客户端查错
}

/////////////////////////////////////////////////////////////////////////////////
type BasicResponse struct {
	Operation string `json:"op"`  // 协议类型
	RequestID string `json:"rid"` // 当服务器应答client的请求时，可以带上此id，用于客户端查错
	Timestamp int64  `json:"ts"`  // 服务器的时间戳
}

func NewBasicResponse(operation string, requestID string) *BasicResponse {
	var response = &BasicResponse{
		Operation: operation,
		RequestID: requestID,
		Timestamp: tools.GetTimestamp(),
	}

	return response
}
