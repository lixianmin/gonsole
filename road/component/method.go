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

	// Method needs two or three ins: receiver, context.Context and optional []byte or pointer.
	if methodType.NumIn() != 2 && methodType.NumIn() != 3 {
		return false
	}

	if t1 := methodType.In(1); !t1.Implements(typeOfContext) {
		return false
	}

	if methodType.NumIn() == 3 && methodType.In(2).Kind() != reflect.Ptr && methodType.In(2) != typeOfBytes {
		return false
	}

	// Method needs either no out or two outs: interface{}(or []byte), error
	if methodType.NumOut() != 0 && methodType.NumOut() != 2 {
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
			var isRaw = false
			if methodType.NumIn() == 3 && methodType.In(2) == typeOfBytes {
				isRaw = true
			}
			// rewrite handler name
			if nameFunc != nil {
				methodName = nameFunc(methodName)
			}

			handler := &Handler{
				Method:   method,
				IsRawArg: isRaw,
			}

			if methodType.NumIn() == 3 {
				handler.Type = methodType.In(2)
			}
			methods[methodName] = handler
		}
	}

	return methods
}
