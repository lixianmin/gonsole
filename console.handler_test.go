package gonsole

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
)

/********************************************************************
created:    2026-08-10
author:     xmli

bugfix回归：
1. TestHandleAssets_AbsoluteTemplatePath — PageTemplate为绝对路径时，
   静态资源pattern必须是"/assets/xxx.js"而不是带路径前缀的乱码
   （旧实现 strings.Index(relativePath, \"web/dist\") 对不包含该子串的绝对路径返回-1）
2. TestHandleLogFiles_PathTraversal — /logs/ 路径必须拒绝\"..\"穿越
   （net/http的ServeMux会cleanPath，但自定义mux不会，需显式拦截）
Copyright (C) - All Rights Reserved
*********************************************************************/

type mockServeMux struct {
	patterns []string
	handlers map[string]func(http.ResponseWriter, *http.Request)
}

func (m *mockServeMux) HandleFunc(pattern string, handler func(http.ResponseWriter, *http.Request)) {
	m.patterns = append(m.patterns, pattern)
	m.handlers[pattern] = handler
}

func newMockServeMux() *mockServeMux {
	return &mockServeMux{handlers: map[string]func(http.ResponseWriter, *http.Request){}}
}

// TestHandleAssets_AbsoluteTemplatePath
// Red: 旧实现用 strings.Index(relativePath, \"web\"+sep+\"dist\") 定位dist目录，
//
//	绝对路径模板（不含字面量 web/dist）时 Index=-1，pattern = relativePath[8:] 错乱，
//	\"/assets/app.js\" 不会被注册 → 静态资源404
//
// Green: filepath.Rel 精确计算，注册 \"/assets/app.js\"
func TestHandleAssets_AbsoluteTemplatePath(t *testing.T) {
	var dir = t.TempDir()
	var templatePath = filepath.Join(dir, "console.html")
	if err := os.WriteFile(templatePath, []byte("<html>{{.Data}}</html>"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(dir, "assets"), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "assets", "app.js"), []byte("var x=1;"), 0644); err != nil {
		t.Fatal(err)
	}

	var mux = newMockServeMux()
	_ = NewConsole(mux, WithPageTemplate(templatePath))

	var found = false
	for _, pattern := range mux.patterns {
		if pattern == "/assets/app.js" {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("patterns=%v, want /assets/app.js registered", mux.patterns)
	}
}

// TestHandleLogFiles_PathTraversal
// Red: 旧实现无\"..\"拦截，/logs/../secret.txt 会读到 LogListRoot 之外的任意文件（200+内容）
// Green: 显式拒绝，返回404
func TestHandleLogFiles_PathTraversal(t *testing.T) {
	var dir = t.TempDir()
	var logsRoot = filepath.Join(dir, "logs")
	if err := os.MkdirAll(logsRoot, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(logsRoot, "ok.log"), []byte("log content"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "secret.txt"), []byte("TOP SECRET"), 0644); err != nil {
		t.Fatal(err)
	}

	// RequestFileByRange按相对路径打开文件，切到dir使路径解析可控
	var oldWd, _ = os.Getwd()
	if err := os.Chdir(dir); err != nil {
		t.Fatal(err)
	}
	defer os.Chdir(oldWd)

	var mux = newMockServeMux()
	var templatePath = filepath.Join(dir, "console.html")
	if err := os.WriteFile(templatePath, []byte("<html>{{.Data}}</html>"), 0644); err != nil {
		t.Fatal(err)
	}
	_ = NewConsole(mux, WithPageTemplate(templatePath), WithLogListRoot("logs"))

	var handler, ok = mux.handlers["/logs/"]
	if !ok {
		t.Fatalf("patterns=%v, want /logs/ registered", mux.patterns)
	}

	// 直接调用handler，绕过net/http ServeMux的cleanPath（自定义mux场景）
	var request = httptest.NewRequest("GET", "/logs/../secret.txt", nil)
	var recorder = httptest.NewRecorder()
	handler(recorder, request)

	if recorder.Code != http.StatusNotFound {
		t.Fatalf("status=%d, want 404 (path traversal must be rejected), body=%q", recorder.Code, recorder.Body.String())
	}
}
