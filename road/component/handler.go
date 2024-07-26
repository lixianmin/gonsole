package component

import (
	"reflect"
)

/********************************************************************
created:    2023-06-05
author:     lixianmin

Copyright (C) - All Rights Reserved
*********************************************************************/

// Handler represents a message.Message's handler's meta information.
type Handler struct {
	Receiver reflect.Value  // receiver of method
	Method   reflect.Method // method stub
	Type     reflect.Type   // low-level type of method
	IsRawArg bool           // 参数是否是[]byte
	Route    string
}
