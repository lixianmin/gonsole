package beans

import (
	"bufio"
	"github.com/golang-jwt/jwt/v5"
	"github.com/lixianmin/gonsole/jwtx"
	"github.com/lixianmin/got/timex"
	"os"
	"path/filepath"
	"strings"
	"time"
)

/********************************************************************
created:    2020-06-04
author:     lixianmin

Copyright (C) - All Rights Reserved
*********************************************************************/

type LogFileInfo struct {
	Size        int64  `json:"size"`
	Path        string `json:"path"`
	ModTime     string `json:"mod_time"`
	Sample      string `json:"sample"`
	AccessToken string `json:"access_token"`
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

	// todo secret key需要是项目启动时传入的, 不能是固定的, 否则hacker可以自己定制jwt token
	const jwtSecretKey = "Hey Pet!!"
	var data = jwt.MapClaims{}
	data["ts"] = time.Now().UnixMilli()
	var token, _ = jwtx.Sign(jwtSecretKey, data, jwtx.WithExpiration(time.Minute))

	_ = filepath.Walk(logRoot, func(path string, info os.FileInfo, err error) error {
		if info != nil && !info.IsDir() {
			logFiles = append(logFiles, LogFileInfo{
				Size:        info.Size(),
				Path:        path,
				ModTime:     info.ModTime().Format(timex.Layout),
				Sample:      readFileSample(path, info.Size()),
				AccessToken: token,
			})
		}

		return nil
	})

	// 我怀疑默认的顺序就是文件名的字母序，所以可能没有必要重新按照Path重新排序
	//// 使用路径名排序
	//sort.Slice(logFiles, func(i, j int) bool {
	//	return logFiles[i].Path < logFiles[j].Path
	//})

	bean.LogFiles = logFiles
	return bean
}
