package beans

import (
	"os"
	"testing"
)

/********************************************************************
created:    2026-08-10
author:     xmli

readHeadLines bugfix回归：文件最后一行没有换行符时也必须返回
Copyright (C) - All Rights Reserved
*********************************************************************/

func TestReadHeadLines(t *testing.T) {
	tests := []struct {
		name      string
		content   string
		args      ReadFileArgs
		wantLines int
	}{
		{
			name:      "普通文件-尾行有换行",
			content:   "line1\nline2\nline3\n",
			args:      ReadFileArgs{Num: 10, Filter: "", StartLine: 1},
			wantLines: 3,
		},
		{
			name:      "尾行无换行-全部返回",
			content:   "line1\nline2\nline3",
			args:      ReadFileArgs{Num: 10, Filter: "", StartLine: 1},
			wantLines: 3,
		},
		{
			name:      "尾行无换行-部分返回",
			content:   "line1\nline2\nline3",
			args:      ReadFileArgs{Num: 2, Filter: "", StartLine: 1},
			wantLines: 2,
		},
		{
			name:      "过滤-命中尾行",
			content:   "line1\nline2\nline3",
			args:      ReadFileArgs{Num: 10, Filter: "line3", StartLine: 1},
			wantLines: 1,
		},
		{
			name:      "跳过起始行",
			content:   "line1\nline2\nline3\nline4\n",
			args:      ReadFileArgs{Num: 10, Filter: "", StartLine: 3},
			wantLines: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpfile, err := os.CreateTemp("", "test*.log")
			if err != nil {
				t.Fatal(err)
			}
			defer os.Remove(tmpfile.Name())

			if _, err := tmpfile.WriteString(tt.content); err != nil {
				t.Fatal(err)
			}
			tmpfile.Close()

			tt.args.FullPath = tmpfile.Name()
			var lines = readHeadLines(tt.args)
			if len(lines) != tt.wantLines {
				t.Fatalf("readHeadLines() got %d lines, want %d, lines=%v", len(lines), tt.wantLines, lines)
			}
		})
	}
}
