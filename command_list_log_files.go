package gonsole

import (
	"github.com/lixianmin/gonsole/tools"
	"github.com/lixianmin/got/timex"
	"io/ioutil"
	"os"
	"path/filepath"
)

/********************************************************************
created:    2020-06-04
author:     lixianmin

Copyright (C) - All Rights Reserved
*********************************************************************/

type LogFileInfo struct {
	Size    int64  `json:"size"`
	Path    string `json:"path"`
	ModTime string `json:"mod_time"`
	Sample  string `json:"sample"`
}

type CommandListLogFiles struct {
	BasicResponse
	LogFiles []LogFileInfo `json:"logFiles"`
}

func readFileSample(filePath string, fileSize int64) string {
	const halfSize = 128
	if fileSize < 2*halfSize {
		var data, err = ioutil.ReadFile(filePath)
		if err != nil {
			return ""
		}

		return string(data)
	} else {
		fin, err := os.Open(filePath)
		if err != nil {
			return ""
		}

		defer fin.Close()

		var data [halfSize]byte
		_, err = fin.Read(data[0:])
		var head = string(data[0:])
		_, err = fin.ReadAt(data[0:], fileSize-halfSize)
		var tail = string(data[0:])

		var sample = head + "\n......\n" + tail
		return sample
	}
}

func newCommandListLogFiles(logRoot string) *CommandListLogFiles {
	var bean = &CommandListLogFiles{}
	bean.Operation = "listLogFiles"
	bean.Timestamp = tools.GetTimestamp()

	var logFiles = make([]LogFileInfo, 0, 4)
	_ = filepath.Walk(logRoot, func(path string, info os.FileInfo, err error) error {
		if info != nil && !info.IsDir() {
			logFiles = append(logFiles, LogFileInfo{
				Size:    info.Size(),
				Path:    path,
				ModTime: timex.FormatTime(info.ModTime()),
				Sample:  readFileSample(path, info.Size()),
			})
		}

		return nil
	})

	bean.LogFiles = logFiles
	return bean
}
