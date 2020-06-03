package gonsole

import "time"

/********************************************************************
created:    2020-06-02
author:     lixianmin

Copyright (C) - All Rights Reserved
*********************************************************************/

func GetTimestamp() int64 {
	var nanos = time.Now().UnixNano()
	var millis = nanos / 1000000
	return millis
}
