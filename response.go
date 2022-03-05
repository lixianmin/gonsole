package gonsole

import "github.com/lixianmin/got/convert"

/********************************************************************
created:    2020-09-02
author:     lixianmin

Copyright (C) - All Rights Reserved
*********************************************************************/

type Response struct {
	Operation string      `json:"op"`
	Data      interface{} `json:"data"`
}

func NewEmptyResponse() *Response {
	var ret = &Response{Operation: "empty"}
	return ret
}

func NewDefaultResponse(data interface{}) *Response {
	var ret = &Response{Operation: "default", Data: data}
	return ret
}

func NewHtmlResponse(data string) *Response {
	var ret = &Response{Operation: "html", Data: data}
	return ret
}

func NewTableResponse(table interface{}) *Response {
	var data = convert.String(convert.ToJson(table))
	var ret = &Response{Operation: "table", Data: data}
	return ret
}
