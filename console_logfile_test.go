package gonsole

import (
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/lixianmin/gonsole/ifs"
	"github.com/lixianmin/gonsole/jwtx"
)

/********************************************************************
created: 2026-10-22
author:  xmli

TestConsoleLogFileDownload_Integration 集成测试：
下载日志文件的完整链路（真实的net/http mux + 真实的jwt令牌 + 真实的http请求），
令牌放在X-Access-Token请求头中，不再出现在url的query里
Copyright (C) - All Rights Reserved
*********************************************************************/

func TestConsoleLogFileDownload_Integration(t *testing.T) {
	var dir = t.TempDir()
	var logsRoot = filepath.Join(dir, "logs")
	if err := os.MkdirAll(logsRoot, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(logsRoot, "app.log"), []byte("hello log"), 0644); err != nil {
		t.Fatal(err)
	}

	// RequestFileByRange按相对路径打开文件，切到temp目录使相对路径"logs/app.log"可控
	var oldWd, _ = os.Getwd()
	if err := os.Chdir(dir); err != nil {
		t.Fatal(err)
	}
	defer os.Chdir(oldWd)

	var secretKey = "integration-secret-key"
	var templatePath = filepath.Join(dir, "console.html")
	if err := os.WriteFile(templatePath, []byte("<html>{{.Data}}</html>"), 0644); err != nil {
		t.Fatal(err)
	}

	// 真实的http mux，而不是mock
	var mux = http.NewServeMux()
	var console = NewConsole(mux, WithPageTemplate(templatePath), WithLogListRoot("logs"), WithSecretKey(secretKey))
	console.App().Listen()

	var server = httptest.NewServer(mux)
	defer server.Close()

	// 令牌用与log.list命令完全相同的方式生成
	var token, err = jwtx.Sign(secretKey, jwt.MapClaims{"ts": time.Now().UnixMilli()}, jwtx.WithExpiration(time.Minute))
	if err != nil {
		t.Fatal(err)
	}

	// 场景1: 不带令牌，下载被拒绝
	var response1, err1 = http.Get(server.URL + "/logs/app.log")
	if err1 != nil {
		t.Fatal(err1)
	}
	defer response1.Body.Close()
	if response1.StatusCode != http.StatusUnauthorized {
		t.Fatalf("status=%d, want 401 (missing access token)", response1.StatusCode)
	}

	// 场景2: 令牌放在请求头里，下载成功
	var request, err2 = http.NewRequest("GET", server.URL+"/logs/app.log", nil)
	if err2 != nil {
		t.Fatal(err2)
	}
	request.Header.Set(ifs.HeaderAccessToken, token)

	var response2, err3 = http.DefaultClient.Do(request)
	if err3 != nil {
		t.Fatal(err3)
	}
	defer response2.Body.Close()
	if response2.StatusCode != http.StatusOK {
		t.Fatalf("status=%d, want 200, body=%q", response2.StatusCode, readAll(t, response2))
	}

	if body := readAll(t, response2); body != "hello log" {
		t.Fatalf("body=%q, want %q", body, "hello log")
	}
}

func readAll(t *testing.T, response *http.Response) string {
	t.Helper()
	var data, err = io.ReadAll(response.Body)
	if err != nil {
		t.Fatal(err)
	}
	return string(data)
}
