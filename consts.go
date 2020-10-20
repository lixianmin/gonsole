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
const GitBranchName string = "" // git分支名
const GitCommitId string = ""   // git提交id
const AppBuildTime string = ""  // 应用构建时间
