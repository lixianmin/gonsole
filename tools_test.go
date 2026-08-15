package gonsole

import (
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
)

/********************************************************************
created:    2021-01-07
author:     lixianmin

Copyright (C) - All Rights Reserved
*********************************************************************/

func TestToHtmlTable(t *testing.T) {
	tests := []struct {
		name string
		data any
		want string // 包含特定标识符即可，不比较完整HTML
	}{
		{
			name: "nil",
			data: nil,
			want: "",
		},
		{
			name: "空struct",
			data: struct{}{},
			want: "<table",
		},
		{
			name: "slice",
			data: []struct{ Name string }{{Name: "test"}},
			want: "<table",
		},
		{
			name: "普通struct",
			data: struct{ Name string }{Name: "test"},
			want: "<table",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ToHtmlTable(tt.data)
			// nil返回空字符串，其他返回table标签开头
			if tt.data == nil {
				if got != "" {
					t.Errorf("ToHtmlTable() with nil = %v, want empty string", got)
				}
			} else {
				if len(got) == 0 || got[:6] != "<table" {
					t.Errorf("ToHtmlTable() = %v, want HTML table", got)
				}
			}
		})
	}
}

// TestRequestFileByRange_RangeBehavior 行为级回归：
// Red: 旧实现用fmt.Sscanf解析Range，bytes=0-0（明确请求1个字节）与未指定Range无法区分
//
//	（end==0被当作"整个文件"的哨兵），导致bytes=0-0返回整个文件
//
// Green: parseRangeHeader正确解析，bytes=0-0返回1个字节
func TestRequestFileByRange_RangeBehavior(t *testing.T) {
	tmpfile, err := os.CreateTemp("", "test*.log")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpfile.Name())

	var content = strings.Repeat("a", 100)
	if _, err := tmpfile.WriteString(content); err != nil {
		t.Fatal(err)
	}
	tmpfile.Close()

	tests := []struct {
		name        string
		rangeHeader string
		wantLen     int // 返回的body字节数
		wantStatus  int
	}{
		{"无Range头-整个文件", "", 100, http.StatusOK},
		{"bytes=0-0-只返回1字节", "bytes=0-0", 1, http.StatusOK},
		{"bytes=0-9-返回10字节", "bytes=0-9", 10, http.StatusOK},
		{"bytes=90-到文件尾", "bytes=90-", 10, http.StatusOK},
		{"bytes=-10-最后10字节", "bytes=-10", 10, http.StatusOK},
		{"越界区间-400", "bytes=1000-2000", 0, http.StatusBadRequest},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var request = httptest.NewRequest("GET", "/file.log", nil)
			if tt.rangeHeader != "" {
				request.Header.Set("Range", tt.rangeHeader)
			}

			var recorder = httptest.NewRecorder()
			RequestFileByRange(tmpfile.Name(), recorder, request)

			if recorder.Code != tt.wantStatus {
				t.Fatalf("status=%d, want %d", recorder.Code, tt.wantStatus)
			}
			// 400时body是错误消息本身（"out of index, length:..."），不检查长度
			if tt.wantStatus == http.StatusOK && recorder.Body.Len() != tt.wantLen {
				t.Fatalf("body len=%d, want %d (Range=%q)", recorder.Body.Len(), tt.wantLen, tt.rangeHeader)
			}
		})
	}
}
