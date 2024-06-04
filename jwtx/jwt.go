package jwtx

import (
	"errors"
	"github.com/golang-jwt/jwt/v5"
	"github.com/lixianmin/got/convert"
	"time"
)

/********************************************************************
created:    2022-05-13
author:     lixianmin

Copyright (C) - All Rights Reserved
*********************************************************************/

var (
	ErrInvalidToken = errors.New("token is invalid")
)

func Sign(secretKey string, data jwt.MapClaims, options ...Option) (string, error) {
	var args = createArguments(options)
	var expireAt = time.Now().Add(args.expiration)
	data["exp"] = expireAt.Unix()

	var token = jwt.NewWithClaims(args.signingMethod, data)
	var signed, err = token.SignedString(convert.Bytes(secretKey))
	return signed, err
}

func Parse(secretKey string, signedToken string) (jwt.MapClaims, error) {
	var claims = jwt.MapClaims{}
	var token, err = jwt.ParseWithClaims(signedToken, claims, func(token *jwt.Token) (i interface{}, e error) {
		return convert.Bytes(secretKey), nil
	})

	// 如果是过期的话，数据也有可能是可以使用的
	if err != nil {
		return claims, err
	}

	if !token.Valid {
		return nil, ErrInvalidToken
	}

	return claims, nil
}
