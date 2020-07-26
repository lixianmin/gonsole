package tools

import (
	"fmt"
	"strings"
	"testing"
)

/********************************************************************
created:    2020-07-23
author:     lixianmin

Copyright (C) - All Rights Reserved
*********************************************************************/

func TestReadFileTail(t *testing.T) {
	var fullPath = "../logs/log_debug.log"
	var lines = ReadTailLines(fullPath, 10)
	fmt.Println(strings.Join(lines, "\n"))
}
