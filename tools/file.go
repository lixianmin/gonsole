package tools

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
)

/********************************************************************
created:    2020-07-23
author:     lixianmin

Copyright (C) - All Rights Reserved
*********************************************************************/

func ReadTailLines(fullPath string, num int, filter string) []string {
	if num <= 0 {
		return nil
	}

	var fin, err = os.Open(fullPath)
	if err != nil {
		return nil
	}

	defer fin.Close()

	var reader = bufio.NewReader(fin)
	var lines = make([]string, 0, num)
	var cache = make([]string, num)

	filter = strings.ToLower(filter)
	var nextIndex = 0
	var lineNum = 0

	for {
		var line, err = reader.ReadString('\n')
		if err != nil {
			lines = append(lines, cache[nextIndex:]...)
			lines = append(lines, cache[:nextIndex]...)
			return lines
		}

		lineNum += 1
		if filter == "" || strings.Contains(strings.ToLower(line), filter) {
			cache[nextIndex] = strconv.Itoa(lineNum) + " " + line
			nextIndex = (nextIndex + 1) % num
		}
	}
}

func searchOffset(fin *os.File, num int) (int64, error) {
	info, err := fin.Stat()
	if err != nil {
		return 0, err
	}

	var fileSize = info.Size()

	const bufferSize int64 = 1024
	var buffer [bufferSize]byte

	var counter = 0
	for i := 0; true; i++ {
		var stepSize = bufferSize
		var offset = int64(i+1) * bufferSize

		var isLastRead = false
		if offset > fileSize {
			stepSize = fileSize - int64(i)*bufferSize
			offset = fileSize
			isLastRead = true
		}

		_, err := fin.Seek(-offset, io.SeekEnd)
		if err != nil {
			return 0, err
		}

		n, err := fin.Read(buffer[0:stepSize])
		if err != nil {
			return 0, err
		}

		if n != int(stepSize) {
			return 0, fmt.Errorf("n=%d, stepSize=%d", n, stepSize)
		}

		for j := n - 1; j >= 0; j-- {
			var b = buffer[j]
			if b == '\n' {
				counter += 1
				if counter > num {
					var result = -(stepSize - int64(j) + int64(i)*bufferSize - 1)
					return result, nil
				}
			}
		}

		if isLastRead {
			var result = -fileSize
			return result, nil
		}
	}

	return -fileSize, nil
}
