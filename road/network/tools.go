package network

import (
	"context"
	"github.com/lixianmin/gonsole/ifs"
	"github.com/lixianmin/gonsole/road/serde"
	"github.com/lixianmin/logo"
)

/********************************************************************
created:    2020-09-01
author:     lixianmin

Copyright (C) - All Rights Reserved
*********************************************************************/

func GetSessionFromCtx(ctx context.Context) Session {
	fetus := ctx.Value(ifs.CtxKeySession)
	if fetus == nil {
		logo.Warn("ctx doesn't contain the session")
		return nil
	}

	return fetus.(*sessionImpl)
}

// serializeOrRaw serializes the interface if it is not a []byte
func serializeOrRaw(serde serde.Serde, v interface{}) ([]byte, error) {
	if data, ok := v.([]byte); ok {
		return data, nil
	}

	data, err := serde.Serialize(v)
	if err != nil {
		return nil, err
	}

	return data, nil
}
