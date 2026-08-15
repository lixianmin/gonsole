package tools

import (
	"os"
	"sort"
	"testing"
)

/********************************************************************
created:    2020-07-23
author:     lixianmin

Copyright (C) - All Rights Reserved
*********************************************************************/

func TestReadTailLines(t *testing.T) {
	// 创建临时文件
	tmpfile, err := os.CreateTemp("", "test*.log")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpfile.Name())

	// 写入测试数据
	content := "line1\nline2\nline3\nline4\nline5\n"
	if _, err := tmpfile.WriteString(content); err != nil {
		t.Fatal(err)
	}
	tmpfile.Close()

	tests := []struct {
		name    string
		num     int
		filter  string
		wantLen int // 期望行数
		wantErr bool
	}{
		{"读取全部", 10, "", 5, false},
		{"读取部分", 3, "", 3, false},
		{"带过滤", 10, "line3", 1, false},
		{"零行数", 0, "", 0, true},
		{"不存在的文件", 10, "", 0, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			path := tmpfile.Name()
			if tt.name == "不存在的文件" {
				path = "/nonexistent/path/file.log"
			}

			got, err := ReadTailLines(path, tt.num, tt.filter)
			if (err != nil) != tt.wantErr {
				t.Fatalf("ReadTailLines() error = %v, wantErr %v", err, tt.wantErr)
			}
			if err != nil {
				return
			}
			if len(got) != tt.wantLen {
				t.Errorf("ReadTailLines() got %d lines, want %d", len(got), tt.wantLen)
			}
		})
	}
}

// TestReadTailLines_NoTrailingNewline bugfix回归：文件最后一行没有换行符时也必须返回。
// 原实现遇到io.EOF直接返回，丢掉了最后一行（ReadString在EOF时仍会返回部分数据）
func TestReadTailLines_NoTrailingNewline(t *testing.T) {
	tmpfile, err := os.CreateTemp("", "test*.log")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpfile.Name())

	// 注意：最后一行 "line5" 没有 \n
	content := "line1\nline2\nline3\nline4\nline5"
	if _, err := tmpfile.WriteString(content); err != nil {
		t.Fatal(err)
	}
	tmpfile.Close()

	tests := []struct {
		name    string
		num     int
		filter  string
		wantLen int
	}{
		{"读取全部-无尾换行", 10, "", 5},
		{"读取部分-无尾换行", 3, "", 3},
		{"带过滤-无尾换行", 10, "line5", 1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ReadTailLines(tmpfile.Name(), tt.num, tt.filter)
			if err != nil {
				t.Fatalf("ReadTailLines() error = %v", err)
			}
			if len(got) != tt.wantLen {
				t.Errorf("ReadTailLines() got %d lines, want %d, lines=%v", len(got), tt.wantLen, got)
			}
		})
	}
}

func TestTimeSort(t *testing.T) {
	tests := []struct {
		name string
		data []string
		asc  bool // true=升序, false=降序
		want []string
	}{
		{
			name: "升序排序",
			data: []string{"2022-10-19\t", "2021-10-18\t", "2021-10-19\t", "2021-10-20\t"},
			asc:  true,
			want: []string{"2021-10-18\t", "2021-10-19\t", "2021-10-20\t", "2022-10-19\t"},
		},
		{
			name: "降序排序",
			data: []string{"2022-10-19\t", "2021-10-18\t", "2021-10-19\t", "2021-10-20\t"},
			asc:  false,
			want: []string{"2022-10-19\t", "2021-10-20\t", "2021-10-19\t", "2021-10-18\t"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data := make([]string, len(tt.data))
			copy(data, tt.data)

			if tt.asc {
				sort.Slice(data, func(i, j int) bool {
					return data[i] < data[j]
				})
			} else {
				sort.Slice(data, func(i, j int) bool {
					return data[i] > data[j]
				})
			}

			for i := range data {
				if data[i] != tt.want[i] {
					t.Errorf("sort result[%d] = %v, want %v", i, data[i], tt.want[i])
				}
			}
		})
	}
}
