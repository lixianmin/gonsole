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

type ReadFileArgs struct {
	FullPath  string
	Num       int
	Filter    string
	StartLine int
}

func ReadFileTail(note string, texts []string, maxNum int) string {
	if len(texts) == 1 {
		return note
	}

	var args, err = parseReadFileArgs(texts, maxNum)
	if err != nil {
		return note
	}

	var lines = tools.ReadTailLines(args.FullPath, args.Num, args.Filter)
	var message = fmt.Sprintf("<br> 返回行数：%d <br>", len(lines)) + strings.Join(lines, "<br>")
	return message
}

func parseReadFileArgs(texts [] string, maxNum int) (args ReadFileArgs, err error) {
	var fs = flag.NewFlagSet("fs", flag.ContinueOnError)
	fs.IntVar(&args.Num, "n", 50, "返回n行")
	fs.StringVar(&args.Filter, "f", "", "按f过滤")
	fs.IntVar(&args.StartLine, "s", 1, "搜索的起始index")

	err = fs.Parse(texts[1:])
	if err != nil || fs.NArg() == 0 {
		return
	}

	args.FullPath = fs.Arg(0)
	args.Num = mathx.MinInt(args.Num, maxNum)
	args.StartLine = mathx.MaxInt(1, args.StartLine)
	return
}
