package component

import (
	"context"
	"reflect"
	"unicode"
	"unicode/utf8"
)

/********************************************************************
created:    2020-08-29
author:     lixianmin

Copyright (C) - All Rights Reserved
*********************************************************************/

var (
	typeOfError   = reflect.TypeOf((*error)(nil)).Elem()
	typeOfBytes   = reflect.TypeOf(([]byte)(nil))
	typeOfContext = reflect.TypeOf(new(context.Context)).Elem()
)

func isExported(name string) bool {
	w, _ := utf8.DecodeRuneInString(name)
	return unicode.IsUpper(w)
}

// isHandlerMethod decide a method is suitable handler method
func isHandlerMethod(method reflect.Method) bool {
	var methodType = method.Type
	// Method must be exported.
	if method.PkgPath != "" {
		return false
	}

	// 输入参数列表3个或4个, 如果是3个参数则需要有返回值, 如果是4个参数则第4个参数是respond回调方法
	var numIn = methodType.NumIn()
	if numIn != 3 && numIn != 4 {
		return false
	}

	// t1必须是context.Context
	if t1 := methodType.In(1); !t1.Implements(typeOfContext) {
		return false
	}

	// t2必须是[]byte或pointer
	if t2 := methodType.In(2); t2.Kind() != reflect.Ptr && t2 != typeOfBytes {
		return false
	}

	if numIn == 4 {
		// 如果有t3, 则必须是一个回调方法
		if t3 := methodType.In(3); t3.Kind() != reflect.Func {
			return false
		}
	} else if methodType.NumOut() != 2 {
		// 否则必须有两个返回值
		return false
	}

	if methodType.NumOut() == 2 && (methodType.Out(1) != typeOfError || methodType.Out(0) != typeOfBytes && methodType.Out(0).Kind() != reflect.Ptr) {
		return false
	}

	return true
}

func suitableHandlerMethods(type1 reflect.Type, nameFunc func(string) string) map[string]*Handler {
	var methods = make(map[string]*Handler)
	var numMethod = type1.NumMethod()

	for index := 0; index < numMethod; index++ {
		var method = type1.Method(index)
		var methodType = method.Type
		var methodName = method.Name
		if isHandlerMethod(method) {
			// 重写methodName
			if nameFunc != nil {
				methodName = nameFunc(methodName)
			}

			var handler = &Handler{
				Method:      method,
				RequestType: methodType.In(2),
			}

			if methodType.NumIn() == 4 {
				handler.RespondMethodType = reflect.PointerTo(methodType.In(3))
			}

			methods[methodName] = handler
		}
	}

	return methods
}
