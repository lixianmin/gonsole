package gonsole

/********************************************************************
created:    2019-11-16
author:     lixianmin

Copyright (C) - All Rights Reserved
*********************************************************************/

const (
	flagBuiltin   = 0x0001 // 内置命令, 并不希望三方代码能看到
	FlagPublic    = 0x0002 // 不需要登录就可以使用的命令
	FlagInvisible = 0x0004 // 在inputBox中无法看到和使用的命令
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

// BuildInfo 封装Git编译信息，通过GetBuildInfo()获取只读副本
type BuildInfo struct {
	BranchName    string
	CommitId      string
	CommitMessage string
	CommitTime    string
	AppBuildTime  string
}

// GetBuildInfo 返回Git编译信息的只读副本
func GetBuildInfo() BuildInfo {
	return BuildInfo{
		BranchName:    GitBranchName,
		CommitId:      GitCommitId,
		CommitMessage: GitCommitMessage,
		CommitTime:    GitCommitTime,
		AppBuildTime:  AppBuildTime,
	}
}

// 内置的两个指令
type builtinCommand struct {
	Name    string
	Example string
	Note    string
}

var builtinSubCommands = []builtinCommand{
	{Name: "sub", Example: "sub top", Note: "订阅主题"},
	{Name: "unsub", Example: "unsub top", Note: "取消订阅主题"},
}

// GetBuiltinSubCommands 返回内置sub/unsub命令的副本
func GetBuiltinSubCommands() []builtinCommand {
	result := make([]builtinCommand, len(builtinSubCommands))
	copy(result, builtinSubCommands)
	return result
}
