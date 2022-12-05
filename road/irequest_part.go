package road

import "context"

/********************************************************************
created:    2022-12-05
author:     lixianmin

Copyright (C) - All Rights Reserved
*********************************************************************/

type IRequestPart interface {
	OnAdded(ctx context.Context, request interface{}) error
}
