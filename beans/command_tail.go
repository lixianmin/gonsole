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

func readTail(note string, texts []string, maxTailNum int) string {
	if len(texts) == 1 {
		return note
	}

	var fullPath = ""
	var num = 20
	var err error

	if len(texts) == 2 {
		fullPath = texts[1]
	} else if len(texts) == 4 && texts[1] == "-n" {
		num, err = strconv.Atoi(texts[2])
		if err != nil {
			return note
		}

		num = mathx.MinInt(num, maxTailNum)
		fullPath = texts[3]
	} else {
		return note
	}

	var lines = tools.ReadFileTail(fullPath, num)
	var message = fmt.Sprintf("<br/> 返回行数：%d <br/>", len(lines)) + strings.Join(lines, "<br/>")
	return message
}
