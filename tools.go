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
	file, err := os.Open(fullPath)
	if err != nil {
		logo.JsonD("err", err)
		http.NotFound(writer, request)
		return
	}
	defer file.Close()

	info, err := file.Stat()
	if err != nil {
		logo.JsonD("err", err)
		http.NotFound(writer, request)
		return
	}

	var fileSize = info.Size()
	var start, end int64 = 0, fileSize - 1
	var hasRange = parseRangeHeader(request.Header.Get("Range"), &start, &end)
	if hasRange {
		// bytes=-100 形式：start为负数，表示从文件尾倒数
		if start < 0 {
			start = fileSize + start
			end = fileSize - 1
		} else if end < 0 {
			// bytes=100- 形式：end为负数，表示到文件尾
			end = fileSize - 1
		}

		// 先校验再clamp：越界区间（如bytes=999999-）必须返回400，而不是clamp后静默返回最后一个字节
		if start >= fileSize || start > end {
			writer.WriteHeader(http.StatusBadRequest)
			_, _ = writer.Write([]byte(fmt.Sprintf("out of index, length:%d", fileSize)))
			return
		}

		start = mathx.Clamp(start, 0, fileSize-1)
		end = mathx.Clamp(end, start, fileSize-1)
	}

	var header = writer.Header()
	header.Add("Cache-Control", "max-age=864000") // 这个会建议http/2从memory cache或disk cache读取文件
	header.Add("Accept-ranges", "bytes")
	header.Add("Content-Length", strconv.FormatInt(end-start+1, 10))
	if hasRange {
		header.Add("Content-Range", fmt.Sprintf("bytes %d-%d/%d", start, end, fileSize))
	}
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

// parseRangeHeader 解析Range请求头，支持三种形式：
//   - bytes=0-100   指定区间
//   - bytes=100-    从100到文件尾（end返回-1哨兵）
//   - bytes=-100    最后100字节（start返回负数哨兵，由调用方基于fileSize换算）
//
// bugfix: 原实现用fmt.Sscanf解析，无法区分"bytes=0-0"（明确请求1个字节）与未指定Range
// （end==0被当作"整个文件"的哨兵），导致bytes=0-0返回整个文件；且遇到格式错误的头会
// 部分解析出错误区间。
func parseRangeHeader(header string, start, end *int64) bool {
	if !strings.HasPrefix(header, "bytes=") {
		return false
	}

	var body = header[len("bytes="):]
	var dash = strings.Index(body, "-")
	if dash < 0 {
		return false
	}

	var a = body[:dash]
	var b = body[dash+1:]
	if a == "" {
		// 后缀形式 bytes=-N：最后N个字节
		var n, err = strconv.ParseInt(b, 10, 64)
		if err != nil || n <= 0 {
			return false
		}

		*start = -n
		*end = -1
		return true
	}

	var startValue, err1 = strconv.ParseInt(a, 10, 64)
	if err1 != nil || startValue < 0 {
		return false
	}
	*start = startValue

	if b != "" {
		var endValue, err2 = strconv.ParseInt(b, 10, 64)
		if err2 != nil || endValue < startValue {
			return false
		}
		*end = endValue
	} else {
		*end = -1 // 哨兵：到文件尾
	}

	return true
}
