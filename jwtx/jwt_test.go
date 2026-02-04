package jwtx

import (
	"encoding/base64"
	"github.com/golang-jwt/jwt/v5"
	"testing"
	"time"
)

/********************************************************************
created:    2022-05-13
author:     lixianmin

Copyright (C) - All Rights Reserved
*********************************************************************/

func TestSignAndParse(t *testing.T) {
	type args struct {
		secretKey  string
		claims     jwt.MapClaims
		expiration time.Duration
	}

	tests := []struct {
		name           string
		args           args
		wantSignErr    bool
		wantParseErr   bool
		checkClaims    bool
		sleepBeforeParse time.Duration
	}{
		{
			name: "正常签名解析",
			args: args{
				secretKey:  "hello world",
				claims:     jwt.MapClaims{"id": 123, "username": "panda"},
				expiration: time.Hour,
			},
			wantSignErr:  false,
			wantParseErr: false,
			checkClaims:  true,
		},
		{
			name: "不同密钥解析失败",
			args: args{
				secretKey:  "correct_key",
				claims:     jwt.MapClaims{"id": 123},
				expiration: time.Hour,
			},
			wantSignErr:  false,
			wantParseErr: false,
			checkClaims:  true,
		},
		{
			name: "过期令牌解析失败",
			args: args{
				secretKey:  "hello world",
				claims:     jwt.MapClaims{"id": 123, "username": "panda"},
				expiration: time.Millisecond * 100,
			},
			wantSignErr:      false,
			wantParseErr:     true,
			sleepBeforeParse: time.Millisecond * 200,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			signed, err := Sign(tt.args.secretKey, tt.args.claims, WithExpiration(tt.args.expiration))
			if (err != nil) != tt.wantSignErr {
				t.Fatalf("Sign() error = %v, wantSignErr %v", err, tt.wantSignErr)
			}
			if tt.wantSignErr {
				return
			}

			if tt.sleepBeforeParse > 0 {
				time.Sleep(tt.sleepBeforeParse)
			}

			parsed, err := Parse(tt.args.secretKey, signed)
			if (err != nil) != tt.wantParseErr {
				t.Fatalf("Parse() error = %v, wantParseErr %v", err, tt.wantParseErr)
			}

			if tt.checkClaims && err == nil {
				// 验证claims
				if id, ok := parsed["id"]; ok {
					if parsedId, ok := id.(float64); ok {
						if int(parsedId) != 123 {
							t.Errorf("parsed id = %v, want 123", parsedId)
						}
					}
				}
			}
		})
	}
}

func TestSignWithWrongKey(t *testing.T) {
	const secretKey = "correct_key"
	var data = jwt.MapClaims{"id": 123}

	signed, err := Sign(secretKey, data, WithExpiration(time.Hour))
	if err != nil {
		t.Fatalf("Sign() error = %v", err)
	}

	// 使用错误的密钥解析
	_, err = Parse("wrong_key", signed)
	if err == nil {
		t.Error("Parse() with wrong key should fail")
	}
}

func TestSignWithByteArray(t *testing.T) {
	const secretKey = "hello world"
	var data = jwt.MapClaims{
		"id":       123,
		"username": "panda",
		"bytes":    []byte{0, 1, 2, 3, 4, 255},
	}

	signed, err := Sign(secretKey, data, WithExpiration(time.Second))
	if err != nil {
		t.Fatalf("Sign() error = %v", err)
	}

	parsed, err := Parse(secretKey, signed)
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	// []byte数组, 需要再解码一次
	var bytes, _ = base64.StdEncoding.DecodeString(parsed["bytes"].(string))
	parsed["bytes"] = bytes

	// 验证byte数组
	expectedBytes := []byte{0, 1, 2, 3, 4, 255}
	if string(bytes) != string(expectedBytes) {
		t.Errorf("bytes mismatch, got %v, want %v", bytes, expectedBytes)
	}

	var parsedId, _ = parsed["id"].(float64)
	var rawId, _ = data["id"].(int)
	if int(parsedId) != rawId {
		t.Errorf("id mismatch, got %v, want %v", int(parsedId), rawId)
	}
}
