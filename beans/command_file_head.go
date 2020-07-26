package beans

import (
	"bufio"
	"fmt"
	"os"
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

	var fullPath, num, err = parseReadFileArgs(texts, maxNum)
	if err != nil {
		return note
	}

	var lines = readHeadLines(fullPath, num)
	var message = fmt.Sprintf("<br> 返回行数：%d <br>", len(lines)) + strings.Join(lines, "<br>")
	return message
}

func readHeadLines(fullPath string, num int) []string {
	var fin, err = os.Open(fullPath)
	if err != nil {
		return nil
	}

	defer fin.Close()

	var reader = bufio.NewReader(fin)
	var lines = make([]string, 0, num)
	for i := 0; i < num; i++ {
		var line, err = reader.ReadString('\n')
		if err != nil {
			break
		}

		lines = append(lines, line)
	}

	return lines
}
