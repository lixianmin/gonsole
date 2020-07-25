package gonsole

import (
	"fmt"
	"github.com/lixianmin/gonsole/logger"
	"github.com/lixianmin/got/mathx"
	"github.com/lixianmin/got/timex"
	"io"
	"net/http"
	"os"
	"reflect"
	"strconv"
	"strings"
	"time"
)

/********************************************************************
created:    2020-07-24
author:     lixianmin

Copyright (C) - All Rights Reserved
*********************************************************************/

func ToHtmlTable(data interface{}) string {
	var dataValue = reflect.Indirect(reflect.ValueOf(data))
	switch dataValue.Kind() {
	case reflect.Slice:
		return toHtmlTableSlice(dataValue)
	case reflect.Struct:
		return toHtmlTableStruct(dataValue)
	default:
		logger.Error("data should be slice")
		return ""
	}
}

func toHtmlTableStruct(item reflect.Value) string {
	var sb strings.Builder
	sb.Grow(256)
	sb.WriteString("<table>")

	var numField = writeTableHead(&sb, item)
	sb.WriteString("<tr>")

	// 写入序号
	sb.WriteString("<td>1")

	for j := 0; j < numField; j++ {
		var field = item.Field(j)
		writeTableData(&sb, field)
	}

	sb.WriteString("</table>")
	var html = sb.String()
	return html
}

func toHtmlTableSlice(listValue reflect.Value) string {
	var count = listValue.Len()
	if count == 0 {
		return ""
	}

	var sb strings.Builder
	sb.Grow(256)
	sb.WriteString("<table>")

	var numField = writeTableHead(&sb, listValue.Index(0))
	for i := 0; i < count; i++ {
		var item = listValue.Index(i)
		item = reflect.Indirect(item)
		sb.WriteString("<tr>")

		// 写入序号
		_, _ = fmt.Fprintf(&sb, "<td>%d", i+1)

		for j := 0; j < numField; j++ {
			var field = item.Field(j)
			writeTableData(&sb, field)
		}
	}

	sb.WriteString("</table>")
	var html = sb.String()
	return html
}

func writeTableHead(sb *strings.Builder, item reflect.Value) int {
	// 第一列用于显示序号
	sb.WriteString("<tr> <th>")

	// 每一列的名字
	item = reflect.Indirect(item)
	var itemType = item.Type()
	var numField = itemType.NumField()
	for i := 0; i < numField; i++ {
		var field = itemType.Field(i)
		sb.WriteString("<th>")
		sb.WriteString(field.Name)
	}

	return numField
}

func writeTableData(sb *strings.Builder, item reflect.Value) {
	sb.WriteString("<td>")

	var kind = item.Kind()
	switch kind {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		_, _ = fmt.Fprintf(sb, "%d", item.Int())
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		_, _ = fmt.Fprintf(sb, "%d", item.Uint())
	case reflect.Float32, reflect.Float64:
		_, _ = fmt.Fprintf(sb, "%.3f", item.Float())
	case reflect.String:
		var v = item.String()
		sb.WriteString(v)
	case reflect.Struct:
		var t, ok = item.Interface().(time.Time)
		if ok {
			var v = timex.FormatTime(t)
			sb.WriteString(v)
			break
		}
		fallthrough
	default:
		logger.Error("invalid item type=%+v", item.Interface())
	}
}

// https://delveshal.github.io/2018/05/17/golang-%E5%AE%9E%E7%8E%B0%E6%96%87%E4%BB%B6%E6%96%AD%E7%82%B9%E7%BB%AD%E4%BC%A0-demo/
func RequestFileByRange(fullPath string, writer http.ResponseWriter, request *http.Request) {
	var start, end int64
	_, _ = fmt.Sscanf(request.Header.Get("Range"), "bytes=%d-%d", &start, &end)
	file, err := os.Open(fullPath)
	if err != nil {
		logger.Debug(err)
		http.NotFound(writer, request)
		return
	}

	info, err := file.Stat()
	if err != nil {
		logger.Debug(err)
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

	start = mathx.ClampInt64(start, 0, fileSize-1)
	end = mathx.ClampInt64(end, start, fileSize-1)

	// 下载整个文件时，不会传入[start, end]，此时需要自己设置为fileSize-1
	if end == 0 {
		end = fileSize - 1
	}

	var header = writer.Header()
	header.Add("Accept-ranges", "bytes")
	header.Add("Content-Length", strconv.FormatInt(end-start+1, 10))
	header.Add("Content-Range", "bytes "+strconv.FormatInt(start, 10)+"-"+strconv.FormatInt(end, 10)+"/"+strconv.FormatInt(info.Size()-start, 10))
	header.Add("Content-Disposition", "attachment; filename="+info.Name())

	_, err = file.Seek(start, 0)
	if err != nil {
		logger.Debug(err)
		writer.WriteHeader(http.StatusInternalServerError)
		return
	}

	_, err = io.CopyN(writer, file, end-start+1)
	if err != nil {
		logger.Debug(err)
	}
}
