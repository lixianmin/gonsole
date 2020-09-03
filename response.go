package gonsole

/********************************************************************
created:    2020-09-02
author:     lixianmin

Copyright (C) - All Rights Reserved
*********************************************************************/

type Response struct {
	Operation string      `json:"op"`
	Data      interface{} `json:"data"`
}

func NewDefaultResponse(data interface{}) *Response {
	var ret = &Response{Operation: "default", Data: data}
	return ret
}

func NewHtmlResponse(data string) *Response {
	var ret = &Response{Operation: "html", Data: data}
	return ret
}
