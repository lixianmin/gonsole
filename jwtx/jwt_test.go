package jwtx

import (
	"encoding/base64"
	"github.com/golang-jwt/jwt/v4"
	"testing"
	"time"
)

/********************************************************************
created:    2022-05-13
author:     lixianmin

Copyright (C) - All Rights Reserved
*********************************************************************/

func TestSign(t *testing.T) {
	const secretKey = "hello world"
	var data = jwt.MapClaims{
		"id":       123,
		"username": "panda",
		"bytes":    []byte{0, 1, 2, 3, 4, 255},
	}

	var signed, err = Sign(secretKey, data, WithExpiration(time.Second))
	if err != nil {
		t.Fail()
	}

	parsed, err := Parse(secretKey, signed)
	if err != nil {
		t.Fail()
	}

	// []byte数组, 需要再解码一次
	var bytes, _ = base64.StdEncoding.DecodeString(parsed["bytes"].(string))
	parsed["bytes"] = bytes

	var parsedId, _ = parsed["id"].(float64) // parse出来的这一份是 float64
	var rawId, _ = data["id"].(int)          // 原始的这一份是 int
	if int(parsedId) != rawId {
		t.Fail()
	}
}

func TestExpiration(t *testing.T) {
	const secretKey = "hello world"
	var expiration = time.Millisecond * 500
	var data = jwt.MapClaims{
		"id":       123,
		"username": "panda",
	}

	var signed, err = Sign(secretKey, data, WithExpiration(expiration))
	if err != nil {
		t.Fail()
	}

	time.Sleep(expiration + time.Second)
	_, err = Parse(secretKey, signed)
	if err == nil {
		t.Fail()
	}
}
