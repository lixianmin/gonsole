package road

import (
	"context"
	"github.com/lixianmin/gonsole/ifs"
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
