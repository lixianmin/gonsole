package beans

import (
	"github.com/lixianmin/gonsole/ifs"
)

/********************************************************************
created:    2019-11-16
author:     lixianmin

Copyright (C) - All Rights Reserved
*********************************************************************/

var beanCreators = map[string]func() ifs.Bean{}

func init() {
	registerBeanCreators()
}

func registerBeanCreators() {
	beanCreators["challenge"] = func() ifs.Bean { return &Challenge{} }
	beanCreators["sub"] = func() ifs.Bean { return &Subscribe{} }
	beanCreators["unsub"] = func() ifs.Bean { return &Unsubscribe{} }
}