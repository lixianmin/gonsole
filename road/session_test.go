package road

import (
	"errors"
	"reflect"
	"testing"
)

/********************************************************************
created:    2024-07-26
author:     lixianmin

Copyright (C) - All Rights Reserved
*********************************************************************/

type testPerson struct {
	Name string
}

// TestMakeFunc 测试 reflect.MakeFunc 的表格驱动测试
func TestMakeFunc(t *testing.T) {
	var pt = reflect.TypeOf((*testPerson)(nil)).Elem()

	var methodType = reflect.FuncOf(
		[]reflect.Type{reflect.PointerTo(pt), reflect.TypeOf((*error)(nil)).Elem()},
		[]reflect.Type{}, false,
	)

	tests := []struct {
		name      string
		person    *testPerson
		inputErr  error
		wantPanic bool
	}{
		{
			name:     "valid person with nil error",
			person:   &testPerson{Name: "test"},
			inputErr: nil,
		},
		{
			name:     "valid person with error",
			person:   &testPerson{Name: "test"},
			inputErr: errors.New("test error"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var called bool
			var capturedPerson *testPerson
			var capturedErr error

			var respond = reflect.MakeFunc(methodType, func(args []reflect.Value) []reflect.Value {
				called = true
				// args[0] 是 *testPerson
				if !args[0].IsNil() {
					capturedPerson = args[0].Interface().(*testPerson)
				}
				// args[1] 是 error
				if !args[1].IsNil() {
					capturedErr = args[1].Interface().(error)
				}
				return []reflect.Value{}
			})

			// 构造 error 参数
			var errValue reflect.Value
			if tt.inputErr != nil {
				errValue = reflect.ValueOf(tt.inputErr)
			} else {
				// nil error - 创建一个 nil 的 error interface value
				errValue = reflect.Zero(reflect.TypeOf((*error)(nil)).Elem())
			}

			respond.Call([]reflect.Value{
				reflect.ValueOf(tt.person),
				errValue,
			})

			if !called {
				t.Fatal("MakeFunc handler was not called")
			}

			if capturedPerson == nil || capturedPerson.Name != tt.person.Name {
				t.Errorf("capturedPerson = %v, want %v", capturedPerson, tt.person)
			}

			if tt.inputErr != nil && (capturedErr == nil || capturedErr.Error() != tt.inputErr.Error()) {
				t.Errorf("capturedErr = %v, want %v", capturedErr, tt.inputErr)
			}
		})
	}
}
