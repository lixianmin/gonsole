package road

import "fmt"

/********************************************************************
created:    2020-09-02
author:     lixianmin

Copyright (C) - All Rights Reserved
*********************************************************************/

// var ErrKickedByRateLimit = NewError("KickedByRateLimit", "cost too many tokens in a rate limit window")
var ErrPacketProcessed = NewError("PacketProcessed", "packet has been processed, no need to process any more")
var ErrEmptyHandler = NewError("EmptyHandler", "handler is empty")

type Error struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

func NewError(code string, format string, args ...interface{}) *Error {
	var message = format
	if len(args) > 0 {
		message = fmt.Sprintf(format, args...)
	}

	var err = &Error{
		Code:    code,
		Message: message,
	}

	return err
}

func (err *Error) Error() string {
	return fmt.Sprintf("code=%q message=%q", err.Code, err.Message)
}
