package gonsole

import (
	"github.com/lixianmin/gonsole/beans"
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
	beanCreators["challenge"] = func() ifs.Bean { return &beans.Challenge{} }
	beanCreators["command"] = func() ifs.Bean { return &beans.CommandRequest{} }
	beanCreators["hint"] = func() ifs.Bean { return &beans.HintRequest{} }
	beanCreators["ping"] = func() ifs.Bean { return &beans.Ping{} }
	beanCreators["sub"] = func() ifs.Bean { return &beans.Subscribe{} }
	beanCreators["unsub"] = func() ifs.Bean { return &beans.Unsubscribe{} }
}

func createBean(beanType string) ifs.Bean {
	var creator, ok = beanCreators[beanType]
	if ok {
		var bean = creator()
		return bean
	}

	return nil
}
