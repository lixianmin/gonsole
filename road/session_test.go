package road

import (
	"fmt"
	"reflect"
	"testing"
)

/********************************************************************
created:    2024-07-26
author:     lixianmin

Copyright (C) - All Rights Reserved
*********************************************************************/

type Person struct {
	Name string
}

func targetMethod(p *Person, err error) {

}

func TestMakeFunc(t *testing.T) {
	var pt = reflect.TypeOf((*Person)(nil)).Elem()

	var methodType = reflect.FuncOf(
		[]reflect.Type{reflect.PointerTo(pt), reflect.TypeOf((*error)(nil)).Elem()},
		[]reflect.Type{}, false,
	)

	var respond = reflect.MakeFunc(methodType, func(args []reflect.Value) []reflect.Value {
		var args0, args1 = args[0], args[1]
		var response = args0.Interface()

		var err error
		if !args1.IsNil() {
			err = args1.Interface().(error)
		}

		fmt.Printf("Response: %v, Error: %v\n", response, err)
		return []reflect.Value{}
	})

	var person = &Person{Name: "Example"}
	//var err = fmt.Errorf("sample error")

	person = nil
	//err = nil

	respond.Call([]reflect.Value{
		reflect.ValueOf(person),
		reflect.ValueOf((*error)(nil)).Elem(),
	})
}
