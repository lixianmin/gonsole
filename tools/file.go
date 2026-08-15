package tools

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
)

/********************************************************************
created:    2020-07-23
author:     lixianmin

Copyright (C) - All Rights Reserved
*********************************************************************/

func ReadTailLines(fullPath string, num int, filter string) ([]string, error) {
	if num <= 0 {
		return nil, fmt.Errorf("invalid num: %d", num)
	}

	var fin, err = os.Open(fullPath)
	if err != nil {
		return nil, fmt.Errorf("open file %s: %w", fullPath, err)
	}

	defer fin.Close()

	var reader = bufio.NewReader(fin)
	var lines = make([]string, 0, num)
	var cache = make([]string, num)

	filter = strings.ToLower(filter)
	var nextIndex = 0
	var lineNum = 0
	var resultCount = 0

	// 把一行写入环形缓存；返回true表示文件已读完
	var appendLine = func(line string) bool {
		lineNum++
		if filter == "" || strings.Contains(strings.ToLower(line), filter) {
			cache[nextIndex] = strconv.Itoa(lineNum) + " " + line
			nextIndex = (nextIndex + 1) % num
			resultCount++
		}
		return false
	}

	for {
		var line, err = reader.ReadString('\n')
		if err != nil {
			// bugfix: ReadString在读到文件末尾且最后一行没有换行符时，会返回(部分数据, io.EOF)，
			// 原实现直接返回导致最后一行被丢弃。这里先处理部分数据再退出。
			if len(line) > 0 {
				appendLine(line)
			}

			// resultCount不足num时，不应该出现空白行
			if resultCount >= num {
				lines = append(lines, cache[nextIndex:]...)
			}

			lines = append(lines, cache[:nextIndex]...)
			return lines, nil
		}

		appendLine(line)
	}
}
