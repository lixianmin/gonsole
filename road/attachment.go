package road

/********************************************************************
created:    2023-06-12
author:     lixianmin

Copyright (C) - All Rights Reserved
*********************************************************************/

type Attachment interface {
	Put(key interface{}, value interface{})
	UInt32(key interface{}) uint32
	Int32(key interface{}) int32
	UInt64(key interface{}) uint64
	Int64(key interface{}) int64
	Int(key interface{}) int
	Float32(key interface{}) float32
	Float64(key interface{}) float64
	Bool(key interface{}) bool
	String(key interface{}) string
	Get1(key interface{}) interface{}
	Get2(key interface{}) (interface{}, bool)
}
