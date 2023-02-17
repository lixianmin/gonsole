package beans

import (
	"crypto/sha256"
	"encoding/base64"
	"github.com/golang-jwt/jwt/v4"
	"github.com/lixianmin/gonsole/ifs"
	"github.com/lixianmin/gonsole/jwtx"
	"github.com/lixianmin/gonsole/road"
	"github.com/lixianmin/got/osx"
	"github.com/lixianmin/got/timex"
	"time"
)

/********************************************************************
created:    2020-07-03
author:     lixianmin

Copyright (C) - All Rights Reserved
*********************************************************************/

type CommandAuth struct {
	Code          string `json:"code"`
	Token         string `json:"token,omitempty"`
	GPID          string `json:"gpid,omitempty"`
	ClientAddress string `json:"client,omitempty"`
}

func NewCommandAuth(session road.Session, args []string, userPasswords map[string]string, autoLoginTime time.Duration, port int) *CommandAuth {
	var bean = &CommandAuth{}

	if len(args) < 3 {
		bean.Code = "格式：auth username"
		return bean
	}

	var username = args[1]
	var digest = args[2]
	//logo.JsonI("username", username, "digest", digest)

	// 判断username是否正确
	var password, ok = userPasswords[username]
	if !ok {
		bean.Code = "invalid_username"
		session.Attachment().Put(ifs.KeyIsAuthorized, false)
		return bean
	}

	const jwtSecretKey = "Hey Pet!!"

	// 缓过sha256与base64编码后的digest的长度一定是44, 这是因为sha256返回256 bits的数据, 折合8 bytes, 计算base64编码后的结果长度应该是4 * ceil(n/3)
	var isDigest = len(digest) == 44
	if isDigest {
		// 当是digest的时候, 判断digest是否正确
		var targetDigest = sumSha256(password)
		if targetDigest != digest {
			bean.Code = "invalid_password"
			session.Attachment().Put(ifs.KeyIsAuthorized, false)
			return bean
		}

		// 在client上传digest的时候, 返回一个jwt, 存储在localStorage中, 避免重复输入密码
		// 如果要存储, 推荐使用bcrypt加盐后存储, 这里是要返回jwt, 因此不需要bcrypt
		var data = jwt.MapClaims{}
		data["username"] = username
		data["digest"] = digest

		bean.Token, _ = jwtx.Sign(jwtSecretKey, data, jwtx.WithExpiration(autoLoginTime))
	} else {
		// 如果client转入的是jwt, 则需要解jwt
		var token = digest
		var data, err = jwtx.Parse(jwtSecretKey, token)
		if err != nil || data["username"] != username || data["digest"] != sumSha256(password) {
			bean.Code = "token_expired"
			session.Attachment().Put(ifs.KeyIsAuthorized, false)
			return bean
		}
	}

	bean.Code = "ok"
	bean.GPID = osx.GetGPID(port)
	bean.ClientAddress = session.RemoteAddr().String()
	session.Attachment().Put(ifs.KeyIsAuthorized, true)
	return bean
}

func sumSha256(data string) string {
	// 缓过sha256与base64编码后的digest的长度一定是44, 这是因为sha256返回256 bits的数据, 折合8 bytes, 计算base64编码后的结果长度应该是4 * ceil(n/3)
	var digest = sha256.Sum256([]byte(data))
	var encoded = base64.StdEncoding.EncodeToString(digest[:])
	return encoded
}

func fetchJwt(username string, digest string) (string, error) {
	const secret = "Hey Pet!!"
	var data = jwt.MapClaims{}
	data["username"] = username
	data["digest"] = digest
	var token, err = jwtx.Sign(secret, data, jwtx.WithExpiration(timex.Day))
	return token, err
}
