package gonsole

import "testing"

// TestParseRangeHeader bugfix回归：Range请求头的解析。
// 原实现用fmt.Sscanf：bytes=0-0（明确请求1个字节）与未指定Range无法区分，
// 且格式错误的头会部分解析出错误区间
func TestParseRangeHeader(t *testing.T) {
	tests := []struct {
		name      string
		header    string
		wantOk    bool
		wantStart int64
		wantEnd   int64
	}{
		{"无Range头", "", false, 0, 0},
		{"非bytes范围", "items=0-10", false, 0, 0},
		{"完整区间", "bytes=0-100", true, 0, 100},
		{"单个字节0-0", "bytes=0-0", true, 0, 0},
		{"前缀到文件尾", "bytes=100-", true, 100, -1},
		{"后缀最后100字节", "bytes=-100", true, -100, -1},
		{"最后一个字节", "bytes=-1", true, -1, -1},
		{"结束小于起始", "bytes=100-50", false, 0, 0},
		{"格式错误", "bytes=abc", false, 0, 0},
		{"后缀为零", "bytes=-0", false, 0, 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var start, end int64
			var ok = parseRangeHeader(tt.header, &start, &end)
			if ok != tt.wantOk {
				t.Fatalf("parseRangeHeader(%q) ok=%v, want %v", tt.header, ok, tt.wantOk)
			}
			if ok && (start != tt.wantStart || end != tt.wantEnd) {
				t.Fatalf("parseRangeHeader(%q) = (%d, %d), want (%d, %d)", tt.header, start, end, tt.wantStart, tt.wantEnd)
			}
		})
	}
}
