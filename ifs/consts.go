package ifs

/********************************************************************
created:    2020-08-27
author:     lixianmin

Copyright (C) - All Rights Reserved
*********************************************************************/

const (
	KeyIsAuthorized = "console.isAuthorized"
	KeyContext      = "console.context"
)

// HeaderAccessToken 下载日志文件时携带令牌的请求头
// 令牌放在请求头而不是url的query里，避免令牌进入浏览器历史与Referer
const HeaderAccessToken = "X-Access-Token"
