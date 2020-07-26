package beans

import (
	"fmt"
	"github.com/lixianmin/gonsole/tools"
	"github.com/lixianmin/got/mathx"
	"strconv"
	"strings"
)

/********************************************************************
created:    2020-07-23
author:     lixianmin

Copyright (C) - All Rights Reserved
*********************************************************************/

func ReadFileTail(note string, texts []string, maxNum int) string {
	if len(texts) == 1 {
		return note
	}

	var fullPath, num, err = parseReadFileArgs(texts, maxNum)
	if err != nil {
		return note
	}

	var lines = tools.ReadTailLines(fullPath, num)
	var message = fmt.Sprintf("<br> 返回行数：%d <br>", len(lines)) + strings.Join(lines, "<br>")
	return message
}

func parseReadFileArgs(texts [] string, maxNum int) (fullPath string, num int, err error) {
	fullPath = ""
	num = 20

	if len(texts) == 2 {
		fullPath = texts[1]
	} else if len(texts) == 4 && texts[1] == "-n" {
		num, err = strconv.Atoi(texts[2])
		if err != nil {
			return
		}

		num = mathx.MinInt(num, maxNum)
		fullPath = texts[3]
	}

	return
}
