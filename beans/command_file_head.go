package beans

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
)

/********************************************************************
created:    2020-07-26
author:     lixianmin

Copyright (C) - All Rights Reserved
*********************************************************************/

func ReadFileHead(note string, texts []string, maxNum int) string {
	if len(texts) == 1 {
		return note
	}

	var args, err = parseReadFileArgs(texts, maxNum)
	if err != nil {
		return note
	}

	var lines = readHeadLines(args)
	var message = fmt.Sprintf("<br> 返回行数：%d <br>", len(lines)) + strings.Join(lines, "<br>")
	return message
}

func readHeadLines(args ReadFileArgs) []string {
	var fin, err = os.Open(args.FullPath)
	if err != nil {
		return nil
	}

	defer fin.Close()

	var reader = bufio.NewReader(fin)
	var lines = make([]string, 0, args.Num)
	var skipLines = args.StartLine - 1

	for i := 0; i < skipLines; i++ {
		var _, err = reader.ReadString('\n')
		if err != nil {
			return lines
		}
	}

	var filter = strings.ToLower(args.Filter)
	var lineNum = skipLines
	var counter = 0
	for counter < args.Num {
		var line, err = reader.ReadString('\n')
		if err != nil {
			break
		}

		lineNum += 1
		if filter == "" || strings.Contains(strings.ToLower(line), filter) {
			var item = strconv.Itoa(lineNum) + " " + line
			lines = append(lines, item)
			counter += 1
		}
	}

	return lines
}
