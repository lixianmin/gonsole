package tools

import (
	"bufio"
	"fmt"
	"io"
	"os"
)

/********************************************************************
created:    2020-07-23
author:     lixianmin

Copyright (C) - All Rights Reserved
*********************************************************************/

func ReadFileTail(fullPath string, num int) []string {
	if num <= 0 {
		return nil
	}

	var fin, err = os.Open(fullPath)
	if err != nil {
		return nil
	}

	defer fin.Close()

	offset, err := searchOffset(fin, num)
	if err != nil {
		return nil
	}

	_, err = fin.Seek(offset, io.SeekEnd)
	if err != nil {
		return nil
	}

	var reader = bufio.NewReader(fin)
	var lines = make([]string, 0, num)
	for {
		var line, err = reader.ReadString('\n')
		if err != nil {
			return lines
		}

		lines = append(lines, line)
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