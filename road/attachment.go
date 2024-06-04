package road

/********************************************************************
created:    2023-06-12
author:     lixianmin

Copyright (C) - All Rights Reserved
*********************************************************************/

type Attachment interface {
	Put(key any, value any)
	UInt32(key any) uint32
	Int32(key any) int32
	UInt64(key any) uint64
	Int64(key any) int64
	Int(key any) int
	Float32(key any) float32
	Float64(key any) float64
	Bool(key any) bool
	String(key any) string
	Get1(key any) any
	Get2(key any) (any, bool)
}
