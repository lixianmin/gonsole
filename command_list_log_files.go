package gonsole

import (
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
	bean.Timestamp = GetTimestamp()

	var logFiles []LogFileInfo
	_ = filepath.Walk(logRoot, func(path string, info os.FileInfo, err error) error {
		if !info.IsDir() {
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
