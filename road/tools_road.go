package road

import (
	"context"

	"github.com/lixianmin/gonsole/road/serde"
	"github.com/lixianmin/got/convert"
	"github.com/lixianmin/logo"
)

/********************************************************************
created:    2020-09-01
author:     lixianmin

Copyright (C) - All Rights Reserved
*********************************************************************/

// keyType is a type for context keys
type keyType struct{}

var keyNonce = keyType{}
var keySession = keyType{}

func GetSessionFromCtx(ctx context.Context) Session {
	var fetus = ctx.Value(keySession)
	if fetus == nil {
		logo.Warn("ctx doesn't contain the session")
		return nil
	}

	return fetus.(*sessionImpl)
}

func SendDefault(session Session, data any) error {
	if session != nil {
		return session.Send("console.default", data)
	}

	return nil
}

func SendStream(session Session, text string, done bool) error {
	if session != nil {
		var item struct {
			Text string `json:"text,omitempty"`
			Done bool   `json:"done,omitempty"`
		}

		item.Text = text
		item.Done = done

		var json = convert.ToJsonS(item)
		return session.Send("console.stream", json)
	}

	return nil
}

// serializeOrRaw serializes the interface if it is not a []byte
func serializeOrRaw(serde serde.Serde, v any) ([]byte, error) {
	if data, ok := v.([]byte); ok {
		return data, nil
	}

	var data, err = serde.Serialize(v)
	if err != nil {
		return nil, err
	}

	return data, nil
}
