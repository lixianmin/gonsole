package tools

import (
	"fmt"
	"github.com/lixianmin/got/timex"
	"github.com/lixianmin/road/logger"
	"reflect"
	"strings"
	"time"
	"unicode"
)

/********************************************************************
created:    2020-08-02
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

	sb.WriteString("<table><tr>")
	var numField = writeTableHead(&sb, item)

	sb.WriteString("<tr>")
	for j := 0; j < numField; j++ {
		var fieldType = item.Type().Field(j)
		var fieldValue = item.Field(j)
		writeTableData(&sb, fieldType, fieldValue)
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

	// 表头：第一列用于显示序号
	sb.WriteString("<table><tr><th>")
	var numField = writeTableHead(&sb, listValue.Index(0))
	for i := 0; i < count; i++ {
		var item = listValue.Index(i)
		item = reflect.Indirect(item)

		// 写入序号
		_, _ = fmt.Fprintf(&sb, "<tr><td>%d", i+1)

		for j := 0; j < numField; j++ {
			var fieldType = item.Type().Field(j)
			var fieldValue = item.Field(j)
			writeTableData(&sb, fieldType, fieldValue)
		}
	}

	sb.WriteString("</table>")
	var html = sb.String()
	return html
}

func writeTableHead(sb *strings.Builder, item reflect.Value) int {
	// 每一列的名字
	item = reflect.Indirect(item)
	var itemType = item.Type()
	var numField = itemType.NumField()
	for i := 0; i < numField; i++ {
		var field = itemType.Field(i)
		var name = field.Name
		if unicode.IsLower(rune(name[0])) {
			continue
		}

		sb.WriteString("<th>")
		sb.WriteString(name)
	}

	return numField
}

func writeTableData(sb *strings.Builder, fieldType reflect.StructField, fieldValue reflect.Value) {
	var name = fieldType.Name
	if unicode.IsLower(rune(name[0])) {
		return
	}

	sb.WriteString("<td>")

	var kind = fieldValue.Kind()
	switch kind {
	case reflect.Bool:
		if fieldValue.Bool() {
			sb.WriteString("true")
		} else {
			sb.WriteString("false")
		}
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		_, _ = fmt.Fprintf(sb, "%d", fieldValue.Int())
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		_, _ = fmt.Fprintf(sb, "%d", fieldValue.Uint())
	case reflect.Float32, reflect.Float64:
		_, _ = fmt.Fprintf(sb, "%.3f", fieldValue.Float())
	case reflect.String:
		var v = fieldValue.String()
		sb.WriteString(v)
	case reflect.Struct:
		var t, ok = fieldValue.Interface().(time.Time)
		if ok {
			var v = timex.FormatTime(t)
			sb.WriteString(v)
			break
		}
		// 如果是不认识的struct，就打印错误输出
		fallthrough
	default:
		logger.Error("invalid fieldValue kind=%d, type=%+v", kind, fieldValue.Interface())
	}
}
