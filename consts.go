package gonsole

/********************************************************************
created:    2019-11-16
author:     lixianmin

Copyright (C) - All Rights Reserved
*********************************************************************/

const (
//OK            = 0   // 正确返回
//InternalError = 500 // 内部错误
)

// GitBranchName 参考：《编译时向 go 程序写入 git 版本信息》
// http://mengqi.info/html/2015/201502171941-build-go-program-with-git-version.html
/*
IMPORT_PATH=github.com/lixianmin/gonsole
FLAGS="-w -s -X $IMPORT_PATH.GitBranchName=`git rev-parse --abbrev-ref HEAD` -X $IMPORT_PATH.GitCommitId=`git log --pretty=format:\"%h\" -1` -X '$IMPORT_PATH.GitCommitMessage=`git show -s --format=%s`' -X $IMPORT_PATH.GitCommitTime=`git log --date=format:'%Y-%m-%dT%H:%M:%S' --pretty=format:%ad -1` -X $IMPORT_PATH.AppBuildTime=`date +%Y-%m-%dT%H:%M:%S`"
go build -ldflags "$FLAGS" -mod vendor -gcflags "-N -l"
*/
var GitBranchName string    // git分支名: git rev-parse --abbrev-ref HEAD
var GitCommitId string      // git提交id: git log --pretty=format:\"%h\" -1
var GitCommitMessage string // git提交的message: git show -s --format=%s
var GitCommitTime string    // git提交的时间: git log --date=format:'%Y-%m-%dT%H:%M:%S' --pretty=format:%ad -1
var AppBuildTime string     // 应用构建时间: date +%Y-%m-%dT%H:%M:%S

// 内置的两个指令
var subUnsubNames = []string{"sub", "unsub"}
var subUnsubNotes = []string{"订阅主题，例：sub top", "取消订阅主题，例：unsub top"}
