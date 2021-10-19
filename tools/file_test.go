package tools

import (
	"fmt"
	"sort"
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
	var lines = ReadTailLines(fullPath, 10, "server")
	fmt.Println(strings.Join(lines, "\n"))
}

func TestTimeSort (t *testing.T) {
	var timeList = []string {"2022-10-19\t", "2021-10-18\t", "2021-10-19\t", "2021-10-20\t"}
	sort.Slice(timeList, func(i, j int) bool {
		return timeList[i] < timeList[j]
	})
	fmt.Println(timeList)

	sort.Slice(timeList, func(i, j int) bool {
		return timeList[i] > timeList[j]
	})
	fmt.Println(timeList)
}
