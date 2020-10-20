package beans

import (
	"runtime"
)

/********************************************************************
created:    2020-10-20
author:     lixianmin

Copyright (C) - All Rights Reserved
*********************************************************************/

// 参考：《编译时向 go 程序写入 git 版本信息》
// http://mengqi.info/html/2015/201502171941-build-go-program-with-git-version.html
var GitBranchName string // git分支名
var GitCommitId string   // git提交id
var AppBuildTime string  // 应用构建时间

type CommandAppInfo struct {
	GoVersion     string
	GitBranchName string
	GitCommitId   string
	AppBuildTime  string
}

func NewCommandAppInfo() *CommandAppInfo {
	var info = &CommandAppInfo{}
	info.GoVersion = runtime.Version()
	info.GitBranchName = GitBranchName
	info.GitCommitId = GitCommitId
	info.AppBuildTime = AppBuildTime

	return info
}
