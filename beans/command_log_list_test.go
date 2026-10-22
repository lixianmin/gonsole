package beans

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestReadFileSample_PlainText
// Red: 旧实现把"<br/>"拼进样本里，样本是HTML片段而不是纯文本，
//
//	客户端只能通过innerHTML渲染它，日志内容里的尖括号会被当成html解析
//
// Green: 样本是纯文本，换行用"\n"表示，由客户端负责渲染换行
func TestReadFileSample_PlainText(t *testing.T) {
	var dir = t.TempDir()
	var filePath = filepath.Join(dir, "app.log")
	var content = "first line\n<script>alert(1)</script>\nthird line\n"
	if err := os.WriteFile(filePath, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	var sample = readFileSample(filePath, int64(len(content)))

	if strings.Contains(sample, "<br/>") {
		t.Fatalf("sample=%q, want no <br/> in sample", sample)
	}

	if !strings.Contains(sample, "<script>") {
		t.Fatalf("sample=%q, want the log content kept as-is", sample)
	}

	if !strings.Contains(sample, "\n") {
		t.Fatalf("sample=%q, want line breaks kept as \\n", sample)
	}
}
