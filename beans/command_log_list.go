package beans

import (
	"bufio"
	"github.com/lixianmin/got/timex"
	"os"
	"path/filepath"
	"sort"
	"strings"
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

type CommandLogList struct {
	LogFiles []LogFileInfo `json:"logFiles"`
}

func readFileSample(filePath string, fileSize int64) string {
	fin, err := os.Open(filePath)
	if err != nil {
		return ""
	}

	defer fin.Close()

	var reader = bufio.NewReader(fin)
	var sb strings.Builder
	sb.Grow(512)

	const sampleLines = 5
	for i := 0; i < sampleLines; i++ {
		var line, err = reader.ReadString('\n')
		if err != nil {
			break
		}
		sb.WriteString(line)
		sb.WriteString("<br/>")
	}

	var sample = sb.String()
	return sample
}

func NewCommandLogList(logRoot string) *CommandLogList {
	var bean = &CommandLogList{}
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

	sort.Slice(logFiles, func(i, j int) bool {
		return logFiles[i].ModTime < logFiles[j].ModTime
	})

	bean.LogFiles = logFiles
	return bean
}
