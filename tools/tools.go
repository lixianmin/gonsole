package tools

import (
	"os"
	"time"
)

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

func IsPathExist(path string) bool {
	_, err := os.Stat(path)
	if err != nil && os.IsNotExist(err) {
		return false
	}

	return true
}
