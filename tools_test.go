package gonsole

import (
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
