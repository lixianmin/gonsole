package gonsole

import (
	"github.com/lixianmin/gonsole/beans"
)

/********************************************************************
created:    2019-11-16
author:     lixianmin

Copyright (C) - All Rights Reserved
*********************************************************************/

var beanCreators = map[string]func() IBean{}

func init() {
	registerBeanCreators()
}

func registerBeanCreators() {
	beanCreators["challenge"] = func() IBean { return &beans.Challenge{} }
	beanCreators["command"] = func() IBean { return &beans.CommandRequest{} }
	beanCreators["hint"] = func() IBean { return &beans.HintRequest{} }
	beanCreators["ping"] = func() IBean { return &beans.Ping{} }
	beanCreators["sub"] = func() IBean { return &beans.Subscribe{} }
	beanCreators["unsub"] = func() IBean { return &beans.Unsubscribe{} }
}

func createBean(beanType string) IBean {
	var creator, ok = beanCreators[beanType]
	if ok {
		var bean = creator()
		return bean
	}

	return nil
}
