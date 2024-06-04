package road

import "github.com/lixianmin/gonsole/road/intern"

/********************************************************************
created:    2020-10-05
author:     lixianmin

Copyright (C) - All Rights Reserved
*********************************************************************/

type Acceptor interface {
	GetLinkChan() chan intern.Link
}
