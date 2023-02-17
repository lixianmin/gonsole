package jwtx

import (
	"github.com/golang-jwt/jwt/v4"
	"time"
)

/********************************************************************
created:    2022-05-13
author:     lixianmin

Copyright (C) - All Rights Reserved
*********************************************************************/

type arguments struct {
	expiration    time.Duration
	signingMethod jwt.SigningMethod
}

type Option func(*arguments)

func createArguments(options []Option) arguments {
	var args = arguments{
		expiration:    time.Minute,
		signingMethod: jwt.SigningMethodHS256,
	}

	for _, opt := range options {
		opt(&args)
	}

	return args
}

func WithExpiration(expiration time.Duration) Option {
	return func(args *arguments) {
		args.expiration = expiration
	}
}

func WithSigningMethod(method jwt.SigningMethod) Option {
	return func(args *arguments) {
		args.signingMethod = method
	}
}
