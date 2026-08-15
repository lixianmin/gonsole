package beans

import (
	"net"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/lixianmin/gonsole/jwtx"
	"github.com/lixianmin/gonsole/road"
)

/********************************************************************
created:    2026-08-10
author:     xmli

NewCommandAuth jwt分支 bugfix回归：token缺少nonce/digest claim时不能panic
（原实现 data["nonce"].(float64) 对缺失claim直接panic，虽然会被上层recover，
但会导致auth命令静默失败）
Copyright (C) - All Rights Reserved
*********************************************************************/

type fakeRoadSession struct {
	attachment road.Attachment
	nonce      int32
}

func (s *fakeRoadSession) Handshake() error               { return nil }
func (s *fakeRoadSession) Kick(reason string) error       { return nil }
func (s *fakeRoadSession) Send(route string, v any) error { return nil }
func (s *fakeRoadSession) Echo(handler func())            {}
func (s *fakeRoadSession) OnHandShaken(handler func())    {}
func (s *fakeRoadSession) OnClosed(handler func())        {}
func (s *fakeRoadSession) Id() int64                      { return 0 }
func (s *fakeRoadSession) RemoteAddr() net.Addr {
	return &net.TCPAddr{IP: net.IPv4(127, 0, 0, 1), Port: 8888}
}
func (s *fakeRoadSession) Attachment() road.Attachment { return s.attachment }
func (s *fakeRoadSession) Nonce() int32                { return s.nonce }

func TestNewCommandAuth_JwtBranchNoPanic(t *testing.T) {
	const secretKey = "test-secret"
	const username = "xmli"
	const password = "123456"

	tests := []struct {
		name     string
		claims   jwt.MapClaims
		wantCode string
	}{
		{
			name: "缺少nonce claim-不panic",
			claims: jwt.MapClaims{
				"username":    username,
				"digest":      "whatever",
				"fingerprint": "fp",
			},
			wantCode: "invalid_username_or_password",
		},
		{
			name: "缺少digest claim-不panic",
			claims: jwt.MapClaims{
				"username":    username,
				"nonce":       123,
				"fingerprint": "fp",
			},
			wantCode: "invalid_username_or_password",
		},
		{
			name: "username不匹配-不panic",
			claims: jwt.MapClaims{
				"username":    "hacker",
				"digest":      "whatever",
				"nonce":       123,
				"fingerprint": "fp",
			},
			wantCode: "invalid_username_or_password",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var token, err = jwtx.Sign(secretKey, tt.claims, jwtx.WithExpiration(time.Hour))
			if err != nil {
				t.Fatalf("jwtx.Sign() error = %v", err)
			}

			var session = &fakeRoadSession{
				attachment: &road.AttachmentImpl{},
				nonce:      123,
			}

			var bean = NewCommandAuth(session, []string{"auth", username, token, "fp"}, secretKey,
				map[string]string{username: password}, time.Hour, 8888)
			if bean.Code != tt.wantCode {
				t.Fatalf("NewCommandAuth() code = %q, want %q", bean.Code, tt.wantCode)
			}
		})
	}
}

// TestNewCommandAuth_DigestBranch 密码摘要分支：正确digest可以认证成功
func TestNewCommandAuth_DigestBranch(t *testing.T) {
	const secretKey = "test-secret"
	const username = "xmli"
	const password = "123456"
	const nonce int32 = 123

	var session = &fakeRoadSession{
		attachment: &road.AttachmentImpl{},
		nonce:      nonce,
	}

	var digest = sumPasswordDigest(password, nonce)
	var bean = NewCommandAuth(session, []string{"auth", username, digest, "fp"}, secretKey,
		map[string]string{username: password}, time.Hour, 8888)
	if bean.Code != "ok" {
		t.Fatalf("NewCommandAuth() code = %q, want ok", bean.Code)
	}
	if bean.Token == "" {
		t.Fatal("NewCommandAuth() token is empty")
	}
}
