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
	Receiver    reflect.Value  // receiver of method
	Method      reflect.Method // method stub
	RequestType reflect.Type   // request参数的类型
	NumIn       int8
	Route       string
}
