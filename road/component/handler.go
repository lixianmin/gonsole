package component

import (
	"github.com/lixianmin/gonsole/road/message"
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
	Type        reflect.Type   // low-level type of method
	IsRawArg    bool           // whether the data need to serialize
	MessageType message.Kind   // handler allowed message type (either request or notify)
}
