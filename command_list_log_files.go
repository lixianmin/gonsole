package gonsole

import (
	"github.com/lixianmin/gonsole/tools"
	"os"
	"path/filepath"
)

/********************************************************************
created:    2020-06-04
author:     lixianmin

Copyright (C) - All Rights Reserved
*********************************************************************/

type LogFileInfo struct {
	Size int64  `json:"size"`
	Path string `json:"path"`
}

type CommandListLogFiles struct {
	BasicResponse
	LogFiles []LogFileInfo `json:"logFiles"`
}

func newCommandListLogFiles(logRoot string) *CommandListLogFiles {
	var bean = &CommandListLogFiles{}
	bean.Operation = "listLogFiles"
	bean.Timestamp = tools.GetTimestamp()

	var logFiles = make([]LogFileInfo, 0, 4)
	_ = filepath.Walk(logRoot, func(path string, info os.FileInfo, err error) error {
		if info != nil && !info.IsDir() {
			logFiles = append(logFiles, LogFileInfo{
				Size: info.Size(),
				Path: path,
			})
		}

		return nil
	})

	bean.LogFiles = logFiles
	return bean
}
