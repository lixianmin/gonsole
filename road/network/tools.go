package network

import (
	"context"
	"github.com/lixianmin/gonsole/ifs"
	"github.com/lixianmin/gonsole/road/serde"
	"github.com/lixianmin/logo"
	"reflect"
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

	return fetus.(*sessionWrapper)
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

// PCall calls a method that returns an interface and an error and recovers in case of panic
func PCall(method reflect.Method, args []reflect.Value) (rets interface{}, err error) {
	defer func() {
		if rec := recover(); rec != nil {
			logo.Error("method=%d, recover=%v", method.Name, rec)
		}
	}()

	r := method.Func.Call(args)

	// r can have 0 length in case of notify handlers
	// otherwise it will have 2 outputs: an interface and an error
	if len(r) == 2 {
		if v := r[1].Interface(); v != nil {
			err = v.(error)
		} else if !r[0].IsNil() {
			rets = r[0].Interface()
		} else {
			err = ifs.ErrReplyShouldBeNotNull
		}
	}
	return
}
