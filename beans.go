package gonsole

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
	beanCreators["challenge"] = func() IBean { return &Challenge{} }
	beanCreators["command"] = func() IBean { return &CommandRequest{} }
	beanCreators["hint"] = func() IBean { return &HintRequest{} }
	beanCreators["sub"] = func() IBean { return &Subscribe{} }
	beanCreators["unsub"] = func() IBean { return &Unsubscribe{} }
}

func createBean(beanType string) IBean {
	var creator, ok = beanCreators[beanType]
	if ok {
		var bean = creator()
		return bean
	}

	return nil
}
