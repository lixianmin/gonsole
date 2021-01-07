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

// 参考：《编译时向 go 程序写入 git 版本信息》
// http://mengqi.info/html/2015/201502171941-build-go-program-with-git-version.html
var GitBranchName string // git分支名
var GitCommitId string   // git提交id
var AppBuildTime string  // 应用构建时间


// 内置的两个指令
var subUnsubNames = []string{"sub", "unsub"}
var subUnsubNotes = []string{"订阅主题，例：sub top", "取消订阅主题，例：unsub top"}
