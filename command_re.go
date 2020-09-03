package gonsole

/********************************************************************
created:    2020-09-02
author:     lixianmin

Copyright (C) - All Rights Reserved
*********************************************************************/

type CommandRe struct {
	Operation string      `json:"op"`
	Data      interface{} `json:"data"`
}

func NewDefaultCommandRe(data interface{}) *CommandRe {
	var ret = &CommandRe{Operation: "default", Data: data}
	return ret
}

func NewHtmlCommandRe(data string) *CommandRe {
	var ret = &CommandRe{Operation: "html", Data: data}
	return ret
}
