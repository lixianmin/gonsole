package gonsole

import (
	"fmt"
	"github.com/lixianmin/gonsole/tools"
	"github.com/lixianmin/got/mathx"
	"github.com/lixianmin/logo"
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"
	"unicode"
)

/********************************************************************
created:    2020-07-24
author:     lixianmin

Copyright (C) - All Rights Reserved
*********************************************************************/

func ToHtmlTable(data interface{}) string {
	return tools.ToHtmlTable(data)
}

func ToSnakeName(name string) string {
	var texts = make([]string, 0, 2)
	var startIndex = 0
	for i, c := range name {
		if i > 0 && unicode.IsUpper(c) {
			texts = append(texts, string(unicode.ToLower(rune(name[startIndex])))+name[startIndex+1:i])
			startIndex = i
		}
	}

	if startIndex < len(name) {
		texts = append(texts, string(unicode.ToLower(rune(name[startIndex])))+name[startIndex+1:])
	}

	var result = strings.Join(texts, "_")
	return result
}

// RequestFileByRange https://delveshal.github.io/2018/05/17/golang-%E5%AE%9E%E7%8E%B0%E6%96%87%E4%BB%B6%E6%96%AD%E7%82%B9%E7%BB%AD%E4%BC%A0-demo/
func RequestFileByRange(fullPath string, writer http.ResponseWriter, request *http.Request) {
	var start, end int64
	_, _ = fmt.Sscanf(request.Header.Get("Range"), "bytes=%d-%d", &start, &end)
	file, err := os.Open(fullPath)
	if err != nil {
		logo.JsonD("err", err)
		http.NotFound(writer, request)
		return
	}

	info, err := file.Stat()
	if err != nil {
		logo.JsonD("err", err)
		http.NotFound(writer, request)
		return
	}

	var fileSize = info.Size()
	if start >= fileSize || start > end {
		writer.WriteHeader(http.StatusBadRequest)
		_, _ = writer.Write([]byte(fmt.Sprintf("out of index, length:%d", fileSize)))
		return
	}

	// [-1, -1] 是请求最后一个字节
	if start < 0 {
		start = fileSize + start
		end = fileSize + end
	}

	start = mathx.ClampI64(start, 0, fileSize-1)
	end = mathx.ClampI64(end, start, fileSize-1)

	// 下载整个文件时，不会传入[start, end]，此时需要自己设置为fileSize-1
	if end == 0 {
		end = fileSize - 1
	}

	var header = writer.Header()
	header.Add("Cache-Control", "max-age=864000") // 这个会建议http/2从memory cache或disk cache读取文件
	header.Add("Accept-ranges", "bytes")
	header.Add("Content-Length", strconv.FormatInt(end-start+1, 10))
	header.Add("Content-Range", "bytes "+strconv.FormatInt(start, 10)+"-"+strconv.FormatInt(end, 10)+"/"+strconv.FormatInt(info.Size()-start, 10))
	header.Add("Content-Disposition", "attachment; filename="+info.Name())

	_, err = file.Seek(start, 0)
	if err != nil {
		logo.JsonD("err", err)
		writer.WriteHeader(http.StatusInternalServerError)
		return
	}

	_, err = io.CopyN(writer, file, end-start+1)
	if err != nil {
		logo.JsonD("err", err)
	}
}
