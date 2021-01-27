package tools

import (
	"github.com/lixianmin/got/convert"
	"github.com/lixianmin/logo"
	"reflect"
	"strconv"
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
		logo.Error("data should be slice or struct")
		return ""
	}
}

func toHtmlTableStruct(item reflect.Value) string {
	var b = make([]byte, 0, 512)

	b = append(b, "<table><tr>"...)
	b, numField := writeTableHead(b, item)

	b = append(b, "<tr>"...)
	for j := 0; j < numField; j++ {
		var fieldType = item.Type().Field(j)
		var fieldValue = item.Field(j)
		b = writeTableData(b, fieldType, fieldValue)
	}

	b = append(b, "</table>"...)
	var html = string(b)
	return html
}

func toHtmlTableSlice(listValue reflect.Value) string {
	var count = listValue.Len()
	if count == 0 {
		return ""
	}

	var b = make([]byte, 0, 512)

	// 表头：第一列用于显示序号
	b = append(b, "<table><tr><th>"...)
	b, numField := writeTableHead(b, listValue.Index(0))
	for i := 0; i < count; i++ {
		var item = listValue.Index(i)
		item = reflect.Indirect(item)

		// 写入序号
		b = append(b, "<tr><td>"...)
		b = strconv.AppendInt(b, int64(i+1), 10)

		for j := 0; j < numField; j++ {
			var fieldType = item.Type().Field(j)
			var fieldValue = item.Field(j)
			b = writeTableData(b, fieldType, fieldValue)
		}
	}

	b = append(b, "</table>"...)
	var html = string(b)
	return html
}

func writeTableHead(b []byte, item reflect.Value) ([]byte, int) {
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

		b = append(b, "<th>"...)
		b = append(b, name...)
	}

	return b, numField
}

func writeTableData(b []byte, fieldType reflect.StructField, fieldValue reflect.Value) []byte {
	var name = fieldType.Name
	if unicode.IsLower(rune(name[0])) {
		return b
	}

	b = append(b, "<td>"...)
	b = appendField(b, fieldValue.Interface())
	return b
}

func appendField(b []byte, v interface{}) []byte {
	switch v := v.(type) {
	case nil:
		return append(b, "<nil>"...)
	case string:
		return append(b, v...)
	case []byte:
		return append(b, v...)
	case int:
		return strconv.AppendInt(b, int64(v), 10)
	case int8:
		return strconv.AppendInt(b, int64(v), 10)
	case int16:
		return strconv.AppendInt(b, int64(v), 10)
	case int32:
		return strconv.AppendInt(b, int64(v), 10)
	case int64:
		return strconv.AppendInt(b, v, 10)
	case uint:
		return strconv.AppendUint(b, uint64(v), 10)
	case uint8:
		return strconv.AppendUint(b, uint64(v), 10)
	case uint16:
		return strconv.AppendUint(b, uint64(v), 10)
	case uint32:
		return strconv.AppendUint(b, uint64(v), 10)
	case uint64:
		return strconv.AppendUint(b, v, 10)
	case float32:
		return strconv.AppendFloat(b, float64(v), 'f', -1, 64)
	case float64:
		return strconv.AppendFloat(b, v, 'f', -1, 64)
	case bool:
		return strconv.AppendBool(b, v)
	case error:
		return append(b, v.Error()...)
	case time.Time:
		//b = append(b, '"')
		b = v.AppendFormat(b, time.RFC3339Nano)
		//b = append(b, '"')
		return b
	default:
		return append(b, convert.ToJson(v)...)
	}
}
