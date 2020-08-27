package beans

import (
	"flag"
	"fmt"
	"github.com/lixianmin/gonsole/tools"
	"github.com/lixianmin/got/mathx"
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

	var fullPath, num, filter, err = parseReadFileArgs(texts, maxNum)
	if err != nil {
		return note
	}

	var lines = tools.ReadTailLines(fullPath, num, filter)
	var message = fmt.Sprintf("<br> 返回行数：%d <br>", len(lines)) + strings.Join(lines, "<br>")
	return message
}

func parseReadFileArgs(texts [] string, maxNum int) (fullPath string, num int, filter string, err error) {
	var fs = flag.NewFlagSet("fs", flag.ContinueOnError)
	fs.IntVar(&num, "n", 50, "返回n行")
	fs.StringVar(&filter, "f", "", "按f过滤")

	err = fs.Parse(texts[1:])
	if err != nil || fs.NArg() == 0 {
		return
	}

	fullPath = fs.Arg(0)
	num = mathx.MinInt(num, maxNum)
	return
}
