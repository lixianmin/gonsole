package gonsole

/********************************************************************
created:    2020-06-02
author:     lixianmin

Copyright (C) - All Rights Reserved
*********************************************************************/

type BadRequestRe struct {
	BasicResponse
	Code    int    `json:"code"`
	Message string `json:"message"`
}

func newBadRequestRe(requestID string, code int, message string) *BadRequestRe {
	var bean = &BadRequestRe{
		Code:    code,
		Message: message,
	}

	bean.BasicResponse = *newBasicResponse("badRequestRe", requestID)
	return bean
}
