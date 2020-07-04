package tools

import (
	"crypto/hmac"
	"crypto/sha256"
	"fmt"
	"io"
)

/********************************************************************
created:    2020-07-04
author:     lixianmin

Copyright (C) - All Rights Reserved
*********************************************************************/

func HmacSha256(key string, message string) string {
	h := hmac.New(sha256.New, []byte(key))
	_, _ = io.WriteString(h, message)
	return fmt.Sprintf("%x", h.Sum(nil))
}
